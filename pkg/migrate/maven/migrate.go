package maven

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/config"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/log/logfields"
	"e.coding.net/codingcorp/carctl/pkg/migrate/maven/types"
	"e.coding.net/codingcorp/carctl/pkg/migrate/maven/types/nexus"
	reportutil "e.coding.net/codingcorp/carctl/pkg/report"
	"e.coding.net/codingcorp/carctl/pkg/settings"
	"e.coding.net/codingcorp/carctl/pkg/util/fileutil"
	"e.coding.net/codingcorp/carctl/pkg/util/httputil"
	"e.coding.net/codingcorp/carctl/pkg/util/ioutils"
)

var (
	ErrFileConflict = errors.New("failed to put file: 409 conflict")
	MetadataXml     = "maven-metadata.xml"
	Snapshot        = "SNAPSHOT"
	Metadata        = "Metadata"
)

func Migrate(cfg *action.Configuration, out io.Writer) error {
	if settings.Src == "" {
		settings.Src = defaultMavenRepositoryPath()
	}

	isLocalPath := isLocalRepository(settings.Src)
	if isLocalPath {
		// local repository
		return MigrateFromDisk(cfg, out)
	} else {
		// remote repository
		srcUrl, err := url.Parse(settings.Src)
		if err != nil {
			log.Warn("Invalid src url", logfields.String("src", settings.Src), logfields.Error(err))
			return errors.Wrap(err, "invalid src url")
		}
		if srcUrl != nil && srcUrl.Scheme == "" {
			srcUrl.Scheme = "http"
		}
		return MigrateFromUrl(cfg, out, srcUrl)
	}
}

func MigrateFromDisk(cfg *action.Configuration, out io.Writer) error {
	log.Info("Stat source repository ...")

	repositoryFileInfo, err := os.Stat(settings.Src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("source repository not found", logfields.String("path", settings.Src))
			return nil
		}
		return err
	}
	if !repositoryFileInfo.IsDir() {
		return errors.New("source repository is not a directory")
	}

	log.Info("Check authorization of the registry")
	configFile, err := cfg.RegistryClient.ConfigFile()
	if err != nil {
		return errors.Wrap(err, "failed to get config file")
	}

	has, authConfig, err := configFile.GetAuthConfig(settings.Dst)
	if err != nil {
		return errors.Wrap(err, "failed to get registry authorization info")
	}
	if !has {
		return errors.New("Unauthorized: authentication required. Maybe you haven't logged in before.")
	}

	if settings.Verbose {
		log.Debug("Auth config", logfields.String("host", authConfig.ServerAddress),
			logfields.String("username", authConfig.Username),
			logfields.String("password", authConfig.Password))
	}

	if err = migrateRepository(out, authConfig.Username, authConfig.Password); err != nil {
		return err
	}

	return nil
}

func MigrateFromUrl(cfg *action.Configuration, out io.Writer, srcUrl *url.URL) error {
	// ????????? nexus
	if settings.SrcType == "" {
		settings.SrcType = "nexus"
	}
	switch settings.SrcType {
	case "nexus":
		return MigrateFromNexus(cfg, out, srcUrl)
	default:
		return errors.Errorf("This src-type [%s] is not supported", settings.SrcType)
	}
}

func MigrateFromNexus(cfg *action.Configuration, out io.Writer, nexusUrl *url.URL) error {
	log.Infof("Get file list from source repository [%s] ...", settings.Src)

	nexusScheme := nexusUrl.Scheme
	nexusHost := nexusUrl.Host
	urlPath := nexusUrl.Path
	urlPathStrs := strings.Split(strings.Trim(urlPath, "/"), "/")
	repoName := urlPathStrs[1]
	continuationToken := ""

	var nexusItemList []nexus.Item
	for true {
		resp, err := getFileListFromNexus(nexusScheme, nexusHost, repoName, continuationToken)
		if err != nil {
			log.Errorf("failed to get file list, err: %s, continuationToken: %s", err, continuationToken)
			break
		}
		nexusItemList = append(nexusItemList, resp.Items...)
		if strings.TrimSpace(resp.ContinuationToken) == "" {
			break
		}
		continuationToken = resp.ContinuationToken
		// TODO: ??????????????????????????????????????????????????? continuationToken ????????????????????????????????????
	}

	log.Info("Check authorization of the registry")
	configFile, err := cfg.RegistryClient.ConfigFile()
	if err != nil {
		return errors.Wrap(err, "failed to get config file")
	}

	has, authConfig, err := configFile.GetAuthConfig(settings.Dst)
	if err != nil {
		return errors.Wrap(err, "failed to get registry authorization info")
	}
	if !has {
		return errors.New("Unauthorized: authentication required. Maybe you haven't logged in before.")
	}

	if settings.Verbose {
		log.Debug("Auth config", logfields.String("host", authConfig.ServerAddress),
			logfields.String("username", authConfig.Username),
			logfields.String("password", authConfig.Password))
	}

	if err = migrateNexusRepository(out, nexusItemList, authConfig.Username, authConfig.Password); err != nil {
		return err
	}

	return nil
}

