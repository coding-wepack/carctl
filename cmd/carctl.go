package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"e.coding.net/codingcorp/carctl/pkg/log"
	"e.coding.net/codingcorp/carctl/pkg/settings"
)

func main() {
	defer log.Sync()

	cfg := new(action.Configuration)
	cmd, err := newRootCmd(cfg, os.Stdout, os.Args[1:])
	if err != nil {
		log.Warn(err.Error())
		os.Exit(1)
	}

	cobra.CheckErr(cmd.Execute())
}

func info(format string, v ...interface{}) {
	format = fmt.Sprintf("%s\n", format)
	_, _ = fmt.Fprintf(os.Stdout, format, v...)
}

func debug(format string, v ...interface{}) {
	if settings.Verbose {
		format = fmt.Sprintf("%s", format)
		log.Info(fmt.Sprintf(format, v...))
	}
}

func warning(format string, v ...interface{}) {
	format = fmt.Sprintf("WARNING: %s\n", format)
	_, _ = fmt.Fprintf(os.Stderr, format, v...)
}
