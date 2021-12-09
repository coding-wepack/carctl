package maven

import (
	"regexp"
)

const (
	ArtifactNameRegexStr = `^[\w.-]+$`
)

var (
	ArtifactNameRegex = regexp.MustCompile(ArtifactNameRegexStr)
)

type Repository struct {
	// Path is file path to repository
	Path string `json:"path"`

	Groups []Group `json:"groups,omitempty"`
}

type Group struct {
	Name string `json:"name,omitempty"`

	Artifacts []Artifact `json:"artifacts,omitempty"`
}

type Artifact struct {
	Name string `json:"name,omitempty"`

	Versions []Version `json:"versions,omitempty"`
}

type Version struct {
	Name string `json:"name,omitempty"`

	Files []VersionFile `json:"files,omitempty"`
}

type VersionFile struct {
	// Name: e.g., spring-context-4.3.14.RELEASE.jar | spring-context-4.3.14.RELEASE.pom
	Name string `json:"name,omitempty"`

	// Path: /home/user/.m2/repository/org/springframework/spring-context/4.3.14.RELEASE/spring-context-4.3.14.RELEASE.jar
	Path string `json:"path,omitempty"`
}

func (r *Repository) AddVersionFile(groupName, artifactName, versionName, filename, filePath string) error {

	return nil
}

func (r *Repository) AddGroupName(group string) {
	r.Groups = append(r.Groups, Group{Name: group})
}

func (r *Repository) GetGroup(groupName string) (has bool, group Group) {
	for _, g := range r.Groups {
		if g.Name == groupName {
			return true, g
		}
	}
	return false, Group{}
}
