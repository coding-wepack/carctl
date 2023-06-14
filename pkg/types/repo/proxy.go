package repo

import (
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

var (
	// proxySourceUrlRegexStr = `^(?P<schema>https?)://(?P<team_gk>[a-zA-Z0-9_-]+)\.(?P<domain>[^/]+)/api/user/(?P=team_gk)/project/(?P<project>[a-zA-Z0-9_-]+)/artifacts/repositories/(?P<repo_id>\d+)/proxies$`
	// Go 的 regexp 包不支持命名捕获组的反向引用，因此设置后面的 team_gk 为 team_gk1，之后和前面的 team_gk 手动比较是否一致
	proxySourceUrlRegexStr = `^(?P<schema>https?)://(?P<team_gk>[a-zA-Z0-9_-]+)\.(?P<domain>[^/]+)/api/user/(?P<team_gk1>[a-zA-Z0-9_-]+)/project/(?P<project>[a-zA-Z0-9_-]+)/artifacts/repositories/(?P<repo_id>\d+)/proxies$`
	proxySourceUrlRegex    = regexp.MustCompile(proxySourceUrlRegexStr)
)

const (
	// 代理源 URL 正则表达式匹配到的数组长度，为组数 +1，第一个值为原始的完整的 URL
	proxySourceUrlRegexMatchSize = 7
)

type ProxySourcePayload struct {
	Name     string `json:"name"`
	Source   string `json:"source"`
	Priority int    `json:"priority"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type ProxySourceResponse struct {
	Code    int               `json:"code"`
	Data    []ProxySource     `json:"data,omitempty"`
	Message map[string]string `json:"msg,omitempty"`
}

type PostProxySourceResponse struct {
	Code int `json:"code"`
	// Message type may be string or map[string]string
	Message any `json:"msg"`
}

type ProxySource struct {
	Id             int64  `json:"id"`
	RepoId         int64  `json:"repoId"`
	Name           string `json:"name"`
	Source         string `json:"source"`
	Priority       int    `json:"priority"`
	Username       string `json:"username"`
	CredentialType int    `json:"credentialType"`
	CredentialRef  int    `json:"credentialRef"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
}

type ProxySourceUrlRegexParam struct {
	Url           string
	HttpSchema    string
	HttpDomain    string
	TeamGlobalKey string
	Project       string
	RepoId        int64
}

func ValidateProxySourceUrl(proxySourceUrl string) (param *ProxySourceUrlRegexParam, err error) {
	matches := proxySourceUrlRegex.FindStringSubmatch(proxySourceUrl)
	if len(matches) != proxySourceUrlRegexMatchSize {
		return nil, errors.New("invalid proxy source url")
	}

	param = &ProxySourceUrlRegexParam{
		Url: proxySourceUrl,
	}
	var teamGlobalKey1 string
	for i, name := range proxySourceUrlRegex.SubexpNames() {
		switch name {
		case "schema":
			param.HttpSchema = matches[i]
		case "domain":
			param.HttpDomain = matches[i]
		case "team_gk":
			param.TeamGlobalKey = matches[i]
		case "team_gk1":
			teamGlobalKey1 = matches[i]
		case "project":
			param.Project = matches[i]
		case "repo_id":
			repoIdStr := matches[i]
			param.RepoId, err = strconv.ParseInt(repoIdStr, 10, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse repoId to an integer: %q", repoIdStr)
			}
		}
	}

	if teamGlobalKey1 != param.TeamGlobalKey {
		return nil, errors.Errorf("bad request: teamGlobalKey doesn't match: %s != %s",
			param.TeamGlobalKey, teamGlobalKey1)
	}

	return param, nil
}
