package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/coding-wepack/carctl/cmd/require"
	"github.com/coding-wepack/carctl/pkg/action"
)

const migrateHelp = `
This command migrates artifacts from local or remote to a CODING Artifact Registry.

The migrate argument must be an artifact type, available now:

- maven (Default)
- composer 
- pypi 
- npm
- generic
- docker
- helm (TODO)

Examples:

    # Migrate maven repository from local ~/.m2/repository:
    $ carctl migrate maven --dst="https://demo-maven.pkg.coding.net/repository/test-project/dst-repo/"

    # Migrate maven repository with authentication:
    $ carctl migrate maven \
          --src="http://127.0.0.1:8081/repository/maven-releases/" \
          --src-username="test" \
          --src-password="test123" \
          --dst="https://demo-maven.pkg.coding.net/repository/test-project/dst-repo/"

Flags '--src' and '--dst' must be set.
`

func newMigrateCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [TYPE]",
		Short: "Migrate artifacts from anywhere to a CODING Artifact Repository.",
		Long:  migrateHelp,
		Args:  require.MinimumNArgs(1),
	}

	// add subcommands
	cmd.AddCommand(
		newMigrateGenericCmd(cfg, out),
		newMigrateDockerCmd(cfg, out),
		newMigrateMavenCmd(cfg, out),
		newMigrateNpmCmd(cfg, out),
		newMigratePypiCmd(cfg, out),
		newMigrateComposerCmd(cfg, out),
	)

	return cmd
}
