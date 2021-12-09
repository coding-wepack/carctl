package maven

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"e.coding.net/codingcorp/carctl/pkg/config"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/log/logfields"
	"e.coding.net/codingcorp/carctl/pkg/settings"
	"e.coding.net/codingcorp/carctl/pkg/util/fileutil"
)

func Migrate() error {
	if settings.Src == "" {
		settings.Src = defaultMavenRepositoryPath()
	}

	log.Info("stat source repository ...")

	repositoryFileInfo, err := os.Stat(settings.Src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn("source repository not found", logfields.String("path", settings.Src))
			return nil
		}
		return err
	}
	if !repositoryFileInfo.IsDir() {
		return errors.New("source repository is not a directory")
	}

	log.Info("count packages under")

	// TODO: max packages support
	packages, err := fileutil.ListVisibleDirNamesWithSort(settings.Src, -1)
	if err != nil {
		return err
	}
	if len(packages) == 0 {
		log.Info("no packages found in source repository")
		return nil
	}

	log.Info("found packages in source repository", logfields.Int("count", len(packages)),
		logfields.Strings("packages", packages))

	return nil
}

func getVersions(repositoryPath string) (versions []Version, err error) {
	if err = filepath.WalkDir(repositoryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Warn("got an error while walking dir", logfields.String("path", path),
				logfields.String("error", err.Error()))
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") || d.Name() == "_remote.repositories" {
			return nil
		}

		return nil
	}); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}

	return
}

func defaultMavenRepositoryPath() string {
	return filepath.Join(config.GetHomeDir(), ".m2", "repository")
}
