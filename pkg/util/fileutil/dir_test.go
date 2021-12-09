package fileutil

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListVisibleDirNames(t *testing.T) {
	const path = "/home/juan/.m2/test-repository"
	dirs, err := ListDirNames(path, -1)
	assert.NoError(t, err)
	fmt.Println(dirs)

	dirs, err = ListVisibleDirNames(path, -1)
	assert.NoError(t, err)
	fmt.Println(dirs)

	dirs, err = ListVisibleDirNamesWithSort(path, -1)
	assert.NoError(t, err)
	fmt.Println(dirs)
}

func TestWalkDir(t *testing.T) {
	filepath.WalkDir("/home/juan/.m2/test-repository", func(path string, d fs.DirEntry, err error) error {
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
	// var repository maven.Repository
	// filepath.WalkDir("/home/juan/.m2/test-repository", func(path string, d fs.DirEntry, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if d.IsDir() {
	// 		return nil
	// 	}
	// 	if strings.HasPrefix(d.Name(), ".") || d.Name() == "_remote.repositories" {
	// 		return nil
	// 	}
	//
	// 	return nil
	// })
}

// func TestListDirs(t *testing.T) {
// 	const path = "/home/juan/.m2/test-repository"
// 	dirs, err := ListDirs(path, -1)
// 	assert.NoError(t, err)
// 	fmt.Println(dirs)
// }
