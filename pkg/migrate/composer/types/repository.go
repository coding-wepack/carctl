package types

import (
	"io"

	"github.com/coding-wepack/carctl/pkg/migrate/composer/types/nexus"
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

		Files []*nexus.ComposerItem `json:"files,omitempty"`
	}
)

func (r *Repository) ForEach(fn func(path, version, downloadUrl string) error) error {
	for _, f := range r.Files {
		if err := fn(f.Name, f.Version, f.Dist.URL); err != nil {
			if err == ErrForEachContinue {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *Repository) AddVersionFileList(items []*nexus.ComposerItem) {
	r.Files = append(r.Files, items...)
}

func (r *Repository) Render(w io.Writer) {
	data := make([][]string, len(r.Files))
	for i, v := range r.Files {
		data[i] = []string{
			v.Name, v.Version, v.Dist.URL,
		}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Name", "Version", "Path"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}
