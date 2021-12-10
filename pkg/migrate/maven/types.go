package maven

import (
	"io"
	"regexp"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

const (
	ArtifactNameRegexStr = `^[\w.-]+$`
)

var (
	ArtifactNameRegex = regexp.MustCompile(ArtifactNameRegexStr)
)

type (
	Repository struct {
		// Path is file path to repository
		Path string `json:"path"`

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

		Files []FlattenVersionFile `json:"files,omitempty"`
	}

	FlattenVersionFile struct {
		Filename string `json:"filename"`
		Version  string `json:"version"`
		Artifact string `json:"artifact"`
		Group    string `json:"group"`
	}
)

func (r *Repository) Render(w io.Writer) {
	flattenRepository := r.Flatten()
	flattenRepository.Render(w)
}

func (r *Repository) Flatten() *FlattenRepository {
	flattenRepository := &FlattenRepository{
		Path:  r.Path,
		Files: make([]FlattenVersionFile, 0),
	}

	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for _, f := range v.Files {
					flattenRepository.Files = append(flattenRepository.Files, FlattenVersionFile{
						Filename: f.Name,
						Version:  v.Name,
						Artifact: a.Name,
						Group:    g.Name,
					})
				}
			}
		}
	}

	return flattenRepository
}

func (r *Repository) VersionCount() int {
	var count int
	for _, g := range r.Groups {
		for _, a := range g.Artifacts {
			for _, v := range a.Versions {
				for range v.Files {
					count++
				}
			}
		}
	}
	return count
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
	total := len(f.Files)
	data := make([][]string, total)
	for i, v := range f.Files {
		data[i] = []string{
			v.Group, v.Artifact, v.Version, v.Filename,
		}
	}

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Group", "Artifact", "Version", "File"})
	table.SetFooter([]string{"Total", strconv.Itoa(total), "", ""})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	table.AppendBulk(data)
	table.Render()
}
