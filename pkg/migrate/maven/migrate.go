package maven

import (
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

func getArtInfo(path, repositoryPath string) (groupName, artifact, version, filename string, err error) {
	// repositoryPath: /Users/chenxinyu/.m2/repository
	// path: /Users/chenxinyu/.m2/repository/org/kohsuke/stapler/json-lib/2.4-jenkins-2/json-lib-2.4-jenkins-2-sources.jar
	// subPath: org/kohsuke/stapler/json-lib/2.4-jenkins-2/json-lib-2.4-jenkins-2-sources.jar
	// filename: json-lib-2.4-jenkins-2-sources.jar
	subPath := strings.Trim(strings.TrimPrefix(path, repositoryPath), "/")
	filename = filepath.Base(path)

	subPathChunks := strings.Split(subPath, "/")
	size := len(subPathChunks)
	if size < 4 {
		return "", "", "", "", errors.New("invalid path")
	}
	version = subPathChunks[size-2]
	artifact = subPathChunks[size-3]
	groupName = strings.Join(subPathChunks[:size-3], ".")
	return
}

func defaultMavenRepositoryPath() string {
	return filepath.Join(config.GetHomeDir(), ".m2", "repository")
}
