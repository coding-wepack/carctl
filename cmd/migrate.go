package main

import (
	"os"

	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/camigrater/pkg/flags"
	"e.coding.net/codingcorp/camigrater/pkg/log"
	"e.coding.net/codingcorp/camigrater/pkg/log/logfields"
	"e.coding.net/codingcorp/camigrater/pkg/migrate"
)

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate artifacts from anywhere to a CODING Artifact Repository.",
		Long:  `Migrate artifacts from anywhere to a CODING Artifact Repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := migrate.Migrate(); err != nil {
				log.Error("failed to migrate :(", logfields.String("error", err.Error()))
				os.Exit(1)
			}
		},
	}

	// required flags
	// migrateCmd.Flags().StringVarP(&flags.Type, "type", "t", "maven", "e.g., -t=maven. Artifact type. Support: [maven]. TODO: [generic、npm ...]")
	cmd.Flags().StringVar(&flags.Src, "src", "", `--src="file://$HOME/.m2/repository", or --src="https://demo-maven.pkg.coding.net/repository/test-project/src-repo/"`)
	cmd.Flags().StringVar(&flags.Dst, "dst", "", `--dst="https://demo-maven.pkg.coding.net/repository/test-project/dst-repo/"`)

	// Mark flags as required
	// _ = migrateCmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("src")
	_ = cmd.MarkFlagRequired("dst")

	// optional flags
	cmd.Flags().DurationVar(&flags.Sleep, "sleep", 0, "e.g., --sleep=3s. The default is 0, which means there will be no time to sleep")
	cmd.Flags().IntVarP(&flags.Concurrency, "concurrency", "c", 1, "e.g., -c=2. Concurrency controls for how many artifacts can be pushed concurrently")

	return cmd
}
