package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"e.coding.net/codingcorp/carctl/pkg/config"
	"e.coding.net/codingcorp/carctl/pkg/constants"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/settings"
	"e.coding.net/codingcorp/carctl/pkg/util/httputil"
	"e.coding.net/codingcorp/carctl/pkg/util/ioutils"
	"github.com/pkg/errors"
)

func FindDstRepoArtifactsName(cfg *config.AuthConfig, dst, artifactType string) (map[string]bool, error) {
	artifacts, err := FindDstRepoArtifacts(cfg, dst, artifactType)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(artifacts))
	for _, a := range artifacts {
		result[fmt.Sprintf("%s:%s", a.Package, a.PackageVersion)] = true
	}
	return result, nil
}

// FindDstRepoArtifacts 查询目标仓库已存在的制品信息
func FindDstRepoArtifacts(cfg *config.AuthConfig, dst, artifactType string) (data []*Artifacts, err error) {
	// 解析目标 URL，获取域名以及项目名、仓库名
	scheme, host, project, repo, err := parseDst(dst, artifactType)
	if err != nil {
		return
	}

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

	for {
		// 发起分页请求
		req.PageNumber = pageNumber
		count, set, err := doFindWithPage(cfg, openApiUrl, req)
		if err != nil {
			return nil, err
		}
		if settings.Verbose {
			log.Debugf("find with pageNumber:%d, pageSize:%d, totalCount:%d", pageNumber, pageSize, count)
		}
		data = append(data, set...)
		if pageNumber*pageSize > count {
			break
		}
		pageNumber++
	}
	return data, nil
}

func doFindWithPage(cfg *config.AuthConfig, url string, req *DescribeTeamArtifactsReq) (totalCount int, instanceSet []*Artifacts, err error) {
	marshal, err := json.Marshal(req)
	if err != nil {
		err = errors.Wrapf(err, "failed to marshal describe team artifacts reqeust")
		return
	}

	resp, err := httputil.DefaultClient.PostJson(url, bytes.NewReader(marshal), cfg.Username, cfg.Password)
	if err != nil {
		err = errors.Wrapf(err, "failed to describe team artifacts")
		return
	}
	defer ioutils.QuiteClose(resp.Body)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrapf(err, "failed to read resp body")
		return
	}
	result := &DescribeTeamArtifactsResp{}
	err = json.Unmarshal(bodyBytes, result)
	if err != nil {
		if settings.Verbose {
			log.Debugf("unmarshal response body failed, body: %s", string(bodyBytes))
		}
		err = errors.Wrapf(err, "failed to unmarshal resp body")
		return
	}
	respRsl := result.Response
	if respRsl.Error != nil {
		err = errors.Errorf("failed to find exists artifacts: %s", respRsl.Error.Code)
		return
	}
	return respRsl.Data.TotalCount, respRsl.Data.InstanceSet, nil
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
