package fileutil

import (
	"fmt"
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

// func TestListDirs(t *testing.T) {
// 	const path = "/home/juan/.m2/test-repository"
// 	dirs, err := ListDirs(path, -1)
// 	assert.NoError(t, err)
// 	fmt.Println(dirs)
// }
