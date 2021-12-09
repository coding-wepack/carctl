package fileutil

import (
	"os"
	"sort"
	"strings"
)

// func ListDirs(path string, n int) ([]os.DirEntry, error) {
// 	dirs, err := os.ReadDir(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, dir := range dirs {
// 		f, err := dir.Info()
// 		if err != nil {
// 			return nil, err
// 		}
// 		fmt.Println(filepath.Join(path, f.Name()))
// 	}
//
// 	return dirs, nil
// }

func ListDirNames(path string, n int) ([]string, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, dir := range dirEntries {
		if n >= 0 && len(dirs) >= n {
			break
		}
		if dir.IsDir() {
			dirs = append(dirs, dir.Name())
		}
	}

	return dirs, nil
}

func ListVisibleDirNames(path string, n int) ([]string, error) {
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	if len(dirEntries) == 0 {
		return []string{}, nil
	}

	var visibleDirs []string
	for _, dir := range dirEntries {
		if n >= 0 && len(visibleDirs) >= n {
			break
		}
		// skip if it is not a dir
		if !dir.IsDir() {
			continue
		}
		// starts with "." means invisible
		if !strings.HasPrefix(dir.Name(), ".") {
			visibleDirs = append(visibleDirs, dir.Name())
		}
	}

	return visibleDirs, nil
}

func ListVisibleDirNamesWithSort(path string, n int) ([]string, error) {
	dirs, err := ListVisibleDirNames(path, n)
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 {
		return dirs, nil
	}

	if sort.StringsAreSorted(dirs) {
		return dirs, nil
	} else {
		sort.Strings(dirs)
	}

	return dirs, nil
}
