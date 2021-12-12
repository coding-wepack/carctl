package maven

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/config"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/log/logfields"
	"e.coding.net/codingcorp/carctl/pkg/settings"
	"e.coding.net/codingcorp/carctl/pkg/util/fileutil"
	"e.coding.net/codingcorp/carctl/pkg/util/httputil"
	"e.coding.net/codingcorp/carctl/pkg/util/ioutils"
)

var (
	ErrFileConflict = errors.New("failed to put file: 409 conflict")
)

func Migrate(cfg *action.Configuration, out io.Writer) error {
	if settings.Src == "" {
		settings.Src = defaultMavenRepositoryPath()
	}

	log.Info("Stat source repository ...")

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

	log.Info("Check authorization of the registry")
	configFile, err := cfg.RegistryClient.ConfigFile()
	if err != nil {
		return errors.Wrap(err, "failed to get config file")
	}

	has, authConfig, err := configFile.GetAuthConfig(settings.Dst)
	if err != nil {
		return errors.Wrap(err, "failed to get registry authorization info")
	}
	if !has {
		return errors.New("Unauthorized: authentication required. Maybe you haven't logged in before.")
	}

	if settings.Verbose {
		log.Debug("Auth config", logfields.String("host", authConfig.ServerAddress),
			logfields.String("username", authConfig.Username),
			logfields.String("password", authConfig.Password))
	}

	if err = migrateRepository(out, authConfig.Username, authConfig.Password); err != nil {
		return err
	}

	return nil
}

func migrateRepository(w io.Writer, username, password string) error {
	log.Info("Scanning repository ...")

	repository, err := GetRepository(settings.Src, settings.MaxFiles)
	if err != nil {
		return err
	}
	flattenRepository := repository.Flatten()
	log.Info("Successfully to scan the repository",
		logfields.Int("groups", flattenRepository.GetGroupCount()),
		logfields.Int("artifacts", flattenRepository.GetArtifactCount()),
		logfields.Int("versions", flattenRepository.GetVersionCount()),
		logfields.Int("files", flattenRepository.GetFileCount()))
	if flattenRepository.GetFileCount() == 0 {
		log.Warn("no files found, no need to migrate")
		return nil
	}
	if settings.Verbose {
		log.Info("Repository Info:")
		repository.Render(w)
	}

	log.Info("Begin to migrate ...")
	start := time.Now()

	var (
		succeededCount int
		failedCount    int
		skippedCount   int
	)
	for _, g := range repository.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for _, f := range v.Files {
					if err := doMigrate(f.Path, username, password); err != nil {
						if err == ErrFileConflict {
							skippedCount++
							continue
						}
						failedCount++
						if settings.FailFast {
							return errors.Wrapf(err, "failed to migrate %s", f.Path)
						} else {
							log.Warn("an error occurred during migration",
								logfields.String("file", f.Path),
								logfields.String("error", err.Error()))
						}
					} else {
						succeededCount++
						if settings.Verbose {
							log.Info("Successfully migrated:", logfields.String("file", f.Path))
						}
					}
				}
			}
		}
	}

	log.Info("End to migrate.",
		logfields.Duration("duration", time.Now().Sub(start)),
		logfields.Int("total", succeededCount+failedCount+skippedCount),
		logfields.Int("succeededCount", succeededCount),
		logfields.Int("failedCount", failedCount),
		logfields.Int("skippedCount", skippedCount))

	return nil
}

func doMigrate(file, username, password string) error {
	u := getPushUrl(file)
	log.Info("Put file:", logfields.String("file", file), logfields.String("url", u))
	resp, err := httputil.DefaultClient.PutFile(u, file, username, password)
	if err != nil {
		return err
	}
	defer ioutils.QuiteClose(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		if resp.StatusCode == http.StatusConflict {
			log.Warn("409 Conflict: file has been pushed, and the strategy of destination is not overridable, so just skip it",
				logfields.String("file", file))
			return ErrFileConflict
		}
		return errors.Errorf("got an unexpected response status: %s", resp.Status)
	}

	return nil
}

func GetRepository(repositoryPath string, maxFiles int) (repository *Repository, err error) {
	var fileCount int
	repository = &Repository{Path: repositoryPath}
	if err = filepath.WalkDir(repositoryPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if fileutil.IsFileInvisible(d.Name()) {
				return filepath.SkipDir
			}
			if !ArtifactNameRegex.MatchString(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if fileutil.IsFileInvisible(d.Name()) ||
			d.Name() == "_remote.repositories" ||
			strings.HasPrefix(d.Name(), "_") {
			return nil
		}
		if maxFiles >= 0 && fileCount >= maxFiles { // FIXME
			return filepath.SkipDir
		}

		groupName, artifact, version, filename, err := getArtInfo(path, repositoryPath)
		if err != nil {
			return errors.Wrap(err, "failed to get artifact info")
		}
		repository.AddVersionFile(groupName, artifact, version, filename, path)
		fileCount++

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to walk repository")
	}

	return
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
	if size < 3 {
		return "", "", "", "", errors.New("invalid path")
	}
	version = subPathChunks[size-2]
	artifact = subPathChunks[size-3]
	groupName = strings.Join(subPathChunks[:size-3], ".")
	return
}

func getPushUrl(filePath string) string {
	subPath := strings.Trim(strings.TrimPrefix(filePath, settings.Src), "/")
	return strings.TrimSuffix(settings.Dst, "/") + "/" + subPath
}

func defaultMavenRepositoryPath() string {
	return filepath.Join(config.GetHomeDir(), ".m2", "repository")
}
