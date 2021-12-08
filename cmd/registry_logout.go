package main

import (
	"io"

	"e.coding.net/codingcorp/carctl/cmd/require"
	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/action/login"
	"github.com/spf13/cobra"
)

const registryLogoutDesc = `
Remove credentials stored for a remote registry.

Examples:

    $ carctl registry logout yourteam-maven.pkg.coding.net
`

func newRegistryLogoutCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "logout [host]",
		Short: "logout from a registry",
		Long:  registryLogoutDesc,
		Args:  require.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			host := args[0]
			return login.NewRegistryLogout(cfg).Run(out, host)
		},
	}
}
