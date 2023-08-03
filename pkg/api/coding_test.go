package api

import (
	"testing"

	"github.com/coding-wepack/carctl/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestAddArtiProperty(t *testing.T) {
	dst := "http://codingcorp-docker.pkg.yuslin.devnf.codingcorp.net/lys/docker"
	artifactType := "docker"
	pkg := "alpine"
	version := "latest"
	propName := "name2"
	propValue := "value2"
	cfg := &config.AuthConfig{
		Username: "docker-1689150199311",
		Password: "fd88921bec09c1395986f4780985e28e59a54629",
	}
	err := AddProperties(cfg, dst, artifactType, pkg, version, propName, propValue)
	require.NoError(t, err)
}
