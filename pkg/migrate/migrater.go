package migrate

import (
	"e.coding.net/codingcorp/camigrater/pkg/migrate/maven"
	"github.com/pkg/errors"

	"e.coding.net/codingcorp/camigrater/pkg/constants"
	"e.coding.net/codingcorp/camigrater/pkg/flags"
)

func Migrate() error {
	switch flags.Type {
	case constants.TypeMaven:
		return maven.Maven()
	default:
		return errors.Errorf("Unsupported artifact type yet: %s", flags.Type)
	}
}
