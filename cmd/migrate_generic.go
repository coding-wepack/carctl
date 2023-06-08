package main

import (
	"io"

	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/migrate/generic"
	"e.coding.net/codingcorp/carctl/pkg/settings"
)

const migrateGenericHelp = `
This command migrates generic repository from local or remote to a CODING Artifact Repository.

Examples:

    # Migrate local generic repository:
    $ carctl migrate generic --src="file://~/generic/artifacts/" --dst="https://yourteam-generic.pkg.coding.net/repository/project/generic-repo/"

    # Migrate remote jfrog repository with authentication:
    $ carctl migrate generic \
          --src="http://127.0.0.1:8081/repository/generic-releases/" \
		  --src-type="jfrog" \
          --src-username="test" \
          --src-password="test123" \
          --dst="https://demo-generic.pkg.coding.net/repository/test-project/dst-repo/"
`

func newMigrateGenericCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "generic",
		Short:  "migrate generic repository to a CODING Artifact Repository.",
		Long:   migrateGenericHelp,
		PreRun: PreRun,
		RunE: func(c *cobra.Command, args []string) error {
			return generic.Migrate(cfg, out)
		},
	}

	// required flags
	cmd.Flags().StringVar(&settings.Src, "src", "", `e.g., --src="file://~/generic/artifacts/", or --src="https://demo-generic.pkg.coding.net/repository/test-project/src-repo/"`)
	cmd.Flags().StringVar(&settings.SrcType, "src-type", "nexus", "e.g., --src-type=jfrog, or --src-type=coding")
	cmd.Flags().StringVar(&settings.SrcUsername, "src-username", "", "e.g., --src-username=test")
	cmd.Flags().StringVar(&settings.SrcPassword, "src-password", "", "e.g., --src-password=test123")
	cmd.Flags().StringVar(&settings.Dst, "dst", "", `e.g., --dst="https://demo-generic.pkg.coding.net/repository/test-project/dst-repo/"`)

	// Mark flags as required
	_ = cmd.MarkFlagRequired("src")
	_ = cmd.MarkFlagRequired("dst")

	// optional flags
	cmd.Flags().DurationVar(&settings.Sleep, "sleep", 0, "e.g., --sleep=3s. The default is 0, which means there will be no time to sleep")
	cmd.Flags().IntVarP(&settings.Concurrency, "concurrency", "c", 1, "e.g., -c=2. Concurrency controls for how many artifacts can be pushed concurrently")
	cmd.Flags().BoolVar(&settings.FailFast, "failFast", false, "exit directly if there was an error found during migration")
	cmd.Flags().IntVar(&settings.MaxFiles, "max-files", -1, "Maximum number of files to be pushed. Negative number means unlimited.")
	cmd.Flags().BoolVarP(&settings.Force, "force", "f", false, "whether push is forced. if exists does no push.")

	// TODO: --max-arts
	// TODO: --generate-sha1
	// TODO: --save=/asdfa

	return cmd
}
