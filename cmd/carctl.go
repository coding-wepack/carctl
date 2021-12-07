package main

import (
	"os"

	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/camigrater/pkg/log"
)

func main() {
	defer log.Sync()

	cmd, err := newRootCmd()
	if err != nil {
		log.Warn(err.Error())
		os.Exit(1)
	}

	cobra.CheckErr(cmd.Execute())
}
