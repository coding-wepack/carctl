package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"e.coding.net/codingcorp/carctl/pkg/config"
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
func FindDstRepoArtifacts(cfg *config.AuthConfig, dst, artifactType string) ([]*Artifacts, error) {
	// 解析目标 URL，获取域名以及项目名、仓库名
	dstUrl, err := url.Parse(dst)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse dst url %s", settings.GetDstWithoutSlash())
	}
	host := replaceHost(dstUrl.Host, artifactType)
	split := strings.Split(strings.Trim(dstUrl.Path, "/"), "/")
	if len(split) != 2 {
		return nil, errors.New("dst url path format must match /{project}/{repository}")
	}

	// 构建 open api 请求，查询仓库下的制品版本信息
	openApiUrl := fmt.Sprintf("%s://%s/open-api", dstUrl.Scheme, host)
	req := &DescribeTeamArtifactsReq{
		Action:     "DescribeTeamArtifacts",
		PageNumber: 1,
		PageSize:   999999,
		Rule: &DescribeTeamArtifactsRule{
			ProjectName: split[:1],
			Repository:  split[1:],
		},
	}
	marshal, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal describe team artifacts reqeust")
	}
	if settings.Verbose {
		log.Debugf("find exists artifacts, url: %s, username: %s, password: %s, reqBody: %s", openApiUrl, cfg.Username, cfg.Password, string(marshal))
	}

	resp, err := httputil.DefaultClient.PostJson(openApiUrl, bytes.NewReader(marshal), cfg.Username, cfg.Password)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to describe team artifacts")
	}
	defer ioutils.QuiteClose(resp.Body)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read resp body")
	}
	result := &DescribeTeamArtifactsResp{}
	err = json.Unmarshal(bodyBytes, result)
	if err != nil {
		if settings.Verbose {
			log.Debugf("unmarshal response body failed, body: %s", string(bodyBytes))
		}
		return nil, errors.Wrapf(err, "failed to unmarshal resp body")
	}
	if result.Response.Error != nil {
		return nil, errors.Errorf("failed to find exists artifacts: %s", result.Response.Error.Code)
	}
	return result.Response.Data.InstanceSet, nil
}

func replaceHost(regHost, artifactType string) (host string) {
	// regHost 的两种形式，1. {gk}-{artifactType}.pkg.{domain}; 2. {gk}-{artifactType}.{domain}
	host = strings.ReplaceAll(regHost, ".pkg.", ".")
	return strings.ReplaceAll(host, fmt.Sprintf("-%s", artifactType), "")
}
