package types

import (
	"fmt"
	"io"
	"regexp"

	"github.com/pkg/errors"

	"github.com/olekukonko/tablewriter"
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
		// Path is file path to repository
		Path string `json:"path"`

		// FileCount is count of files of the repository
		FileCount int `json:"-"`

		Groups []*Group `json:"groups,omitempty"`
	}

	Group struct {
		Name string `json:"name,omitempty"`

		Artifacts []*Artifact `json:"artifacts,omitempty"`
	}

	Artifact struct {
		Name string `json:"name,omitempty"`

		Versions []*Version `json:"versions,omitempty"`
	}

	Version struct {
		Name string `json:"name,omitempty"`

		Files []*VersionFile `json:"files,omitempty"`
	}

	VersionFile struct {
		// Name: e.g., spring-context-4.3.14.RELEASE.jar | spring-context-4.3.14.RELEASE.pom
		Name string `json:"name,omitempty"`

		// Path: /home/user/.m2/repository/org/springframework/spring-context/4.3.14.RELEASE/spring-context-4.3.14.RELEASE.jar
		Path string `json:"path,omitempty"`
	}
)

type (
	FlattenRepository struct {
		Path string `json:"path"`

		GroupCount    int `json:"-"`
		ArtifactCount int `json:"-"`
		VersionCount  int `json:"-"`
		FileCount     int `json:"-"`

		Files []FlattenVersionFile `json:"files,omitempty"`
	}

	FlattenVersionFile struct {
		Group    string `json:"group"`
		Artifact string `json:"artifact"`
		Version  string `json:"version"`
		Filename string `json:"filename"`
		FilePath string `json:"filePath"`
	}
)

func (r *Repository) Render(w io.Writer) {
	flattenRepository := r.Flatten()
	flattenRepository.Render(w)
}

func (r *Repository) Flatten() *FlattenRepository {
	flattenRepository := &FlattenRepository{
		Path:          r.Path,
		GroupCount:    r.GroupCount(),
		ArtifactCount: r.ArtifactCount(),
		VersionCount:  r.VersionCount(),
		FileCount:     r.GetFileCount(),
		Files:         make([]FlattenVersionFile, 0),
	}

	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for _, f := range v.Files {
					flattenRepository.Files = append(flattenRepository.Files, FlattenVersionFile{
						Group:    g.Name,
						Artifact: a.Name,
						Version:  v.Name,
						Filename: f.Name,
						FilePath: f.Path,
					})
				}
			}
		}
	}

	return flattenRepository
}

func (r *Repository) GetFileCount() int {
	if r.FileCount > 0 {
		return r.FileCount
	}

	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for range v.Files {
					r.FileCount++
				}
			}
		}
	}
	return r.FileCount
}

func (r *Repository) ForEach(fn func(group, artifact, version, path string) error) error {
	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for _, f := range v.Files {
					if err := fn(g.Name, a.Name, v.Name, f.Path); err != nil {
						if err == ErrForEachContinue {
							continue
						}
						return err
					}
				}
			}
		}
	}
	return nil
}

func (r *Repository) AddVersionFile(groupName, artifactName, versionName, filename, filePath string) {
	if !r.HasGroup(groupName) {
		r.AddGroupName(groupName)
	}
	for _, g := range r.Groups {
		if g.Name == groupName {
			if !g.HasArtifact(artifactName) {
				g.AddArtifactName(artifactName)
			}
			for _, art := range g.Artifacts {
				if art.Name == artifactName {
					if !art.HasVersion(versionName) {
						art.AddVersion(versionName)
					}
					for _, v := range art.Versions {
						if v.Name == versionName {

							if !v.HasFile(filename) {
								v.AddFile(filename, filePath)
							}

						}
					}
				}
			}
		}
	}
}

func (r *Repository) HasGroup(groupName string) bool {
	for _, g := range r.Groups {
		if g.Name == groupName {
			return true
		}
	}
	return false
}

func (r *Repository) AddGroupName(group string) {
	r.Groups = append(r.Groups, &Group{Name: group})
}

func (r *Repository) GroupCount() int {
	return len(r.Groups)
}

func (r *Repository) ArtifactCount() int {
	var count int
	for _, g := range r.Groups {
		count += len(g.Artifacts)
	}
	return count
}

func (r *Repository) VersionCount() int {
	var count int
	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			count += len(a.Versions)
		}
	}
	return count
}

func (g *Group) HasArtifact(artifactName string) bool {
	for _, art := range g.Artifacts {
		if art.Name == artifactName {
			return true
		}
	}
	return false
}

func (g *Group) AddArtifactName(artifactName string) {
	g.Artifacts = append(g.Artifacts, &Artifact{Name: artifactName})
}

func (a *Artifact) HasVersion(version string) bool {
	for _, v := range a.Versions {
		if v.Name == version {
			return true
		}
	}
	return false
}

func (a *Artifact) AddVersion(version string) {
	a.Versions = append(a.Versions, &Version{Name: version})
}

func (v *Version) HasFile(filename string) bool {
	for _, f := range v.Files {
		if f.Name == filename {
			return true
		}
	}
	return false
}

func (v *Version) AddFile(filename, filePath string) {
	v.Files = append(v.Files, &VersionFile{Name: filename, Path: filePath})
}

func (f *FlattenRepository) Render(w io.Writer) {
	data := make([][]string, len(f.Files))
	for i, v := range f.Files {
		data[i] = []string{
			v.Group, v.Artifact, v.Version, v.Filename,
		}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Group ID", "Artifact ID", "Version", "File"})
	table.SetFooter([]string{
		f.renderFooterCount("Groups", f.GetGroupCount()),
		f.renderFooterCount("Artifacts", f.GetArtifactCount()),
		f.renderFooterCount("Versions", f.GetVersionCount()),
		f.renderFooterCount("Files", f.GetFileCount()),
	})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}

func (f *FlattenRepository) GetGroupCount() int {
	return f.GroupCount
}

func (f *FlattenRepository) GetArtifactCount() int {
	return f.ArtifactCount
}

func (f *FlattenRepository) GetVersionCount() int {
	return f.VersionCount
}

func (f *FlattenRepository) GetFileCount() int {
	return f.FileCount
}

func (f *FlattenRepository) renderFooterCount(itemName string, count int) string {
	return fmt.Sprintf("Total %s: %d", itemName, count)
}

func (f *FlattenRepository) ForEach(fn func(group, artifact, version, filePath string) error) error {
	for _, r := range f.Files {
		if err := fn(r.Group, r.Artifact, r.Version, r.FilePath); err != nil {
			if err == ErrForEachContinue {
				continue
			}
			return err
		}
	}
	return nil
}
