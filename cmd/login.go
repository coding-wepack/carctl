package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"e.coding.net/codingcorp/carctl/pkg/action"
	"github.com/moby/term" // nolint
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"e.coding.net/codingcorp/carctl/cmd/require"
	"e.coding.net/codingcorp/carctl/pkg/settings"
)

const loginDesc = `
Authenticate to a remote CODING Artifact Registry.

Examples:

    $ carctl login yourteam-maven.pkg.coding.net -u USERNAME -p PASSWORD

    # login by stdin
    $ echo PASSWORD | carctl login yourteam-maven.pkg.coding.net -u USERNAME --password-stdin
`

func newLoginCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login [host]",
		Short: "login to a CODING Artifact Registry",
		Long:  loginDesc,
		Args:  require.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			host := args[0]
			username, password, err := getUsernamePassword(settings.Username, settings.Password, settings.PasswordFromStdin)
			if err != nil {
				return err
			}

			if settings.Verbose {
				debug("Got username: [%s], password: [%s]", username, password)
			}

			return action.NewRegistryLogin(cfg).Run(out, host, username, password, settings.Insecure)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&settings.Username, "username", "u", "", "registry username")
	f.StringVarP(&settings.Password, "password", "p", "", "registry password or identity token")
	f.BoolVarP(&settings.PasswordFromStdin, "password-stdin", "", false, "read password or identity token from stdin")
	f.BoolVarP(&settings.Insecure, "insecure", "", false, "allow connections to TLS registry without certs")

	return cmd
}

// Adapted from https://github.com/oras-project/oras
func getUsernamePassword(usernameOpt string, passwordOpt string, passwordFromStdinOpt bool) (string, string, error) {
	var err error
	username := usernameOpt
	password := passwordOpt

	if passwordFromStdinOpt {
		passwordFromStdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", "", err
		}
		password = strings.TrimSuffix(string(passwordFromStdin), "\n")
		password = strings.TrimSuffix(password, "\r")
	} else {
		warning("Using --password via the CLI is insecure. Use --password-stdin.")

		if username == "" {
			username, err = readLine("Username: ", false)
			if err != nil {
				return "", "", err
			}
			username = strings.TrimSpace(username)
		}

		if password == "" {
			password, err = readLine("Password: ", true)
			if err != nil {
				return "", "", err
			} else if password == "" {
				return "", "", errors.New("password required")
			}
		}
	}

	return username, password, nil
}

// Copied/adapted from https://github.com/oras-project/oras
func readLine(prompt string, silent bool) (string, error) {
	fmt.Print(prompt)
	if silent {
		fd := os.Stdin.Fd()
		state, err := term.SaveState(fd)
		if err != nil {
			return "", err
		}
		term.DisableEcho(fd, state)
		defer term.RestoreTerminal(fd, state)
	}

	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	if silent {
		fmt.Println()
	}

	return string(line), nil
}
