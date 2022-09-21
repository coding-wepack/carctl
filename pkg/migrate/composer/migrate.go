package composer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/log/logfields"
	"e.coding.net/codingcorp/carctl/pkg/migrate/composer/types"
	"e.coding.net/codingcorp/carctl/pkg/migrate/composer/types/nexus"
	reportutil "e.coding.net/codingcorp/carctl/pkg/report"
	"e.coding.net/codingcorp/carctl/pkg/settings"
	"e.coding.net/codingcorp/carctl/pkg/util/httputil"
	"e.coding.net/codingcorp/carctl/pkg/util/ioutils"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var (
	ErrFileConflict = errors.New("failed to put file: 409 conflict")
)

func Migrate(cfg *action.Configuration, out io.Writer) error {
	srcUrl, err := url.Parse(settings.Src)
	if err != nil || srcUrl.Scheme == "" {
		log.Info("only support migrate from nexus...", logfields.String("src", settings.Src))
		if settings.Verbose {
			log.Warn("Can't parse with error", logfields.Error(err))
		}
		return errors.New("source repository is not a directory")
	} else {
		return MigrateFromUrl(cfg, out, srcUrl)
	}
}

func MigrateFromUrl(cfg *action.Configuration, out io.Writer, srcUrl *url.URL) error {
	// 默认为 nexus
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

func GetRepositoryFromNexusItems(repositoryUrl string, nexusItemList []nexus.Item) (repository *types.Repository, err error) {
	repository = &types.Repository{Path: repositoryUrl}
	for _, item := range nexusItemList {
		if strings.HasSuffix(item.Path, "json") {
			composerList, err := getComposerList(item.DownloadURL)
			if err != nil {
				return repository, errors.Wrap(err, "failed to get composer list")
			}
			repository.AddVersionFileList(composerList)
		}
	}
	return
}

func migrateNexusRepository(w io.Writer, nexusItemList []nexus.Item, username, password string) error {
	log.Info("Scanning nexus repository ...")

	// filter and parse composer list
	repository, err := GetRepositoryFromNexusItems(settings.Src, nexusItemList)
	if err != nil {
		return err
	}

	if settings.Verbose {
		log.Info("Repository Info:")
		repository.Render(w)
	}

	// Progress Bar
	// initialize progress container, with custom width
	p := mpb.New(mpb.WithWidth(80))
	total := len(repository.Files)
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

	if err := repository.ForEach(func(path, version, downloadUrl string) error {
		defer bar.Increment()
		if err1 := doNexusMigrate(path, version, downloadUrl, username, password); err1 != nil {
			if err1 == ErrFileConflict {
				report.AddSkippedResult(path, downloadUrl, "409 Conflict")
				return types.ErrForEachContinue
			}

			report.AddFailedResult(path, downloadUrl, err1.Error())

			if settings.FailFast {
				return errors.Wrapf(err1, "failed to migrate %s", path)
			}
		} else {
			report.AddSucceededResult(path, downloadUrl, "Succeeded")
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

func doNexusMigrate(path, version, downloadUrl, username, password string) error {
	// download
	getResp, err := httputil.DefaultClient.GetWithAuth(downloadUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return errors.Wrapf(err, "failed to download from %s", downloadUrl)
	}
	defer ioutils.QuiteClose(getResp.Body)

	// push
	pushUrl := getPushUrl(version)
	resp, err := httputil.DefaultClient.Put(pushUrl, "", getResp.Body, username, password)
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

func getPushUrl(version string) string {
	return strings.TrimSuffix(settings.Dst, "/") + "?version=" + version
}

// getFileListFromNexus 　getFileListFromNexus, result contains zip and project.json
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

// Parse Package Info
func getComposerList(downloadUrl string) (composerList []*nexus.ComposerItem, err error) {

	resp, err := httputil.DefaultClient.GetWithAuth(downloadUrl, settings.SrcUsername, settings.SrcPassword)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get components: %s", downloadUrl)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to get components: %s, status: %s", downloadUrl, resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read resp: %s", downloadUrl)
	}

	var getComponentsResp *nexus.Packages
	err = json.Unmarshal(bodyBytes, &getComponentsResp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal resp: %s", string(bodyBytes))
	}

	for _, p := range getComponentsResp.Packages {
		for _, vInfo := range p {
			composerList = append(composerList, vInfo)
		}
	}

	return composerList, nil
}
