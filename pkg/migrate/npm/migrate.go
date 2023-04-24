package npm

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"e.coding.net/codingcorp/carctl/pkg/remote"
	"e.coding.net/codingcorp/carctl/pkg/util/cmdutil"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/log/logfields"
	"e.coding.net/codingcorp/carctl/pkg/migrate/maven/types"
	reportutil "e.coding.net/codingcorp/carctl/pkg/report"
	"e.coding.net/codingcorp/carctl/pkg/settings"
	"e.coding.net/codingcorp/carctl/pkg/util/fileutil"
	"e.coding.net/codingcorp/carctl/pkg/util/httputil"
	"e.coding.net/codingcorp/carctl/pkg/util/ioutils"
)

const (
	initDir  = "mkdir ./npmCache && echo '%s' > ./.npmrc"
	cleanDir = "rm -rf ./npmCache"
	tarFile  = "./npmCache/%s"
	unTar    = "cd ./npmCache && rm -rf ./%s && mkdir ./%s && tar -xf %s -C %s"
	publish  = "cd ./npmCache/%s/package && cp ../../../.npmrc . && npm publish --registry=%s"
	npmrc    = `registry=%s
always-auth=true
//%s:username=%s
//%s:_password=%s
//%s:email=%s`
)

var ErrFileConflict = errors.New("failed to put file: 409 conflict")

func Migrate(cfg *action.Configuration, out io.Writer) error {
	isLocalPath := isLocalRepository(settings.Src)
	if isLocalPath {
		// local repository
		// return MigrateFromDisk(cfg, out)
		return errors.New("unsupported migrate local npm artifacts")
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
	// 默认为 nexus
	if settings.SrcType == "" {
		settings.SrcType = "nexus"
	}
	switch settings.SrcType {
	case "jfrog":
		return MigrateFromJfrog(cfg, out, srcUrl)
	default:
		return errors.Errorf("This src-type [%s] is not supported", settings.SrcType)
	}
}

func MigrateFromJfrog(cfg *action.Configuration, out io.Writer, jfrogUrl *url.URL) error {
	log.Infof("Get file list from source repository [%s] ...", settings.Src)
	// 获取仓库名称
	urlPathStrs := strings.Split(strings.Trim(jfrogUrl.Path, "/"), "/")
	repository := urlPathStrs[1]

	filesInfo, err := remote.FindFileListFromJfrog(jfrogUrl, repository)
	if err != nil {
		return errors.Wrap(err, "failed to get file list")
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

	files := make([]remote.JfrogFile, 0)
	for _, f := range filesInfo.Res {
		if strings.HasSuffix(f.Name, ".tgz") {
			files = append(files, f)
		}
	}
	if err = migrateJfrogRepository(out, files, authConfig.Username, authConfig.Password); err != nil {
		return err
	}

	return nil
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

func migrateJfrogRepository(w io.Writer, jfrogFileList []remote.JfrogFile, username, password string) error {
	// Progress Bar
	// initialize progress container, with custom width
	p := mpb.New(mpb.WithWidth(80))
	const pbName = "Pushing:"
	// adding a single bar, which will inherit container's width
	bar := p.Add(
		int64(len(jfrogFileList)),
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

	// 创建临时文件夹以及鉴权文件
	clean(true)
	regUrl := getRegUrl(settings.Dst)
	base64Pwd := base64.StdEncoding.EncodeToString([]byte(password))
	authContent := fmt.Sprintf(npmrc, settings.GetDst()+"/", regUrl, username, regUrl, base64Pwd, regUrl, username)
	result, err := cmdutil.Command(fmt.Sprintf(initDir, authContent))
	if err != nil {
		return errors.Wrapf(err, "failed to init migrate data: %s", result)
	}
	defer clean(true)

	for i, file := range jfrogFileList {
		err := doMigrateJfrogArt(file.Name, fmt.Sprintf("%s/%s/%s", settings.GetSrc(), file.Path, file.Name))
		bar.Increment()
		if err != nil && err == ErrFileConflict {
			report.AddSkippedResult(file.Name, file.GetFilePath(), "409 Conflict")
			return types.ErrForEachContinue
		} else if err != nil {
			report.AddFailedResult(file.Name, file.GetFilePath(), err.Error())
			if settings.FailFast {
				return errors.Wrapf(err, "failed to migrate %s", file.GetFilePath())
			}
		} else {
			report.AddSucceededResult(file.Name, file.GetFilePath(), "Succeeded")
		}
		if i%100 == 0 {
			clean(false)
		}
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

func doMigrateJfrogArt(fileName, downloadUrl string) error {
	// download
	getResp, err := httputil.DefaultClient.GetWithAuth(downloadUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return errors.Wrapf(err, "failed to download from %s", downloadUrl)
	}
	defer ioutils.QuiteClose(getResp.Body)
	err = writeZipFile(fileName, getResp.Body)
	if err != nil {
		return err
	}

	// unzip
	path := strings.TrimSuffix(fileName, ".tgz")
	result, err := cmdutil.Command(fmt.Sprintf(unTar, path, path, fileName, path))
	if err != nil {
		return errors.Wrapf(err, "failed to unzip file %s: %s", fileName, result)
	}

	// upload
	result, err = cmdutil.Command(fmt.Sprintf(publish, path, settings.Dst))
	if err != nil {
		return errors.Wrapf(err, "failed to publish artifact: %s", result)
	}
	return nil
}

func writeZipFile(fileName string, read io.ReadCloser) error {
	filePath := fmt.Sprintf(tarFile, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", filePath)
	}
	_, err = io.Copy(file, read)
	if err != nil {
		return errors.Wrapf(err, "failed to write content to file")
	}
	return nil
}

func clean(thorough bool) {
	cmd := cleanDir
	if !thorough {
		cmd = cleanDir + "/*"
	}
	result, err := cmdutil.Command(cmd)
	if err != nil {
		log.Warnf("clean migrate dir failed: %s : %s", err.Error(), result)
	}
	if thorough {
		result, err = cmdutil.Command("rm -rf ./.npmrc")
		if err != nil {
			log.Warnf("clean migrate auth file failed: %s : %s", err.Error(), result)
		}
	}
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

func getPushUrl(filePath string) string {
	subPath := strings.Trim(strings.TrimPrefix(filePath, settings.Src), "/")
	return strings.TrimSuffix(settings.Dst, "/") + "/" + subPath
}

func getRegUrl(url string) string {
	url = strings.Trim(url, "http://")
	url = strings.Trim(url, "https://")
	return strings.Trim(url, "/") + "/"
}

func isLocalRepository(src string) bool {
	if strings.HasPrefix(src, "http") {
		return false
	}
	return true
}
