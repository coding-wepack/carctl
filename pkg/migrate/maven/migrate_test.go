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
