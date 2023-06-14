package repo

import "github.com/spf13/cobra"

const repoHelp = `
The repo command can handle and control artifact repository, like:
- create artifact repositories(TODO)
- insert artifact repository proxies

Examples:

    $ carctl repo add-proxy-from-file datasource.json
`

func NewRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo [COMMAND]",
		Short: "The repo command can handle and control artifact repository.",
		Long:  repoHelp,
	}

	cmd.AddCommand(
		newAddProxyFromFileCmd(),
	)

	return cmd
}
