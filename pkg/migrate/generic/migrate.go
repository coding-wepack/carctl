package generic

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coding-wepack/carctl/pkg/action"
	"github.com/coding-wepack/carctl/pkg/api"
	"github.com/coding-wepack/carctl/pkg/config"
	"github.com/coding-wepack/carctl/pkg/constants"
	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/log/logfields"
	"github.com/coding-wepack/carctl/pkg/remote"
	reportutil "github.com/coding-wepack/carctl/pkg/report"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/httputil"
	"github.com/coding-wepack/carctl/pkg/util/ioutils"
	"github.com/coding-wepack/carctl/pkg/util/logutil"
	"github.com/coding-wepack/carctl/pkg/util/queueutil"
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
		// 过滤匹配 settings.Prefix 的制品
		var matchFiles []remote.JfrogFile
		for _, f := range filesInfo.Res {
			if strings.HasPrefix(f.GetFilePath(), settings.Prefix) {
				matchFiles = append(matchFiles, f)
			}
		}
		filesInfo.Res = matchFiles
	}

	var files []remote.JfrogFile
	for _, f := range filesInfo.Res {
		if isNeedMigrate(f.GetFilePath(), exists) {
			files = append(files, f)
		}
	}
	log.Infof("remote repository file count is:%d, need migrate count is:%d", len(filesInfo.Res), len(files))

	if len(files) == 0 {
		if len(filesInfo.Res) > 0 {
			log.Info("all artifacts have been migrated")
			return nil
		}
		return errors.Errorf("generic repository: %s file not found or files have been migrated, please check your repository or command", repository)
	}

	if err = migrateJfrogRepository(out, files, cfg.Username, cfg.Password); err != nil {
		return err
	}

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
			report.Render2(w)
		}()
	}

	sliceutil.QuickSortReverse(jfrogFileList, func(f remote.JfrogFile) int { return f.Size })
	dataChan := make(chan remote.JfrogFile)
	go queueutil.Producer(jfrogFileList, dataChan)

	var wg sync.WaitGroup
	var goroutineCount int32 = 0
	errChan := make(chan error)
	execJobNum := make([]int32, settings.Concurrency)
	for i := 0; i < settings.Concurrency; i++ {
		wg.Add(1)
		execJobNum[i] = 0
		go queueutil.Consumer(dataChan, errChan, &wg, &execJobNum[i], func(file remote.JfrogFile) error {
			atomic.AddInt32(&goroutineCount, 1)
			useTime, err := doMigrateJfrogArt(file.GetFilePath(), username, password)
			atomic.AddInt32(&goroutineCount, -1)
			bar.Increment()
			if err != nil && err == ErrFileConflict {
				report.AddSkippedResult2(file.Name, file.GetFilePath(), "409 Conflict", int64(file.Size), useTime)
				return nil
			} else if err != nil {
				report.AddFailedResult2(file.Name, file.GetFilePath(), err.Error(), int64(file.Size), useTime)
				if settings.FailFast {
					err = errors.Wrapf(err, "failed to migrate %s", file.GetFilePath())
					errChan <- err
				}
			} else {
				report.AddSucceededResult2(file.Name, file.GetFilePath(), "Succeeded", int64(file.Size), useTime)
			}
			return nil
		})
	}

	go logutil.WriteGoroutineFile(&goroutineCount, execJobNum)

	go func() {
		wg.Wait()
		// 关闭通道，表示所有的 goroutine 已经执行完毕
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
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

func doMigrateJfrogArt(path, username, password string) (useTime int64, err error) {
	start := time.Now()
	defer func() { useTime = time.Since(start).Milliseconds() }()
	downloadUrl := getDownloadUrl(path)
	pushUrl := getPushUrl(path)
	// if settings.Verbose {
	// 	log.Debug("do migrate jfrog artifacts",
	// 		logfields.String("downloadUrl", downloadUrl),
	// 		logfields.String("pushUrl", pushUrl))
	// }

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

func isNeedMigrate(filePath string, exists map[string]bool) bool {
	return !(exists[fmt.Sprintf("%s:latest", filePath)])
}