// getFileListFromNexus ?????? nexus3 API ?????????????????????
func getFileListFromNexus(scheme, nexusHost, repository, continuationToken string) (*nexus.GetAssetsResponse, error) {
	apiUrl := fmt.Sprintf("%s://%s/service/rest/v1/assets?repository=%s", scheme, nexusHost, repository)
	if continuationToken != "" {
		apiUrl = fmt.Sprintf("%s&continuationToken=%s", apiUrl, continuationToken)
	}

	resp, err := httputil.DefaultClient.GetWithAuth(apiUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get components: %s", apiUrl)
	}
	defer ioutils.QuiteClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		// TODO: ????????????????????????????????????????????????
		// ?????????????????? 404?????????????????????????????? nexus3.x???API ????????? /nexus ?????????
		if resp.StatusCode == http.StatusNotFound {
			apiUrl = fmt.Sprintf("%s://%s/nexus/service/rest/v1/assets?repository=%s", scheme, nexusHost, repository)
			if continuationToken != "" {
				apiUrl = fmt.Sprintf("%s&continuationToken=%s", apiUrl, continuationToken)
			}
			resp, err = httputil.DefaultClient.GetWithAuth(apiUrl, settings.SrcUsername, settings.SrcPassword)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get components: %s", apiUrl)
			}
			defer ioutils.QuiteClose(resp.Body)

			if resp.StatusCode != http.StatusOK {
				return nil, errors.Errorf("failed to get components: %s, status: %s", apiUrl, resp.Status)
			}
		} else {
			return nil, errors.Errorf("failed to get components: %s, status: %s", apiUrl, resp.Status)
		}
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read resp: %s", apiUrl)
	}

	var getComponentsResp *nexus.GetAssetsResponse
	err = json.Unmarshal(bodyBytes, &getComponentsResp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal resp: %s", string(bodyBytes))
	}
	return getComponentsResp, nil
}

func migrateRepository(w io.Writer, username, password string) error {
	log.Info("Scanning repository ...")

	repository, err := GetRepository(settings.Src, settings.MaxFiles)
	if err != nil {
		return err
	}
	flattenRepository := repository.Flatten()
	log.Info("Successfully to scan the repository",
		logfields.Int("groups", flattenRepository.GetGroupCount()),
		logfields.Int("artifacts", flattenRepository.GetArtifactCount()),
		logfields.Int("versions", flattenRepository.GetVersionCount()),
		logfields.Int("files", flattenRepository.GetFileCount()))
	if flattenRepository.GetFileCount() == 0 {
		log.Warn("no files found, no need to migrate")
		return nil
	}
	if settings.Verbose {
		log.Info("Repository Info:")
		repository.Render(w)
	}

	// Progress Bar
	// initialize progress container, with custom width
	p := mpb.New(mpb.WithWidth(80))
	total := flattenRepository.GetFileCount()
	const pbName = "Pushing:"
	// adding a single bar, which will inherit container's width
	bar := p.Add(
		int64(total),
		mpb.NewBarFiller(mpb.BarStyle()),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(pbName, decor.WC{W: len(pbName) + 1, C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), "Done!",
			),
		),
		mpb.AppendDecorators(
			// counter
			decor.Counters(0, "%d / %d  "),
			// percentage
			decor.Percentage(),
			// average
			// mpb.AppendDecorators(decor.AverageSpeed(decor.UnitKiB, "  % .1f")),
		),
	)

	log.Info("Begin to migrate ...")
	start := time.Now()

	report := reportutil.NewReport()
	if settings.Verbose {
		defer func() {
			log.Info("Migrate result:")
			report.Render(w)
		}()
	}

	if err := repository.ForEach(func(group, artifact, version, path, downloadUrl string) error {
		defer bar.Increment()
		if err1 := doMigrate(path, username, password); err1 != nil {
			if err1 == ErrFileConflict {
				report.AddSkippedResult(strings.Join([]string{group, artifact, version}, ":"), path, "409 Conflict")
				return types.ErrForEachContinue
			}

			report.AddFailedResult(strings.Join([]string{group, artifact, version}, ":"), path, err1.Error())

			if settings.FailFast {
				return errors.Wrapf(err1, "failed to migrate %s", path)
			}
		} else {
			report.AddSucceededResult(strings.Join([]string{group, artifact, version}, ":"), path, "Succeeded")
		}

		return nil
	}); err != nil {
		return err
	}

	// wait for our bar to complete and flush
	p.Wait()

	log.Info("End to migrate.",
		logfields.Duration("duration", time.Now().Sub(start)),
		logfields.Int("succeededCount", len(report.SucceededResult)),
		logfields.Int("skippedCount", len(report.SkippedResult)),
		logfields.Int("failedCount", len(report.FailedResult)))

	return nil
}

