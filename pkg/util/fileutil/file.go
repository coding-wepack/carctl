package fileutil

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func IsFileInvisible(filename string) bool {
	return strings.HasPrefix(filename, ".")
}

// IsFileExists checks if file specified exists
func IsFileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateFileIfNotExists creates file specified if not exists
func CreateFileIfNotExists(name string) error {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		// create
		return CreateRecursively(name)
	} else if os.IsExist(err) {
		// already exists
		return nil
	}
	return err
}

// CreateRecursively creates file recursively.
func CreateRecursively(name string) error {
	if !strings.Contains(name, "/") {
		// just a single filename
		_, err := os.Create(name)
		return err
	}

	i := strings.LastIndex(name, "/")
	path := name[:i]

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	_, err := os.Create(name)
	return err
}

func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func WriteFile(filePath string, read io.ReadCloser) error {
	file, err := os.Create(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", filePath)
	}
	_, err = io.Copy(file, read)
	if err != nil {
		return errors.Wrapf(err, "failed to write content to file")
	}
	return nil
}
