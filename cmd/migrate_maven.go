package main

import (
	"e.coding.net/codingcorp/carctl/pkg/migrate/maven"
	"github.com/spf13/cobra"
)

const migrateMavenHelp = `
This command migrates maven repository from local or remote to a CODING Artifact Repository.

Examples:

    # Migrate local maven repository:
    $ carctl migrate maven --src="file://$HOME/.m2/repository" --dst=https://yourteam-maven.pkg.coding.net/repository/project/maven-repo/

    # Migrate remote nexus repository with authentication:
    $ carctl migrate maven \
          --src="http://127.0.0.1:8081/repository/maven-releases/" \
          --src-username="test" \
          --src-password="test123" \
          --dst="https://demo-maven.pkg.coding.net/repository/test-project/dst-repo/"
`

func newMigrateMavenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maven",
		Short: "migrate maven repository to a CODING Artifact Repository.",
		Long:  migrateMavenHelp,
		RunE: func(c *cobra.Command, args []string) error {
			return maven.Migrate()
		},
	}

	return cmd
}
