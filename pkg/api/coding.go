package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/coding-wepack/carctl/pkg/config"
	"github.com/coding-wepack/carctl/pkg/constants"
	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/httputil"
	"github.com/coding-wepack/carctl/pkg/util/ioutils"
	"github.com/pkg/errors"
)

func FindDstExistsArtifacts(cfg *config.AuthConfig, dst, artifactType string) (result map[string]bool, err error) {
	// 解析目标 URL，获取域名以及项目名、仓库名
	scheme, host, project, repo, err := parseDst(dst, artifactType)
	if err != nil {
		return
	}
	repo = strings.ToLower(repo)
	// 拼接 open api 域名，构建请求体
	openApiUrl := fmt.Sprintf("%s://%s/open-api", scheme, host)
	pageNumber, pageSize := 1, 1000
	// 构建 open api 请求，查询仓库下的制品版本信息
	req := &DescribeTeamArtifactsReq{
		Action:   "DescribeTeamArtifacts",
		PageSize: pageSize,
		Rule: &DescribeTeamArtifactsRule{
			ProjectName: []string{project},
			Repository:  []string{repo},
		},
	}
	if settings.Verbose {
		log.Debugf("find exists artifacts, url: %s, username: %s, password: %s, project: %s, repository: %s",
			openApiUrl, cfg.Username, cfg.Password, project, repo)
	}

	result = make(map[string]bool)
	resp := &DescribeTeamArtifactsResp{}
	for {
		// 发起分页请求
		req.PageNumber = pageNumber
		err = execute(cfg, openApiUrl, req, resp)
		if err != nil {
			return nil, err
		}
		respRsl := resp.Response
		if respRsl.Error != nil {
			err = errors.Errorf("failed to find exists artifacts: %s", respRsl.Error.Code)
			return
		}
		if settings.Verbose {
			log.Debugf("find exists artifacts. pageNumber:%d, pageSize:%d, totalCount:%d", pageNumber, pageSize, respRsl.Data.TotalCount)
		}
		for _, instance := range respRsl.Data.InstanceSet {
			result[fmt.Sprintf("%s:%s", instance.Package, instance.PackageVersion)] = true
		}
		if pageNumber*pageSize > respRsl.Data.TotalCount {
			break
		}
		pageNumber++
	}
	return
}

// FindDstExistsFiles 查询目标仓库已存在的制品文件
func FindDstExistsFiles(cfg *config.AuthConfig, dst, artifactType string) (data map[string]bool, err error) {
	// 解析目标 URL，获取域名以及项目名、仓库名
	scheme, host, project, repo, err := parseDst(dst, artifactType)
	if err != nil {
		return
	}

	// 拼接 open api 域名，构建请求体
	openApiUrl := fmt.Sprintf("%s://%s/open-api", scheme, host)
	pageSize := 1000
	// 构建 open api 请求，查询仓库下的制品版本信息
	req := &DescribeRepoFileListReq{
		Action:     "DescribeArtifactRepositoryFileList",
		PageSize:   pageSize,
		Project:    project,
		Repository: repo,
	}
	if settings.Verbose {
		log.Debugf("find exists files, url: %s, username: %s, password: %s, project: %s, repository: %s",
			openApiUrl, cfg.Username, cfg.Password, project, repo)
	}

	data = make(map[string]bool)
	resp := &DescribeRepoFileListResp{}
	continuationToken := ""
	for {
		// 发起分页请求
		req.ContinuationToken = continuationToken
		err = execute(cfg, openApiUrl, req, resp)
		if err != nil {
			return nil, err
		}
		respRsl := resp.Response
		if respRsl.Error != nil {
			err = errors.Errorf("failed to find exists files: %s", respRsl.Error.Code)
			return
		}
		respData := respRsl.Data
		if settings.Verbose {
			log.Debugf("find exists files with pageSize:%d, resultSize:%d, continuationToken:%s",
				pageSize, len(respData.InstanceSet), respData.ContinuationToken)
		}
		if len(respData.InstanceSet) == 0 {
			break
		}
		for _, f := range respData.InstanceSet {
			data[f.Path] = true
		}
		if len(respData.ContinuationToken) == 0 {
			break
		}
		continuationToken = respData.ContinuationToken
	}
	return data, nil
}

func execute[T any, R any](cfg *config.AuthConfig, url string, req T, resp R) (err error) {
	marshal, err := json.Marshal(req)
	if err != nil {
		err = errors.Wrapf(err, "failed to marshal describe team artifacts reqeust")
		return
	}

	openApiResp, err := httputil.DefaultClient.PostJson(url, bytes.NewReader(marshal), cfg.Username, cfg.Password)
	if err != nil {
		err = errors.Wrapf(err, "failed to describe team artifacts")
		return
	}
	defer ioutils.QuiteClose(openApiResp.Body)
	bodyBytes, err := io.ReadAll(openApiResp.Body)
	if err != nil {
		err = errors.Wrapf(err, "failed to read resp body")
		return
	}
	err = json.Unmarshal(bodyBytes, resp)
	if err != nil {
		if settings.Verbose {
			log.Debugf("unmarshal response body failed, body: %s", string(bodyBytes))
		}
		err = errors.Wrapf(err, "failed to unmarshal resp body")
	}
	return
}

func parseDst(dst, artifactType string) (scheme, host, project, repo string, err error) {
	dstUrl, err := url.Parse(dst)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse dst url %s", settings.GetDstWithoutSlash())
		return
	}
	scheme = dstUrl.Scheme
	host = replaceHost(dstUrl.Host, artifactType)

	split := strings.Split(strings.Trim(dstUrl.Path, "/"), "/")
	if strings.EqualFold(constants.TypeMaven, artifactType) {
		if len(split) != 3 {
			err = errors.New("dst url path format must match /repository/{project}/{repository}")
			return
		}
		split = split[1:]
	} else if len(split) != 2 {
		err = errors.New("dst url path format must match /{project}/{repository}")
		return
	}
	return dstUrl.Scheme, replaceHost(dstUrl.Host, artifactType), split[0], split[1], nil
}

func replaceHost(regHost, artifactType string) (host string) {
	// regHost 的两种形式，1. {gk}-{artifactType}.pkg.{domain}; 2. {gk}-{artifactType}.{domain}
	host = strings.ReplaceAll(regHost, ".pkg.", ".")
	return strings.ReplaceAll(host, fmt.Sprintf("-%s", artifactType), "")
}
