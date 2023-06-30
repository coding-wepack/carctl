package generic

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coding-wepack/carctl/pkg/action"
	"github.com/coding-wepack/carctl/pkg/api"
	"github.com/coding-wepack/carctl/pkg/config"
	"github.com/coding-wepack/carctl/pkg/constants"
	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/log/logfields"
	"github.com/coding-wepack/carctl/pkg/migrate/generic/types"
	"github.com/coding-wepack/carctl/pkg/remote"
	reportutil "github.com/coding-wepack/carctl/pkg/report"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/httputil"
	"github.com/coding-wepack/carctl/pkg/util/ioutils"
	"github.com/coding-wepack/carctl/pkg/util/sliceutil"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var ErrFileConflict = errors.New("failed to put file: 409 conflict")

func Migrate(cfg *action.Configuration, out io.Writer) error {
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
	var exists map[string]bool
	if !settings.Force {
		exists, err = api.FindDstExistsArtifacts(&authConfig, settings.GetDstWithoutSlash(), constants.TypeGeneric)
		if err != nil {
			return errors.Wrap(err, "failed to find dst repo exists artifacts")
		}
	}

	isLocalPath := isLocalRepository(settings.Src)
	if isLocalPath {
		// TODO local repository
		// return MigrateFromDisk(cfg, out)
		return errors.New("unsupported migrate local generic artifacts")
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
		return MigrateFromUrl(&authConfig, out, srcUrl, exists)
	}
}

func MigrateFromUrl(cfg *config.AuthConfig, out io.Writer, srcUrl *url.URL, exists map[string]bool) error {
	// 默认为 nexus
	if settings.SrcType == "" {
		settings.SrcType = "nexus"
	}
	switch settings.SrcType {
	case "jfrog":
		return MigrateFromJfrog(cfg, out, srcUrl, exists)
	default:
		return errors.Errorf("This src-type [%s] is not supported", settings.SrcType)
	}
}

func MigrateFromJfrog(cfg *config.AuthConfig, out io.Writer, jfrogUrl *url.URL, exists map[string]bool) error {
	log.Infof("Get file list from source repository [%s] ...", settings.Src)
	// 获取仓库名称
	urlPathStrs := strings.Split(strings.Trim(jfrogUrl.Path, "/"), "/")
	repository := urlPathStrs[1]

	filesInfo, err := remote.FindFileListFromJfrog(jfrogUrl, repository)
	if err != nil {
		return errors.Wrap(err, "failed to get file list")
	}

	if len(settings.Prefix) != 0 {
		totalCount := len(filesInfo.Res)
		// 过滤匹配 settings.Prefix 的制品
		var matchFiles []remote.JfrogFile
		for _, f := range filesInfo.Res {
			if strings.HasPrefix(f.GetFilePath(), settings.Prefix) {
				matchFiles = append(matchFiles, f)
			}
		}
		filesInfo.Res = matchFiles
		log.Infof("remote repository file count: %d, match prefix count: %d", totalCount, len(matchFiles))
	}

	if len(filesInfo.Res) == 0 {
		return errors.Errorf("generic repository: %s file not found, please check your repository or command", repository)
	}

	if err = migrateJfrogRepository(out, jfrogUrl, filesInfo.Res, cfg.Username, cfg.Password, exists); err != nil {
		return err
	}

	return nil
}

func migrateJfrogRepository(w io.Writer, jfrogUrl *url.URL, jfrogFileList []remote.JfrogFile, username, password string, exists map[string]bool) error {
	log.Info("Scanning jfrog repository ...")

	sliceutil.QuickSortReverse(jfrogFileList, func(f remote.JfrogFile) int64 { return f.Size })
	repository, err := GetRepositoryFromJfrogFile(jfrogUrl, jfrogFileList, exists)
	if err != nil {
		return err
	}
	log.Info("Successfully to scan the repository", logfields.Int("file count", repository.Count))
	if repository.Count == 0 {
		log.Warn("no files found or files have been migrated, no need to migrate")
		return nil
	}
	if settings.Verbose {
		log.Info("Repository Info:")
		repository.Render(w)
	}
	if settings.DryRun {
		return nil
	}

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
			report.RenderV2(w)
		}()
	}

	if err = repository.ParallelForEach(func(fileName, filePath string, size int64) error {
		useTime, err := doMigrateJfrogArt(filePath, username, password)
		bar.Increment()
		if err != nil && err == ErrFileConflict {
			report.AddSkippedResultV2(fileName, filePath, "409 Conflict", size, useTime)
			return nil
		} else if err != nil {
			report.AddFailedResultV2(fileName, filePath, err.Error(), size, useTime)
			if settings.FailFast {
				err = errors.Wrapf(err, "failed to migrate %s", filePath)
			}
		} else {
			report.AddSucceededResultV2(fileName, filePath, "Succeeded", size, useTime)
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

func doMigrateJfrogArt(path, username, password string) (useTime int64, err error) {
	start := time.Now()
	defer func() { useTime = time.Since(start).Milliseconds() }()
	downloadUrl := getDownloadUrl(path)
	pushUrl := getPushUrl(path)

	// download
	getResp, err := httputil.DefaultClient.GetWithAuth(downloadUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return useTime, errors.Wrapf(err, "failed to download from %s", downloadUrl)
	}
	defer ioutils.QuiteClose(getResp.Body)

	// push
	resp, err := httputil.DefaultClient.Put(pushUrl, "", getResp.Body, username, password)
	if err != nil {
		return useTime, errors.Wrapf(err, "failed to push to %s", pushUrl)
	}
	defer ioutils.QuiteClose(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		if resp.StatusCode == http.StatusConflict {
			// log.Warn("409 Conflict: file has been pushed, and the strategy of destination is not overridable, so just skip it",
			// 	logfields.String("file", file))
			return useTime, ErrFileConflict
		}
		return useTime, errors.Errorf("got an unexpected response status: %s", resp.Status)
	}

	return useTime, nil
}

func getDownloadUrl(filePath string) string {
	subPath := strings.Trim(strings.TrimPrefix(filePath, settings.Src), "/")
	return strings.TrimSuffix(settings.Src, "/") + "/" + subPath
}

func getPushUrl(filePath string) string {
	subPath := strings.Trim(strings.TrimPrefix(filePath, settings.Src), "/")
	return settings.GetDstHasSubSlash() + subPath
}

func isLocalRepository(src string) bool {
	if strings.HasPrefix(src, "http") {
		return false
	}
	return true
}

func GetRepositoryFromJfrogFile(jfrogUrl *url.URL, jfrogFileList []remote.JfrogFile, exists map[string]bool) (repository *types.Repository, err error) {
	fileCount := 0
	repositoryUrl := fmt.Sprintf("%s%s", jfrogUrl.Host, jfrogUrl.Path)
	repository = &types.Repository{Path: repositoryUrl}
	for _, f := range jfrogFileList {
		file := &types.File{
			FileName: f.Name,
			FilePath: f.GetFilePath(),
			Size:     f.Size,
		}
		fileCount++
		if settings.Force || isNeedMigrate(file, exists) {
			repository.Files = append(repository.Files, file)
			repository.Count++
		}
	}
	log.Infof("remote repository file count:%d, need migrate count:%d", fileCount, repository.Count)
	return
}

func isNeedMigrate(file *types.File, exists map[string]bool) bool {
	return !(exists[fmt.Sprintf("%s:latest", file.FilePath)])
}
