package maven

import (
	"bufio"
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
	"sync/atomic"
	"time"

	"github.com/coding-wepack/carctl/pkg/action"
	"github.com/coding-wepack/carctl/pkg/api"
	"github.com/coding-wepack/carctl/pkg/config"
	"github.com/coding-wepack/carctl/pkg/constants"
	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/log/logfields"
	"github.com/coding-wepack/carctl/pkg/migrate/maven/types"
	"github.com/coding-wepack/carctl/pkg/migrate/maven/types/nexus"
	"github.com/coding-wepack/carctl/pkg/remote"
	reportutil "github.com/coding-wepack/carctl/pkg/report"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/fileutil"
	"github.com/coding-wepack/carctl/pkg/util/httputil"
	"github.com/coding-wepack/carctl/pkg/util/ioutils"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
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

	// exists artifacts
	var existsVersions map[string]bool
	var existsFiles map[string]bool
	if !settings.Force {
		existsFiles, err = api.FindDstExistsFiles(&authConfig, settings.GetDstWithoutSlash(), constants.TypeMaven)
		if err != nil {
			return errors.Wrap(err, "failed to find dst repo exists files")
		}
		existsVersions, err = api.FindDstExistsArtifacts(&authConfig, settings.GetDstWithoutSlash(), constants.TypeMaven)
		if err != nil {
			return errors.Wrap(err, "failed to find dst repo exists artifacts")
		}
	}

	isLocalPath := isLocalRepository(settings.Src)
	if isLocalPath {
		// local repository
		return MigrateFromDisk(&authConfig, out, existsVersions, existsFiles)
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
		return MigrateFromUrl(&authConfig, out, srcUrl, existsVersions, existsFiles)
	}
}

func MigrateFromDisk(cfg *config.AuthConfig, out io.Writer, existsVersions, existsFiles map[string]bool) error {
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

	if err = migrateRepository(out, cfg.Username, cfg.Password, existsVersions, existsFiles); err != nil {
		return err
	}

	return nil
}

func MigrateFromUrl(cfg *config.AuthConfig, out io.Writer, srcUrl *url.URL, existsVersions, existsFiles map[string]bool) error {
	// 默认为 nexus
	if settings.SrcType == "" {
		settings.SrcType = "nexus"
	}
	switch settings.SrcType {
	case "nexus":
		return MigrateFromNexus(cfg, out, srcUrl, existsVersions, existsFiles)
	case "jfrog":
		return MigrateFromJfrog(cfg, out, srcUrl, existsVersions, existsFiles)
	default:
		return errors.Errorf("This src-type [%s] is not supported", settings.SrcType)
	}
}

func MigrateFromNexus(cfg *config.AuthConfig, out io.Writer, nexusUrl *url.URL, existsVersions, existsFiles map[string]bool) error {
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
		// TODO: 是否要做个最大限制跳出循环，并输出 continuationToken 让用户手动执行下一次同步
	}

	if err := migrateNexusRepository(out, nexusItemList, cfg.Username, cfg.Password, existsVersions, existsFiles); err != nil {
		return err
	}

	return nil
}

func MigrateFromJfrog(cfg *config.AuthConfig, out io.Writer, jfrogUrl *url.URL, existsVersions, existsFiles map[string]bool) error {
	log.Infof("Get file list from source repository [%s] ...", settings.Src)
	// 获取仓库名称
	urlPathStrs := strings.Split(strings.Trim(jfrogUrl.Path, "/"), "/")
	repository := urlPathStrs[1]

	filesInfo, err := remote.FindFileListFromJfrog(jfrogUrl, repository)
	if err != nil {
		return errors.Wrap(err, "failed to get file list")
	}

	if err = migrateJfrogRepository(out, filesInfo.Res, cfg.Username, cfg.Password, existsVersions, existsFiles); err != nil {
		return err
	}

	return nil
}

// getFileListFromNexus 使用 nexus3 API 来获取文件列表
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
		// TODO: 优化的优雅一点，临时修复比较粗糙
		// 如果状态码为 404，则尝试兼容老版本的 nexus3.x，API 是带有 /nexus 前缀的
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

