package pypi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/coding-wepack/carctl/pkg/action"
	"github.com/coding-wepack/carctl/pkg/api"
	"github.com/coding-wepack/carctl/pkg/config"
	"github.com/coding-wepack/carctl/pkg/constants"
	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/log/logfields"
	"github.com/coding-wepack/carctl/pkg/migrate/pypi/types"
	"github.com/coding-wepack/carctl/pkg/migrate/pypi/types/nexus"
	reportutil "github.com/coding-wepack/carctl/pkg/report"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/httputil"
	"github.com/coding-wepack/carctl/pkg/util/ioutils"
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
		exists, err = api.FindDstRepoArtifactsName(&authConfig, settings.GetDstWithoutSlash(), constants.TypePypi)
		if err != nil {
			return errors.Wrap(err, "failed to find dst repo exists artifacts")
		}
	}

	srcUrl, err := url.Parse(settings.Src)
	if err != nil || srcUrl.Scheme == "" {
		log.Info("src is not url, only support migrate from remote repository, eg: nexus", logfields.String("src", settings.Src))
		if settings.Verbose {
			log.Warn("Can't parse with error", logfields.Error(err))
		}
		return errors.New("source repository is not remote repository")
	} else {
		return MigrateFromUrl(&authConfig, out, srcUrl, exists)
	}
}

