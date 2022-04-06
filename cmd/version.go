package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "0.1.3"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the CLI version",
		Example: "camigrater version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}

	return cmd
}