func migrateRepository(w io.Writer, username, password string, existsVersions, existsFiles map[string]bool) error {
	log.Info("Scanning repository ...")

	repository, err := GetRepository(settings.Src, settings.MaxFiles, existsVersions, existsFiles)
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
		log.Warn("no files found or files have been migrated, no need to migrate")
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

	var count int32
	if err := repository.ParallelForEach(func(group, artifact, version, path, downloadUrl string, size int64) error {
		defer bar.Increment()
		atomic.AddInt32(&count, 1)
		if err1 := doLocalMigrate(path, username, password); err1 != nil {
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
		atomic.AddInt32(&count, -1)

		return nil
	}); err != nil {
		return err
	}

	go func() {
		file, err := os.Create("goroutine.txt")
		if err != nil {
			log.Error("", logfields.Error(err))
			return
		}
		defer file.Close()
		write := bufio.NewWriter(file)

		for {
			_, err = write.WriteString(fmt.Sprintf("%s goroutine number : %d\n", time.Now().Format("15:04:05.000"), count))
			write.Flush()
			if err != nil {
				log.Error("", logfields.Error(err))
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()

	// wait for our bar to complete and flush
	p.Wait()

	log.Info("End to migrate.",
		logfields.Duration("duration", time.Now().Sub(start)),
		logfields.Int("succeededCount", len(report.SucceededResult)),
		logfields.Int("skippedCount", len(report.SkippedResult)),
		logfields.Int("failedCount", len(report.FailedResult)))

	return nil
}

func migrateJfrogRepository(w io.Writer, jfrogFiles []remote.JfrogFile, username, password string, existsVersions, existsFiles map[string]bool) error {
	log.Info("Scanning jfrog repository ...")

	repository, err := GetRepositoryFromJfrogFile(settings.Src, jfrogFiles, existsVersions, existsFiles)
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
		log.Warn("no files found or files have been migrated, no need to migrate")
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

	log.Info("Begin to migrate maven artifacts ...")
	start := time.Now()

	report := reportutil.NewReport()
	if settings.Verbose {
		defer func() {
			log.Info("Migrate result:")
			report.Render2(w)
		}()
	}

	if err = repository.ParallelForEach(func(group, artifact, version, path, downloadUrl string, size int64) error {
		defer bar.Increment()
		if useTime, err1 := doRemoteMigrate(path, downloadUrl, username, password); err1 != nil {
			if err1 == ErrFileConflict {
				report.AddSkippedResultV2(strings.Join([]string{group, artifact, version}, ":"), downloadUrl, "409 Conflict", size, useTime)
				return types.ErrForEachContinue
			}

			report.AddFailedResultV2(strings.Join([]string{group, artifact, version}, ":"), downloadUrl, err1.Error(), size, useTime)

			if settings.FailFast {
				return errors.Wrapf(err1, "failed to migrate %s", path)
			}
		} else {
			report.AddSucceededResultV2(strings.Join([]string{group, artifact, version}, ":"), downloadUrl, "Succeeded", size, useTime)
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

func migrateNexusRepository(w io.Writer, nexusItemList []nexus.Item, username, password string, existsVersions, existsFiles map[string]bool) error {
	log.Info("Scanning nexus repository ...")

	repository, err := GetRepositoryFromNexusItems(settings.Src, nexusItemList, existsVersions, existsFiles)
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
		log.Warn("no files found or files have been migrated, no need to migrate")
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

	if err := repository.ParallelForEach(func(group, artifact, version, path, downloadUrl string, size int64) error {
		defer bar.Increment()
		if _, err1 := doRemoteMigrate(path, downloadUrl, username, password); err1 != nil {
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

func doLocalMigrate(file, username, password string) error {
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

func doRemoteMigrate(path, downloadUrl, username, password string) (useTime int64, err error) {
	start := time.Now()
	defer func() { useTime = time.Since(start).Milliseconds() }()
	var resp *http.Response
	for i := 0; i < 3; i++ {
		resp, err = downloadAndUpload(path, downloadUrl, username, password)
		if err == nil {
			break
		}
		log.Warn("migrate maven artifacts failed, retry in 1 second...")
		time.Sleep(time.Second)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		if resp.StatusCode == http.StatusConflict {
			// log.Warn("409 Conflict: file has been pushed, and the strategy of destination is not overridable, so just skip it",
			// 	logfields.String("file", file))
			return useTime, ErrFileConflict
		}
		return useTime, errors.Errorf("got an push unexpected response status: %s", resp.Status)
	}

	return useTime, nil
}

func downloadAndUpload(path, downloadUrl, username, password string) (resp *http.Response, err error) {
	// download
	getResp, err := httputil.DefaultClient.GetWithAuth(downloadUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to download from %s", downloadUrl)
	}
	defer ioutils.QuiteClose(getResp.Body)

	// push
	pushUrl := getPushUrl(path)
	resp, err = httputil.DefaultClient.Put(pushUrl, "", getResp.Body, username, password)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to push to %s", pushUrl)
	}
	defer ioutils.QuiteClose(resp.Body)
	return
}

func GetRepository(repositoryPath string, maxFiles int, existsVersions, existsFiles map[string]bool) (repository *types.Repository, err error) {
	var fileCount int
	var needMigrateFileCount int
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
		if maxFiles >= 0 && needMigrateFileCount >= maxFiles { // FIXME
			return filepath.SkipDir
		}

		groupName, artifact, version, filename, err := getArtInfo(path, repositoryPath)
		if err != nil {
			return errors.Wrap(err, "failed to get artifact info")
		}
		fileCount++
		if settings.Force || isNeedMigrate(existsVersions, existsFiles, groupName, artifact, version, filename) {
			repository.AddVersionFile(groupName, artifact, version, filename, path)
			needMigrateFileCount++
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to walk repository")
	}
	// 上面逻辑处理文件的时候，包含 Metadata 文件，因不属于制品版本级别，所以无法通过已存在制品版本来过滤，此处单独过滤
	needMigrateFileCount = repository.CleanInvalidMetadata(needMigrateFileCount)

	log.Infof("repository file count is:%d, need migrate count is:%d", fileCount, needMigrateFileCount)
	return
}

func GetRepositoryFromJfrogFile(repositoryUrl string, jfrogFiles []remote.JfrogFile, existsVersions, existsFiles map[string]bool) (repository *types.Repository, err error) {
	var fileCount int
	var needMigrateFileCount int
	repository = &types.Repository{Path: repositoryUrl}
	for _, file := range jfrogFiles {
		subPath := fmt.Sprintf("%s/%s", file.Path, file.Name)
		groupName, artifact, version, filename, err := getArtInfoFromSubPath(subPath)
		if err != nil {
			log.Warnf("get maven info failed: %s", err.Error())
			continue
		}
		fileCount++
		if settings.Force || isNeedMigrate(existsVersions, existsFiles, groupName, artifact, version, filename) {
			downloadUrl := fmt.Sprintf("%s/%s", settings.GetSrcWithoutSlash(), subPath)
			repository.AddVersionFileBase(groupName, artifact, version, filename, subPath, downloadUrl, int64(file.Size))
			needMigrateFileCount++
		}
	}

	// 上面逻辑处理文件的时候，包含 Metadata 文件，因不属于制品版本级别，所以无法通过已存在制品版本来过滤，此处单独过滤
	needMigrateFileCount = repository.CleanInvalidMetadata(needMigrateFileCount)

	log.Infof("remote repository file count is:%d, need migrate count is:%d", fileCount, needMigrateFileCount)
	return
}

func GetRepositoryFromNexusItems(repositoryUrl string, nexusItemList []nexus.Item, existsVersions, existsFiles map[string]bool) (repository *types.Repository, err error) {
	var fileCount int
	var needMigrateFileCount int
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
		fileCount++
		if settings.Force || isNeedMigrate(existsVersions, existsFiles, groupName, artifact, version, filename) {
			// SNAPSHOT 版本特例
			repository.AddVersionFileBase(groupName, artifact, version, filename, item.Path, item.DownloadUrl, 0)
			needMigrateFileCount++
		}
	}
	// 上面逻辑处理文件的时候，包含 Metadata 文件，因不属于制品版本级别，所以无法通过已存在制品版本来过滤，此处单独过滤
	needMigrateFileCount = repository.CleanInvalidMetadata(needMigrateFileCount)

	log.Infof("remote repository file count is:%d, need migrate count is:%d", fileCount, needMigrateFileCount)
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
		return "", "", "", "", errors.Errorf("invalid maven path: %s", subPath)
	}
	// 如果以 maven-metadata.xml 结尾，但 path 中不包含 SNAPSHOT 字样，此文件为 version 上层文件夹路径下
	// e.g. org/kohsuke/stapler/json-lib/maven-metadata.xml
	// 将 此文件作为特殊 version 对待，versionName = Metadata
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
	return settings.GetDstHasSubSlash() + subPath
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

func isNeedMigrate(existsVersions, existsFiles map[string]bool, groupName, artifact, version, filename string) bool {
	if settings.Force {
		return true
	}
	// 检查制品是否存在
	if !strings.EqualFold("Metadata", version) {
		artifactName := fmt.Sprintf("%s:%s:%s", groupName, artifact, version)
		if !(existsVersions[artifactName]) {
			// 制品不存在，需要迁移
			return true
		}
	}
	// 制品存在，判断文件是否存在
	var fileName string
	groupName = strings.Join(strings.Split(groupName, "."), "/")
	if strings.EqualFold("Metadata", version) {
		fileName = join("/", groupName, artifact, filename)
	} else {
		fileName = join("/", groupName, artifact, version, filename)
	}
	return !existsFiles[fileName]
}

func join(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}
