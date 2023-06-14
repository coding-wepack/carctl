package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxySourceRegex(t *testing.T) {
	const (
		u1 = "http://wepack.cxy.dev.coding.io/api/user/wepack/project/registry/artifacts/repositories/3/proxies"
		u2 = "https://coding-artifacts.coding.net/api/user/coding-artifacts/project/registry/artifacts/repositories/3/proxies"
		u3 = "http://wepack.coding.io/api/user/wepack/project/public/artifacts/repositories/100/proxies/subpath"
	)

	assert.True(t, proxySourceUrlRegex.MatchString(u1))
	assert.True(t, proxySourceUrlRegex.MatchString(u2))
	assert.False(t, proxySourceUrlRegex.MatchString(u3))
}

func TestValidateProxySourceUrl(t *testing.T) {
	const u1 = "http://wepack.cxy.dev.coding.io/api/user/wepack/project/registry/artifacts/repositories/3/proxies"
	expectedParam1 := ProxySourceUrlRegexParam{
		Url:           u1,
		HttpSchema:    "http",
		HttpDomain:    "cxy.dev.coding.io",
		TeamGlobalKey: "wepack",
		Project:       "registry",
		RepoId:        3,
	}
	p1, e1 := ValidateProxySourceUrl(u1)
	require.NoError(t, e1)
	require.NotNil(t, p1)
	assert.Equal(t, expectedParam1, *p1)

	const u2 = "https://coding-artifacts.coding.net/api/user/coding-artifacts/project/registry/artifacts/repositories/1/proxies"
	expectedParam2 := ProxySourceUrlRegexParam{
		Url:           u2,
		HttpSchema:    "https",
		HttpDomain:    "coding.net",
		TeamGlobalKey: "coding-artifacts",
		Project:       "registry",
		RepoId:        1,
	}
	p2, e2 := ValidateProxySourceUrl(u2)
	require.NoError(t, e2)
	require.NotNil(t, p2)
	assert.Equal(t, expectedParam2, *p2)

	const u3 = "http://wepack.coding.io/api/user/wepack/project/public/artifacts/repositories/100/proxies/subpath"
	p3, e3 := ValidateProxySourceUrl(u3)
	require.EqualError(t, e3, "invalid proxy source url")
	require.Nil(t, p3)

	const u4 = "https://wepack-test.coding.net/api/user/wepack/project/demo/artifacts/repositories/7/proxies"
	p4, e4 := ValidateProxySourceUrl(u4)
	require.ErrorContains(t, e4, "bad request: teamGlobalKey doesn't match: ")
	require.Nil(t, p4)

	const u5 = "http://wepack.coding.com/api/user/wepack/project/test/artifacts/repositories/-9/proxies"
	p5, e5 := ValidateProxySourceUrl(u5)
	require.ErrorContains(t, e5, "invalid proxy source url")
	require.Nil(t, p5)
}
