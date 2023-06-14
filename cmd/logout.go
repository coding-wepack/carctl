package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/coding-wepack/carctl/cmd/require"
	"github.com/coding-wepack/carctl/pkg/action"
)

const logoutDesc = `
Remove credentials stored for a remote registry.

Examples:

    $ carctl registry logout yourteam-maven.pkg.coding.net
`

func newLogoutCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "logout [host]",
		Short: "Logout from a registry",
		Long:  logoutDesc,
		Args:  require.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			host := args[0]
			return action.NewRegistryLogout(cfg).Run(out, host)
		},
	}
}