func getFileListFromNexus(scheme, nexusHost, repository, continuationToken string) (*nexus.GetAssetsResponse, error) {
	apiUrl := fmt.Sprintf("%s://%s/service/rest/v1/assets?repository=%s", scheme, nexusHost, repository)
	if continuationToken != "" {
		apiUrl = fmt.Sprintf("%s&continuationToken=%s", apiUrl, continuationToken)
	}

	resp, err := httputil.DefaultClient.GetWithAuth(apiUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get components: %s", apiUrl)
	}
	if resp.StatusCode != http.StatusOK {
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

func MigrateFromUrl(cfg *config.AuthConfig, out io.Writer, srcUrl *url.URL, exists map[string]bool) error {
	if settings.SrcType == "" {
		settings.SrcType = "nexus"
	}
	switch settings.SrcType {
	case "nexus":
		return MigrateFromNexus(cfg, out, srcUrl, exists)
	default:
		return errors.Errorf("This src-type [%s] is not supported", settings.SrcType)
	}
}

func MigrateFromNexus(cfg *config.AuthConfig, out io.Writer, nexusUrl *url.URL, exists map[string]bool) error {
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
	}

	if err := migrateNexusRepository(out, nexusItemList, cfg.Username, cfg.Password, exists); err != nil {
		return err
	}

	return nil
}

func GetRepositoryFromNexusItems(repositoryUrl string, nexusItemList []nexus.Item, exists map[string]bool) (repository *types.Repository, err error) {
	var fileCount int
	repository = &types.Repository{Path: repositoryUrl}
	for _, item := range nexusItemList {
		// may be some filter
		if item.Pypi.Name != "" && item.Pypi.Version != "" {
			fileCount++
			if settings.Force || isNeedMigrate(item.Pypi.Name, item.Pypi.Version, exists) {
				repository.AddVersionFile(item)
				repository.FileCount++
			}
		}
	}
	log.Infof("remote repository file count is:%d, need migrate count is:%d", fileCount, repository.FileCount)
	return
}

func migrateNexusRepository(w io.Writer, nexusItemList []nexus.Item, username, password string, exists map[string]bool) error {
	log.Info("Begin to migrate ...")

	repository, err := GetRepositoryFromNexusItems(settings.Src, nexusItemList, exists)
	if err != nil {
		return err
	}

	if repository.FileCount == 0 {
		log.Warn("no files found, no need to migrate")
		return nil
	}
	if settings.Verbose {
		log.Info("Repository Info:")
		repository.Render(w)
	}

	p := mpb.New(mpb.WithWidth(80))
	total := repository.FileCount
	const pbName = "Pushing:"
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

	start := time.Now()

	report := reportutil.NewReport()
	if settings.Verbose {
		defer func() {
			log.Info("Migrate result:")
			report.Render(w)
		}()
	}

	if err := repository.ForEach(func(downloadUrl, filePath, name, version, sha256Digest string) error {
		defer bar.Increment()
		// doNexusMigrate(downloadUrl, filePath, name, version, sha256Digest, username, password string)
		if err1 := doNexusMigrate(downloadUrl, filePath, name, version, sha256Digest, username, password); err1 != nil {
			if err1 == ErrFileConflict {
				report.AddSkippedResult(strings.Join([]string{name, version}, "="), downloadUrl, "409 Conflict")
				return types.ErrForEachContinue
			}

			report.AddFailedResult(strings.Join([]string{name, version}, "="), downloadUrl, err1.Error())

			if settings.FailFast {
				return errors.Wrapf(err1, "failed to migrate %s", filePath)
			}
		} else {
			report.AddSucceededResult(strings.Join([]string{name, version}, "="), downloadUrl, "Succeeded")
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

func doNexusMigrate(downloadUrl, filePath, name, version, sha256Digest, username, password string) error {
	// download
	getResp, err := httputil.DefaultClient.GetWithAuth(downloadUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return errors.Wrapf(err, "failed to download from %s", downloadUrl)
	}
	defer ioutils.QuiteClose(getResp.Body)

	// post　MultipartForm contains  json and file
	// json key: "name", "version", "sha256_digest", "filetype": Egg, Wheel, Source
	pushUrl := getPushUrl(filePath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("content", filepath.Base(filePath))
	if err != nil {
		return errors.Wrapf(err, "failed to parse upload form file %s", filePath)
	}

	// write file
	_, err = io.Copy(part, getResp.Body)
	if err != nil {
		return errors.Wrapf(err, "failed to copy file stream to upload form %s", downloadUrl)
	}

	// cover pypi filetype from file extensions
	base := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(base))
	// for .tar.gz
	nonExt := base[:len(base)-len(ext)]
	if strings.ToLower(filepath.Ext(nonExt)) == ".tar" {
		ext = ".tar" + ext
	}
	fileType, has := DistExtensions[ext]
	if !has {
		return errors.Errorf("un support file extension, file: %s", path.Ext(filePath))
	}

	err = writer.WriteField("name", name)
	err = writer.WriteField("version", version)
	err = writer.WriteField("sha256_digest", sha256Digest)
	err = writer.WriteField("filetype", fileType)
	err = writer.Close()
	if err != nil {
		return errors.Wrapf(err, "failed to write json to upload form %s", downloadUrl)
	}

	resp, err := httputil.DefaultClient.Post(pushUrl, writer.FormDataContentType(), body, username, password)
	if err != nil {
		return errors.Wrapf(err, "failed to push to %s", pushUrl)
	}
	defer ioutils.QuiteClose(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		if resp.StatusCode == http.StatusConflict {
			return ErrFileConflict
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		return errors.Errorf("got an unexpected response status: %s, resp: %s", resp.Status, bodyString)
	}

	return nil
}

func getPushUrl(filePath string) string {
	// return strings.TrimSuffix(settings.Dst, "/") + "/"
	return settings.GetDstHasSubSlash()
}

const (
	WhlExt      = ".whl"
	ExeExt      = ".exe"
	EggExt      = ".egg"
	ZipExt      = ".zip"
	GzExt       = "gz"
	Gz2Ext      = "gz2"
	Bz2Ext      = ".bz2"
	DistInfoExt = ".dist-info"
	TarExt      = ".tar"
	GzTarExt    = ".tar.gz"
	BzTarExt    = ".tar.bz2"
	XzTarExt    = ".tar.xz"
	ZTarExt     = ".tar.Z" // maybe deprecated in the future version
)

const (
	BdistWheel   = "bdist_wheel"
	BdistWininst = "bdist_wininst"
	BdistEgg     = "bdist_egg"
	Sdist        = "sdist"
)

var DistExtensions = map[string]string{
	WhlExt:   BdistWheel,
	ExeExt:   BdistWininst, // deprecated since Python 3.8
	EggExt:   BdistEgg,
	BzTarExt: Sdist,
	GzTarExt: Sdist,
	ZipExt:   Sdist,
}

func isNeedMigrate(pkg, version string, exists map[string]bool) bool {
	return !(exists[fmt.Sprintf("%s:%s", pkg, version)])
}
