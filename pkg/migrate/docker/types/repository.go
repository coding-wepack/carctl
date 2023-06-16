package types

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

const (
	ArtifactNameRegexStr = `^[\w.-]+$`
)

var (
	ArtifactNameRegex = regexp.MustCompile(ArtifactNameRegexStr)
)

var (
	ErrForEachContinue = errors.New("continue")
)

type (
	Repository struct {
		IsTls bool `json:"isTls"`
		// Path is image path to repository
		Path string `json:"path"`

		// Count is count of image of the repository
		Count int `json:"-"`

		// Images is image info of the repository
		Images []*Image `json:"images,omitempty"`
	}

	Image struct {
		SrcPath string `json:"srcPath,omitempty"`
		PkgName string `json:"pkgName,omitempty"`
		Version string `json:"version,omitempty"`
	}
)

func (r *Repository) Render(w io.Writer) {
	data := make([][]string, len(r.Images))
	for i, v := range r.Images {
		data[i] = []string{v.PkgName, v.Version, v.SrcPath}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Artifact", "Version", "SrcPath"})
	table.SetFooter([]string{"", "", fmt.Sprintf("Total Images: %d", r.Count)})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}

func (r *Repository) ForEach(fn func(name, srcTag, dstTag string, isTlsSrc, isTlsDst bool) error) error {
	isTlsDst, dst := parseDstUrl(settings.GetDstWithoutSlash())
	path := strings.Trim(r.Path, "/")
	for _, image := range r.Images {
		srcTag := fmt.Sprintf("%s/%s", path, image.SrcPath)
		dstTag := fmt.Sprintf("%s/%s:%s", dst, image.PkgName, image.Version)
		if err := fn(image.SrcPath, srcTag, dstTag, r.IsTls, isTlsDst); err != nil {
			if err == ErrForEachContinue {
				continue
			}
			return err
		}
	}
	return nil
}

func parseDstUrl(dstUrl string) (isTls bool, registryUrl string) {
	isTls = strings.HasPrefix(dstUrl, "https://")
	if isTls {
		registryUrl = strings.TrimPrefix(dstUrl, "https://")
	} else {
		registryUrl = strings.TrimPrefix(dstUrl, "http://")
	}
	return
}
