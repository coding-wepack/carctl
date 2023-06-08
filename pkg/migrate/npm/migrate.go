package npm

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"e.coding.net/codingcorp/carctl/pkg/api"
	"e.coding.net/codingcorp/carctl/pkg/config"
	"e.coding.net/codingcorp/carctl/pkg/constants"
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
		exists, err = api.FindDstRepoArtifactsName(&authConfig, settings.GetDstWithoutSlash(), constants.TypeNpm)
		if err != nil {
			return errors.Wrap(err, "failed to find dst repo exists artifacts")
		}
	}

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

	count := 0
	files := make([]remote.JfrogFile, 0)
	for _, f := range filesInfo.Res {
		if strings.HasSuffix(f.Name, ".tgz") {
			count++
			if settings.Force || isNeedMigrate(exists, f.Name) {
				files = append(files, f)
			}
		}
	}
	log.Infof("remote repository file count is:%d, need migrate count is:%d", count, len(files))

	if len(files) == 0 {
		if count > 0 {
			log.Info("all artifacts have been migrated")
			return nil
		}
		return errors.Errorf("npm repository: %s file not found, please check your repository or command parameters", repository)
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

	log.Info("Begin to migrate npm artifacts ...")
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
	authContent := fmt.Sprintf(npmrc, settings.GetDstHasSubSlash(), regUrl, username, regUrl, base64Pwd, regUrl, username)
	result, err := cmdutil.Command(fmt.Sprintf(initDir, authContent))
	if err != nil {
		return errors.Wrapf(err, "failed to init migrate data: %s", result)
	}
	defer clean(true)

	for i, file := range jfrogFileList {
		err := doMigrateJfrogArt(file.Name, fmt.Sprintf("%s/%s/%s", settings.GetSrcWithoutSlash(), file.Path, file.Name))
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
	for i := 0; i < 3; i++ {
		cmd := fmt.Sprintf(publish, path, settings.GetDstHasSubSlash())
		result, err = cmdutil.Command(cmd)
		if err == nil {
			break
		}
		log.Infof("failed to publish artifact, wait 1 second and try again! %s: err: %s", result, err.Error())
		time.Sleep(time.Second)
	}
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

func isNeedMigrate(exists map[string]bool, fileName string) bool {
	fileName = strings.TrimSuffix(fileName, ".tgz")
	split := strings.Split(fileName, "-")
	pkg := strings.Join(split[:len(split)-1], "-")
	version := split[len(split)-1]
	return !exists[fmt.Sprintf("%s:%s", pkg, version)]
}