func migrateNexusRepository(w io.Writer, nexusItemList []nexus.Item, username, password string) error {
	log.Info("Scanning nexus repository ...")

	repository, err := GetRepositoryFromNexusItems(settings.Src, nexusItemList)
	if err != nil {
		return err
	}
	flattenRepository := repository.Flatten()
	log.Info("Successfully to scan the repository",
		logfields.Int("groups", flattenRepository.GetGroupCount()),
		logfields.Int("artifacts", flattenRepository.GetArtifactCount()),
		logfields.Int("versions", flattenRepository.GetVersionCount()),
		logfields.Int("files", flattenRepository.GetFileCount()))
	if flattenRepository.GetFileCount() == 0 {
		log.Warn("no files found, no need to migrate")
		return nil
	}
	if settings.Verbose {
		log.Info("Repository Info:")
		repository.Render(w)
	}

	// Progress Bar
	// initialize progress container, with custom width
	p := mpb.New(mpb.WithWidth(80))
	total := flattenRepository.GetFileCount()
	const pbName = "Pushing:"
	// adding a single bar, which will inherit container's width
	bar := p.Add(
		int64(total),
		mpb.NewBarFiller(mpb.BarStyle()),
		mpb.PrependDecorators(
			// display our name with one space on the right
			decor.Name(pbName, decor.WC{W: len(pbName) + 1, C: decor.DidentRight}),
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				decor.AverageETA(decor.ET_STYLE_GO, decor.WC{W: 4}), "Done!",
			),
		),
		mpb.AppendDecorators(
			// counter
			decor.Counters(0, "%d / %d  "),
			// percentage
			decor.Percentage(),
			// average
			// mpb.AppendDecorators(decor.AverageSpeed(decor.UnitKiB, "  % .1f")),
		),
	)

	log.Info("Begin to migrate ...")
	start := time.Now()

	report := reportutil.NewReport()
	if settings.Verbose {
		defer func() {
			log.Info("Migrate result:")
			report.Render(w)
		}()
	}

	if err := repository.ForEach(func(group, artifact, version, path, downloadUrl string) error {
		defer bar.Increment()
		if err1 := doNexusMigrate(path, downloadUrl, username, password); err1 != nil {
			if err1 == ErrFileConflict {
				report.AddSkippedResult(strings.Join([]string{group, artifact, version}, ":"), downloadUrl, "409 Conflict")
				return types.ErrForEachContinue
			}

			report.AddFailedResult(strings.Join([]string{group, artifact, version}, ":"), downloadUrl, err1.Error())

			if settings.FailFast {
				return errors.Wrapf(err1, "failed to migrate %s", path)
			}
		} else {
			report.AddSucceededResult(strings.Join([]string{group, artifact, version}, ":"), downloadUrl, "Succeeded")
		}

		return nil
	}); err != nil {
		return err
	}

	// wait for our bar to complete and flush
	p.Wait()

	log.Info("End to migrate.",
		logfields.Duration("duration", time.Now().Sub(start)),
		logfields.Int("succeededCount", len(report.SucceededResult)),
		logfields.Int("skippedCount", len(report.SkippedResult)),
		logfields.Int("failedCount", len(report.FailedResult)))

	return nil
}

func doMigrate(file, username, password string) error {
	u := getPushUrl(file)
	// log.Info("Put file:", logfields.String("file", file), logfields.String("url", u))
	resp, err := httputil.DefaultClient.PutFile(u, file, username, password)
	if err != nil {
		return err
	}
	defer ioutils.QuiteClose(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		if resp.StatusCode == http.StatusConflict {
			// log.Warn("409 Conflict: file has been pushed, and the strategy of destination is not overridable, so just skip it",
			// 	logfields.String("file", file))
			return ErrFileConflict
		}
		return errors.Errorf("got an unexpected response status: %s", resp.Status)
	}

	return nil
}

func doNexusMigrate(path, downloadUrl, username, password string) error {
	// download
	getResp, err := httputil.DefaultClient.GetWithAuth(downloadUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return errors.Wrapf(err, "failed to download from %s", downloadUrl)
	}
	defer ioutils.QuiteClose(getResp.Body)

	// push
	pushUrl := getPushUrl(path)
	resp, err := httputil.DefaultClient.Put(pushUrl, "", getResp.Body, username, password)
	if err != nil {
		return errors.Wrapf(err, "failed to push to %s", pushUrl)
	}
	defer ioutils.QuiteClose(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		if resp.StatusCode == http.StatusConflict {
			// log.Warn("409 Conflict: file has been pushed, and the strategy of destination is not overridable, so just skip it",
			// 	logfields.String("file", file))
			return ErrFileConflict
		}
		return errors.Errorf("got an unexpected response status: %s", resp.Status)
	}

	return nil
}

