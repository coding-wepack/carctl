package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/coding-wepack/carctl/pkg/action"
	"github.com/coding-wepack/carctl/pkg/migrate/npm"
	"github.com/coding-wepack/carctl/pkg/settings"
)

const migrateNpmHelp = `
This command migrates npm repository from local or remote to a CODING Artifact Repository.

Examples:

    # Migrate local npm repository:
    $ carctl migrate npm --src="https://yourteam.jfrog.io/artifactory/npm/" --dst="https://yourteam-npm.pkg.coding.net/repository/project/npm-repo/"

    # Migrate remote jfrog repository with authentication:
    $ carctl migrate npm \
          --src="http://127.0.0.1:8081/repository/npm-releases/" \
		  --src-type="jfrog" \
          --src-username="test" \
          --src-password="test123" \
          --dst="https://demo-npm.pkg.coding.net/repository/test-project/dst-repo/"
`

func newMigrateNpmCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "npm",
		Short:  "migrate npm repository to a CODING Artifact Repository.",
		Long:   migrateNpmHelp,
		PreRun: PreRun,
		RunE: func(c *cobra.Command, args []string) error {
			return npm.Migrate(cfg, out)
		},
	}

	// required flags
	cmd.Flags().StringVar(&settings.Src, "src", "", `e.g., --src="https://yourteam.jfrog.io/artifactory/npm/", or --src="https://demo-npm.pkg.coding.net/repository/test-project/src-repo/"`)
	cmd.Flags().StringVar(&settings.SrcType, "src-type", "nexus", "e.g., --src-type=jfrog, or --src-type=coding")
	cmd.Flags().StringVar(&settings.SrcUsername, "src-username", "", "e.g., --src-username=test")
	cmd.Flags().StringVar(&settings.SrcPassword, "src-password", "", "e.g., --src-password=test123")
	cmd.Flags().StringVar(&settings.Dst, "dst", "", `e.g., --dst="https://demo-npm.pkg.coding.net/repository/test-project/dst-repo/"`)

	// Mark flags as required
	_ = cmd.MarkFlagRequired("src")
	_ = cmd.MarkFlagRequired("dst")

	// optional flags
	cmd.Flags().DurationVar(&settings.Sleep, "sleep", 0, "e.g., --sleep=3s. The default is 0, which means there will be no time to sleep")
	cmd.Flags().IntVarP(&settings.Concurrency, "concurrency", "c", 1, "e.g., -c=2. Concurrency controls for how many artifacts can be pushed concurrently")
	cmd.Flags().BoolVar(&settings.FailFast, "failFast", false, "exit directly if there was an error found during migration")
	cmd.Flags().IntVar(&settings.MaxFiles, "max-files", -1, "Maximum number of files to be pushed. Negative number means unlimited.")
	cmd.Flags().BoolVarP(&settings.Force, "force", "f", false, "whether push is forced. if exists does no push.")
	cmd.Flags().BoolVar(&settings.DryRun, "dryRun", false, "check need migrate artifacts.")
	cmd.Flags().StringArrayVar(&settings.DropInvalidKey, "dropInvalidKey", []string{}, "e.g., --dropInvalidKey private, used to delete property that cause migration failure in the package.json npm package, such as private: true")

	// TODO: --max-arts
	// TODO: --generate-sha1
	// TODO: --save=/asdfa

	return cmd
}
