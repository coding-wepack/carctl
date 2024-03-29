package settings

import (
	"strings"
	"time"
)

var (
	// Verbose is a flag for output more debug info.
	Verbose bool

	// Insecure allow connections to TLS registry without certs
	Insecure bool

	// Cookie is WePack user's cookie
	Cookie string

	// Username is username
	Username string

	// Password is password.
	Password string

	// PasswordFromStdin reads password from stdin if true.
	PasswordFromStdin bool

	// FailFast will return error once occurred if true
	FailFast bool

	// Src is a file path or an url of the artifacts where you want to migrate.
	Src string

	// SrcType is the src type, [nexus,coding]
	SrcType string

	// SrcUsername is username of Src
	SrcUsername string

	// SrcPassword is password of Src
	SrcPassword string

	// Dst is CODING Artifact Repository url you want to migrate to.
	Dst string

	// Sleep is a wait time duration between artifacts upload.
	Sleep time.Duration

	// Concurrency controls how many artifacts can be uploaded simultaneously.
	Concurrency int

	// MaxFiles are maximum files which would be uploaded
	MaxFiles int

	// Force controls whether push is forced. if exists does no push.
	Force bool

	// Prefix use to filter generic artifacts.
	Prefix string

	// DryRun is print need migrate artifacts
	DryRun bool

	LargeFileMode bool

	//
	DropInvalidKey []string
)

func GetSrcWithoutSlash() string {
	return strings.Trim(Src, "/")
}

func GetDstWithoutSlash() string {
	return strings.Trim(Dst, "/")
}

func GetSrcHasSubSlash() string {
	return GetSrcWithoutSlash() + "/"
}

func GetDstHasSubSlash() string {
	return GetDstWithoutSlash() + "/"
}
