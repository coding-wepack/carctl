package types

import (
	"io"

	"github.com/coding-wepack/carctl/pkg/migrate/pypi/types/nexus"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

var (
	ErrForEachContinue = errors.New("continue")
)

type (
	Repository struct {
		// Path is file path to repository
		Path string `json:"path"`

		// FileCount is count of files of the repository
		FileCount int `json:"-"`

		Files []nexus.Item `json:"files,omitempty"`
	}
)

func (r *Repository) Render(w io.Writer) {
	data := make([][]string, len(r.Files))
	for i, f := range r.Files {
		data[i] = []string{
			f.Repository, f.Pypi.Name, f.Pypi.Version, f.Path,
		}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Repository", "Name", "Version", "Path"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}

func (r *Repository) ForEach(fn func(downloadUrl, filePath, name, version, sha256Digest string) error) error {
	for _, f := range r.Files {
		if err := fn(f.DownloadURL, f.Path, f.Pypi.Name, f.Pypi.Version, f.Checksum.Sha256); err != nil {
			if err == ErrForEachContinue {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *Repository) AddVersionFile(item nexus.Item) {
	r.Files = append(r.Files, item)
}
