package main

import (
	"io"

	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/config"
	"e.coding.net/codingcorp/carctl/pkg/registry"
	"e.coding.net/codingcorp/carctl/pkg/settings"
)

const globalUsage = `The CODING Artifact Registry Manager

Common actions for carctl:

- carctl login:      login to a CODING Artifact Registry
- carctl logout:     logout from a CODING Artifact Registry
- carctl migrate:    migrate artifacts from local or remote to a CODING Artifact Repository
- carctl pull:       pull artifacts from a CODING Artifact Repository to local (TODO)
- carctl push:       push artifacts from local to a CODING Artifact Repository (TODO)
- carctl search:     search for artifacts (TODO)
- carctl list:       list artifacts (TODO)
`

func newRootCmd(cfg *action.Configuration, out io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "carctl",
		Short:        "The CODING Artifact Registry Manager.",
		Long:         globalUsage,
		SilenceUsage: true,
	}

	cmd.CompletionOptions.DisableDefaultCmd = true

	cmd.PersistentFlags().BoolVarP(&settings.Verbose, "verbose", "v", false, "Make the operation more talkative")

	// registry client
	registryClient, err := registry.NewClient(
		registry.ClientOptVerbose(settings.Verbose),
		registry.ClientOptWriter(out),
		// TODO: config file configurable
		registry.ClientOptConfigFile(config.DefaultConfigFilePath()),
	)
	if err != nil {
		return nil, err
	}
	cfg.RegistryClient = registryClient

	cmd.AddCommand(
		newMigrateCmd(),
		newVersionCmd(),
		newRegistryCmd(cfg, out),
	)

	return cmd, nil
}
