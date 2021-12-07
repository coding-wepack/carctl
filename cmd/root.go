package main

import (
	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/camigrater/pkg/flags"
)

const globalUsage = `The CODING Artifact Repository Manager

Common actions for carctl:

- carctl migrate:    migrate artifacts from local or remote to a CODING Artifact Repository
- carctl pull:       pull artifacts from a CODING Artifact Repository to local (TODO)
- carctl push:       push artifacts from local to a CODING Artifact Repository (TODO)
- carctl search:     search for artifacts (TODO)
- carctl list:       list artifacts (TODO)
`

func newRootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "carctl",
		Short:        "The CODING Artifact Repository Manager.",
		Long:         globalUsage,
		SilenceUsage: true,
	}

	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Make the operation more talkative")

	cmd.AddCommand(
		newMigrateCmd(),
		newVersionCmd(),
	)

	return cmd, nil
}
