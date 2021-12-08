package main

import (
	"io"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"github.com/spf13/cobra"
)

const registryHelp = `
This command consists of multiple subcommands to interact with registries.
`

func newRegistryCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "login to or logout from a registry",
		Long:  registryHelp,
	}

	cmd.AddCommand(
		newRegistryLoginCmd(cfg, out),
		newRegistryLogoutCmd(cfg, out),
	)

	return cmd
}
