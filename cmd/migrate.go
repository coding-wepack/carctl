package main

import (
	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/carctl/cmd/require"
	"e.coding.net/codingcorp/carctl/pkg/settings"
)

const migrateHelp = `
This command migrates artifacts from local or remote to a CODING Artifact Repository.

The migrate argument must be an artifact type, available now:

- maven (Default)
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

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [TYPE]",
		Short: "Migrate artifacts from anywhere to a CODING Artifact Repository.",
		Long:  migrateHelp,
		Args:  require.MinimumNArgs(1),
	}

	// required flags
	// migrateCmd.Flags().StringVarP(&flags.Type, "type", "t", "maven", "e.g., -t=maven. Artifact type. Support: [maven]. TODO: [generic、npm ...]")
	cmd.Flags().StringVar(&settings.Src, "src", "", `--src="file://~/.m2/repository", or --src="https://demo-maven.pkg.coding.net/repository/test-project/src-repo/"`)
	cmd.Flags().StringVar(&settings.SrcUsername, "src-username", "", "--src-username=test")
	cmd.Flags().StringVar(&settings.SrcPassword, "src-password", "", "--src-password=test123")
	cmd.Flags().StringVar(&settings.Dst, "dst", "", `--dst="https://demo-maven.pkg.coding.net/repository/test-project/dst-repo/"`)

	// Mark flags as required
	// _ = migrateCmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("src")
	_ = cmd.MarkFlagRequired("dst")

	// optional flags
	cmd.Flags().DurationVar(&settings.Sleep, "sleep", 0, "e.g., --sleep=3s. The default is 0, which means there will be no time to sleep")
	cmd.Flags().IntVarP(&settings.Concurrency, "concurrency", "c", 1, "e.g., -c=2. Concurrency controls for how many artifacts can be pushed concurrently")

	// add subcommands
	cmd.AddCommand(
		newMigrateMavenCmd(),
	)

	return cmd
}