func GetRepository(repositoryPath string, maxFiles int) (repository *types.Repository, err error) {
	var fileCount int
	repository = &types.Repository{Path: repositoryPath}
	if err = filepath.WalkDir(repositoryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if fileutil.IsFileInvisible(d.Name()) {
				return filepath.SkipDir
			}
			if !types.ArtifactNameRegex.MatchString(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if fileutil.IsFileInvisible(d.Name()) ||
			d.Name() == "_remote.repositories" ||
			strings.HasPrefix(d.Name(), "_") {
			return nil
		}
		if maxFiles >= 0 && fileCount >= maxFiles { // FIXME
			return filepath.SkipDir
		}

		groupName, artifact, version, filename, err := getArtInfo(path, repositoryPath)
		if err != nil {
			return errors.Wrap(err, "failed to get artifact info")
		}
		repository.AddVersionFile(groupName, artifact, version, filename, path)
		fileCount++

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to walk repository")
	}

	return
}

func GetRepositoryFromNexusItems(repositoryUrl string, nexusItemList []nexus.Item) (repository *types.Repository, err error) {
	var fileCount int
	repository = &types.Repository{Path: repositoryUrl}
	for _, item := range nexusItemList {
		var groupName, artifact, version, filename string
		if !item.Maven2.IsEmpty() {
			groupName = item.Maven2.GroupId
			artifact = item.Maven2.ArtifactId
			version = item.Maven2.Version
			filename = path.Base(item.Path)
		} else {
			groupName, artifact, version, filename, err = getArtInfoFromSubPath(item.Path)
		}
		// SNAPSHOT ????????????
		repository.AddVersionFileBase(groupName, artifact, version, filename, item.Path, item.DownloadUrl)
		fileCount++
	}
	return
}

func getArtInfo(path, repositoryPath string) (groupName, artifact, version, filename string, err error) {
	// repositoryPath: /Users/chenxinyu/.m2/repository
	// path: /Users/chenxinyu/.m2/repository/org/kohsuke/stapler/json-lib/2.4-jenkins-2/json-lib-2.4-jenkins-2-sources.jar
	// subPath: org/kohsuke/stapler/json-lib/2.4-jenkins-2/json-lib-2.4-jenkins-2-sources.jar
	// filename: json-lib-2.4-jenkins-2-sources.jar
	subPath := strings.Trim(strings.TrimPrefix(path, repositoryPath), "/")
	filename = filepath.Base(path)

	subPathChunks := strings.Split(subPath, "/")
	size := len(subPathChunks)
	if size < 3 {
		return "", "", "", "", errors.New("invalid path")
	}
	version = subPathChunks[size-2]
	artifact = subPathChunks[size-3]
	groupName = strings.Join(subPathChunks[:size-3], ".")
	return
}

func getArtInfoFromSubPath(subPath string) (groupName, artifact, version, filename string, err error) {
	// subPath: org/kohsuke/stapler/json-lib/2.4-jenkins-2/json-lib-2.4-jenkins-2-sources.jar
	// filename: json-lib-2.4-jenkins-2-sources.jar
	filename = filepath.Base(subPath)

	subPathChunks := strings.Split(subPath, "/")
	size := len(subPathChunks)
	if size < 3 {
		return "", "", "", "", errors.New("invalid path")
	}
	// ????????? maven-metadata.xml ???????????? path ???????????? SNAPSHOT ????????????????????? version ????????????????????????
	// e.g. org/kohsuke/stapler/json-lib/maven-metadata.xml
	// ??? ????????????????????? version ?????????versionName = Metadata
	if strings.Contains(filename, MetadataXml) && !strings.Contains(subPathChunks[size-2], Snapshot) {
		version = Metadata
		artifact = subPathChunks[size-2]
		groupName = strings.Join(subPathChunks[:size-2], ".")
	} else {
		version = subPathChunks[size-2]
		artifact = subPathChunks[size-3]
		groupName = strings.Join(subPathChunks[:size-3], ".")
	}
	return
}

func getPushUrl(filePath string) string {
	subPath := strings.Trim(strings.TrimPrefix(filePath, settings.Src), "/")
	return strings.TrimSuffix(settings.Dst, "/") + "/" + subPath
}

func defaultMavenRepositoryPath() string {
	return filepath.Join(config.GetHomeDir(), ".m2", "repository")
}

func isLocalRepository(src string) bool {
	if strings.HasPrefix(src, "http") {
		return false
	}
	return true
}
