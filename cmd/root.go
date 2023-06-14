package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/coding-wepack/carctl/cmd/repo"
	"github.com/coding-wepack/carctl/pkg/action"
	"github.com/coding-wepack/carctl/pkg/config"
	"github.com/coding-wepack/carctl/pkg/registry"
	"github.com/coding-wepack/carctl/pkg/settings"
)

const globalUsage = `The CODING Artifact Registry Manager

Common actions for carctl:

- carctl login:      login to a CODING Artifact Registry
- carctl logout:     logout from a CODING Artifact Registry
- carctl repo:       handle and control artifact repository
- carctl migrate:    migrate artifacts from local or remote to a CODING Artifact Repository
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
		newMigrateCmd(cfg, out),
		newVersionCmd(),
		newLoginCmd(cfg, out),
		newLogoutCmd(cfg, out),
		repo.NewRepoCmd(),
	)

	return cmd, nil
}
