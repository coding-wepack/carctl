package maven

import (
	"os"
	"path/filepath"

	"e.coding.net/codingcorp/carctl/pkg/log/logfields"
	"github.com/pkg/errors"

	"e.coding.net/codingcorp/carctl/pkg/config"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/settings"
)

func Migrate() error {
	if settings.Src == "" {
		settings.Src = defaultMavenRepositoryPath()
	}

	repositoryDir, err := os.Open(settings.Src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("repository not found")
			return nil
		}
		return err
	}

	// TODO: max packages support
	groupNames, err := repositoryDir.Readdirnames(-1)
	if err != nil {
		return err
	}
	if len(groupNames) == 0 {
		log.Info("no packages found in repository")
		return nil
	}

	if settings.Verbose {
		log.Info("found packages in repository", logfields.Int("count", len(groupNames)),
			logfields.Strings("groupNames", groupNames))
	}

	return nil
}

func defaultMavenRepositoryPath() string {
	return filepath.Join(config.GetHomeDir(), ".m2", "repository")
}
