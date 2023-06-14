package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/auth"
	jfconfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/pkg/errors"
)

type JfrogFile struct {
	Repo       string    `json:"repo"`
	Path       string    `json:"path"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Size       int       `json:"size"`
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"created_by"`
	Modified   string    `json:"modified"`
	ModifiedBy string    `json:"modified_by"`
	Updated    time.Time `json:"updated"`
}

func (f *JfrogFile) GetFilePath() string {
	if f.Path == "" || f.Path == "." {
		return f.Name
	}
	return fmt.Sprintf("%s/%s", f.Path, f.Name)
}

func (f *JfrogFile) GetDockerInfo() (srcPath, pkg, version string, err error) {
	split := strings.Split(strings.Trim(f.Path, "/"), "/")
	if len(split) < 2 {
		err = errors.Errorf("the srcPath dir level cannot be less than 2, srcPath: %s", f.Path)
		return
	}
	pkg = strings.Join(split[:len(split)-1], "_")
	version = split[len(split)-1]
	srcPath = fmt.Sprintf("%s:%s", strings.Join(split[:len(split)-1], "/"), version)
	return
}

type JfrogFileResRange struct {
	StartPos int `json:"start_pos"`
	EndPos   int `json:"end_pos"`
	Total    int `json:"total"`
}

type JfrogFileResult struct {
	Res []JfrogFile       `json:"results"`
	Ran JfrogFileResRange `json:"range"`
}

var jfrogAsManager artifactory.ArtifactoryServicesManager

func initJfrogArtifactsManager(jfrogUrl *url.URL) (err error) {
	rtDetails := auth.NewArtifactoryDetails()
	rtDetails.SetUrl(fmt.Sprintf("%s://%s/artifactory/", jfrogUrl.Scheme, jfrogUrl.Host))
	rtDetails.SetUser(settings.SrcUsername)
	rtDetails.SetPassword(settings.SrcPassword)

	serviceConfig, err := jfconfig.NewConfigBuilder().
		SetServiceDetails(rtDetails).
		// Optionally overwrite the default HTTP timeout, which is set to 30 seconds.
		SetHttpTimeout(30 * time.Second).
		// Optionally overwrite the default HTTP retries, which is set to 3.
		SetHttpRetries(3).
		Build()
	if err != nil {
		return errors.Wrap(err, "failed to build jfrog service config")
	}
	jfrogAsManager, err = artifactory.New(serviceConfig)
	if err != nil {
		return errors.Wrap(err, "failed to build jfrog service manager")
	}
	return nil
}

// FindFileListFromJfrog 使用 jfrog AQL 来获取文件列表
func FindFileListFromJfrog(jfrogUrl *url.URL, repository string) (filesInfo *JfrogFileResult, err error) {
	if jfrogAsManager == nil {
		err = initJfrogArtifactsManager(jfrogUrl)
		if err != nil {
			err = errors.Wrap(err, "failed to init jfrog artifacts manager")
			return nil, err
		}
	}

	// 执行 AQL
	reader, err := jfrogAsManager.Aql(fmt.Sprintf("items.find({\"repo\": \"%s\"})", repository))
	if err != nil {
		return nil, errors.Wrap(err, "executed jfrog AQL query failed")
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "read content from jfrog aql result failed")
	}

	filesInfo = new(JfrogFileResult)
	if err = json.Unmarshal(content, filesInfo); err != nil {
		return nil, errors.Wrap(err, "unmarshal jfrog AQL query result failed")
	}
	return
}
