package main

import (
	"io"

	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/carctl/cmd/require"
	"e.coding.net/codingcorp/carctl/pkg/action"
)

const migrateHelp = `
This command migrates artifacts from local or remote to a CODING Artifact Registry.

The migrate argument must be an artifact type, available now:

- maven (Default)
- composer 
- pypi 
- npm (TODO)
- generic (TODO)
- docker (TODO)
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
		newMigrateMavenCmd(cfg, out),
		newMigratePypiCmd(cfg, out),
		newMigrateComposerCmd(cfg, out),
	)

	return cmd
}
