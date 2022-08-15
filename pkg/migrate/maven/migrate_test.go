package maven

import (
	"os"
	"testing"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/registry"
	"e.coding.net/codingcorp/carctl/pkg/settings"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	settings.Verbose = true
	settings.MaxFiles = -1
	settings.Src = "/home/juan/.m2/swagger-core-repository"
	settings.Dst = "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/central"

	log.SetDebug()

	regCli, err := registry.NewClient()
	assert.NoError(t, err)

	cfg := &action.Configuration{RegistryClient: regCli}

	err = Migrate(cfg, os.Stdout)
	assert.NoError(t, err)
}

func TestMigrateWin(t *testing.T) {
	settings.Verbose = true
	settings.MaxFiles = -1
	settings.Src = `E:\\mvnrepo\\log4j\\`
	settings.Dst = "http://codingcorp-maven.pkg.nh4ivfk.dev.coding.io/repository/registry/central"

	log.SetDebug()

	isLocalPath := isLocalRepository(settings.Src)
	assert.True(t, isLocalPath)
}

func TestMigrateFromNexus(t *testing.T) {
	settings.Verbose = true
	settings.MaxFiles = -1
	settings.Src = "http://localhost:8081/repository/maven-test/"
	settings.SrcUsername = "admin"
	settings.SrcPassword = "coding123"
	settings.Dst = "http://wepack-maven.pkg.coding.9.134.115.120.nip.io/repository/primary/maven-test"

	log.SetDebug()

	regCli, err := registry.NewClient()
	assert.NoError(t, err)

	cfg := &action.Configuration{RegistryClient: regCli}

	err = Migrate(cfg, os.Stdout)
	assert.NoError(t, err)
}

func TestGetArtInfoFromSubPath(t *testing.T) {
	const (
		subPathSNAPSHOT            = "test/example/test.example-test2/1.0.6-SNAPSHOT/test.example-test2-1.0.6-20211217.073105-2.jar"
		subPathSNAPSHOTMd5         = "test/example/test.example-test2/1.0.6-SNAPSHOT/test.example-test2-1.0.6-20211217.073105-2.jar.md5"
		subPathSNAPSHOTMetadata    = "test/example/test.example-test2/1.0.6-SNAPSHOT/maven-metadata.xml"
		subPathSNAPSHOTMetadataMd5 = "test/example/test.example-test2/1.0.6-SNAPSHOT/maven-metadata.xml.md5"
		subPathMetadata            = "test/example/test.example-test2/maven-metadata.xml"
		subPathMetadataMd5         = "test/example/test.example-test2/maven-metadata.xml.md5"
	)
	{
		group, artifact, version, filename, _ := getArtInfoFromSubPath(subPathSNAPSHOT)
		assert.Equal(t, "test.example", group)
		assert.Equal(t, "test.example-test2", artifact)
		assert.Equal(t, "1.0.6-SNAPSHOT", version)
		assert.Equal(t, "test.example-test2-1.0.6-20211217.073105-2.jar", filename)
	}
	{
		group, artifact, version, filename, _ := getArtInfoFromSubPath(subPathSNAPSHOTMetadata)
		assert.Equal(t, "test.example", group)
		assert.Equal(t, "test.example-test2", artifact)
		assert.Equal(t, "1.0.6-SNAPSHOT", version)
		assert.Equal(t, "maven-metadata.xml", filename)
	}
	{
		group, artifact, version, filename, _ := getArtInfoFromSubPath(subPathMetadata)
		assert.Equal(t, "test.example", group)
		assert.Equal(t, "test.example-test2", artifact)
		assert.Equal(t, Metadata, version)
		assert.Equal(t, "maven-metadata.xml", filename)
	}

	{
		group, artifact, version, filename, _ := getArtInfoFromSubPath(subPathSNAPSHOTMd5)
		assert.Equal(t, "test.example", group)
		assert.Equal(t, "test.example-test2", artifact)
		assert.Equal(t, "1.0.6-SNAPSHOT", version)
		assert.Equal(t, "test.example-test2-1.0.6-20211217.073105-2.jar.md5", filename)
	}
	{
		group, artifact, version, filename, _ := getArtInfoFromSubPath(subPathSNAPSHOTMetadataMd5)
		assert.Equal(t, "test.example", group)
		assert.Equal(t, "test.example-test2", artifact)
		assert.Equal(t, "1.0.6-SNAPSHOT", version)
		assert.Equal(t, "maven-metadata.xml.md5", filename)
	}
	{
		group, artifact, version, filename, _ := getArtInfoFromSubPath(subPathMetadataMd5)
		assert.Equal(t, "test.example", group)
		assert.Equal(t, "test.example-test2", artifact)
		assert.Equal(t, Metadata, version)
		assert.Equal(t, "maven-metadata.xml.md5", filename)
	}
}
