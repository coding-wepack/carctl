package maven

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
