package docker

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/coding-wepack/carctl/pkg/action"
	"github.com/coding-wepack/carctl/pkg/api"
	"github.com/coding-wepack/carctl/pkg/config"
	"github.com/coding-wepack/carctl/pkg/constants"
	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/log/logfields"
	"github.com/coding-wepack/carctl/pkg/migrate/docker/types"
	"github.com/coding-wepack/carctl/pkg/remote"
	reportutil "github.com/coding-wepack/carctl/pkg/report"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/cmdutil"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var (
	ErrFileConflict = errors.New("failed to put file: 409 conflict")
)

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
		exists, err = api.FindDstRepoArtifactsName(&authConfig, settings.GetDstWithoutSlash(), constants.TypeDocker)
		if err != nil {
			return errors.Wrap(err, "failed to find dst repo exists artifacts")
		}
	}

	if isLocalRepository(settings.Src) {
		// TODO local repository
		// return MigrateFromDisk(cfg, out)
		return errors.New("unsupported migrate local docker artifacts")
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
	repository := strings.Trim(jfrogUrl.Path, "/")
	filesInfo, err := remote.FindFileListFromJfrog(jfrogUrl, repository)
	if err != nil {
		return errors.Wrap(err, "failed to get file list")
	}

	if err = migrateJfrogRepository(out, jfrogUrl, filesInfo.Res, cfg, exists); err != nil {
		return err
	}
	return nil
}

func migrateJfrogRepository(w io.Writer, jfrogUrl *url.URL, jfrogFileList []remote.JfrogFile, auth *config.AuthConfig, exists map[string]bool) error {
	log.Info("Scanning jfrog repository ...")

	repository, err := GetRepositoryFromJfrogFile(jfrogUrl, jfrogFileList, exists)
	if err != nil {
		return err
	}
	log.Info("Successfully to scan the repository", logfields.Int("images count", repository.Count))
	if repository.Count == 0 {
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
	const pbName = "Pushing:"
	// adding a single bar, which will inherit container's width
	bar := p.Add(
		int64(repository.Count),
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

	log.Info("Begin to migrate docker artifacts ...")
	start := time.Now()

	report := reportutil.NewReport()
	if settings.Verbose {
		defer func() {
			log.Info("Migrate result:")
			report.Render(w)
		}()
	}

	if err = repository.ForEach(func(name, srcTag, dstTag string, isTlsSrc, isTlsDst bool) error {
		defer bar.Increment()
		if err1 := doMigrateJfrogArt(srcTag, dstTag, isTlsSrc, isTlsDst, auth); err1 != nil {
			if err1 == ErrFileConflict {
				report.AddSkippedResult(name, srcTag, "409 Conflict")
				return types.ErrForEachContinue
			}

			report.AddFailedResult(name, srcTag, err1.Error())

			if settings.FailFast {
				return errors.Wrapf(err1, "failed to migrate %s", srcTag)
			}
		} else {
			report.AddSucceededResult(name, srcTag, "Succeeded")
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

func doMigrateJfrogArt(srcTag, dstTag string, isTlsSrc, isTlsDst bool, auth *config.AuthConfig) error {
	cmd := fmt.Sprintf("skopeo copy --src-creds=%s:%s --dest-creds=%s:%s --src-tls-verify=%t --dest-tls-verify=%t docker://%s docker://%s",
		settings.SrcUsername, settings.SrcPassword, auth.Username, auth.Password, isTlsSrc, isTlsDst, srcTag, dstTag)
	if settings.Verbose {
		log.Debug(cmd)
	}
	result, err := cmdutil.Command(cmd)
	if err != nil {
		return errors.Wrapf(err, "failed to migrate image from %s to %s, result: %s", srcTag, dstTag, result)
	}

	if settings.Verbose {
		log.Debug(result)
	}
	return nil
}

func GetRepositoryFromJfrogFile(jfrogUrl *url.URL, jfrogFileList []remote.JfrogFile, exists map[string]bool) (repository *types.Repository, err error) {
	fileCount := 0
	isTls := strings.EqualFold(jfrogUrl.Scheme, "https")
	repositoryUrl := fmt.Sprintf("%s%s", jfrogUrl.Host, jfrogUrl.Path)
	repository = &types.Repository{IsTls: isTls, Path: repositoryUrl}
	for _, f := range jfrogFileList {
		if !strings.EqualFold(f.Name, "manifest.json") {
			continue
		}
		srcPath, pkg, version, err := f.GetDockerInfo()
		if err != nil {
			log.Warnf("failed to gat docker tag from file srcPath: %s", f.Path)
			continue
		}
		imageTag := &types.Image{
			SrcPath: strings.Trim(srcPath, "/"),
			PkgName: strings.Trim(pkg, "/"),
			Version: strings.Trim(version, "/"),
		}
		fileCount++
		if settings.Force || isNeedMigrate(exists, imageTag) {
			repository.Images = append(repository.Images, imageTag)
			repository.Count++
		}
	}
	log.Infof("remote repository file count is:%d, need migrate count is:%d", fileCount, repository.Count)
	return
}

func isLocalRepository(src string) bool {
	if strings.HasPrefix(src, "http") {
		return false
	}
	return true
}

func isNeedMigrate(exists map[string]bool, imageTag *types.Image) bool {
	if settings.Force {
		return true
	}
	return !exists[imageTag.SrcPath]
}
