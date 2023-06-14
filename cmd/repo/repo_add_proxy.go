package repo

import (
	"github.com/coding-wepack/carctl/pkg/artifact/repo"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/coding-wepack/carctl/pkg/util/cmdutil"
	"github.com/spf13/cobra"
)

const addProxyFromFileHelp = `
This command sets the artifact repository proxy source list
for the artifact repository based on the data source in the JSON file.

Examples:

    $ carctl repo add-proxy-from-file datasource.json
`

func newAddProxyFromFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "add-proxy-from-file [FILE]",
		Short:  "Add artifact repo proxy sources from a json file",
		Long:   addProxyFromFileHelp,
		PreRun: cmdutil.PreRun,
		Args:   cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := args[0]
			return repo.AddProxySourceFromFile(filename)
		},
	}

	cmd.Flags().StringVarP(&settings.Cookie, "cookie", "c", "",
		"Used to set to HTTP header 'Cookie' when requesting RESTful APIs")

	return cmd
}
