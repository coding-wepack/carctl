package maven

import (
	"testing"

	"e.coding.net/codingcorp/carctl/pkg/settings"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	settings.Verbose = true
	err := Migrate()
	assert.NoError(t, err)
}
