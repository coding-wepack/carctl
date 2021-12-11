package main

import (
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/settings"
	"github.com/spf13/cobra"
)

func PreRun(cmd *cobra.Command, args []string) {
	if settings.Verbose {
		// debug mode enable
		log.SetDebug()
	}
}
