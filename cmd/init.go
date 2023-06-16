package main

import (
	"github.com/coding-wepack/carctl/pkg/log"
	"github.com/coding-wepack/carctl/pkg/settings"
	"github.com/spf13/cobra"
)

func PreRun(cmd *cobra.Command, args []string) {
	if settings.Verbose {
		// debug mode enable
		log.SetDebug()
	}
}
