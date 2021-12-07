package flags

import (
	"time"
)

var (
	// Verbose is a flag for output more debug info.
	Verbose bool

	// Type is artifact type.
	Type string

	// Src is a file path or an url of the artifacts where you want to migrate.
	Src string

	// Dst is CODING Artifact Repository url you want to migrate to.
	Dst string

	// Sleep is a wait time duration between artifacts upload.
	Sleep time.Duration

	// Concurrency controls how many artifacts can be uploaded simultaneously.
	Concurrency int
)
