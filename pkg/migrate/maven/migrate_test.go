package maven

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"e.coding.net/codingcorp/carctl/pkg/settings"
	"e.coding.net/codingcorp/carctl/pkg/util/fileutil"
	"github.com/stretchr/testify/assert"
)

func TestMigrate(t *testing.T) {
	settings.Verbose = true
	err := Migrate()
	assert.NoError(t, err)
}

func TestWalkDir(t *testing.T) {
	filepath.WalkDir("/Users/chenxinyu/.m2/repository", func(path string, d fs.DirEntry, err error) error {
		fmt.Printf("path: %s\n", path)
		// fmt.Printf("parent path: %s\n", filepath.Dir(path))
		fmt.Printf("Name: %s, IsDir: %t\n", d.Name(), d.IsDir())
		if err != nil {
			fmt.Printf("[ERROR] error: %v\n", err)
		}
		fmt.Println("==============================")

		return nil
	})
}

func TestWalkDir2(t *testing.T) {
	const repositoryPath = "/Users/chenxinyu/.m2/repository"
	const n = 100
	var count int

	repository := Repository{Path: repositoryPath}
	err := filepath.WalkDir(repositoryPath, func(path string, d fs.DirEntry, err error) error {
		// /Users/chenxinyu/.m2/repository/org/json/json/20171018/json-20171018.jar
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
		if count >= n {
			return nil
		}

		groupName, artifact, version, filename, err := getArtInfo(path, repositoryPath)
		assert.NoError(t, err)
		fmt.Printf("Path: %s\n", path)
		fmt.Printf("Group: [%s], Artifact: [%s], Version: [%s], Filename: [%s]\n",
			groupName, artifact, version, filename)
		fmt.Println("================================================================================================")
		count++

		repository

		return nil
	})
	assert.NoError(t, err)

	fmt.Printf("count: %v\n", count)
	fmt.Printf("%+v\n", repository)
}
