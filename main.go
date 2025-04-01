package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

var (
	version   string = "development"
	buildDate string = "unknown"
)

func init() {
	cli.VersionPrinter = func(c *cli.Command) {
		fmt.Fprintf(c.Root().Writer, "%s %s\n", c.Name, version)
		fmt.Fprintf(c.Root().Writer, "Build date: %s\n", buildDate)
	}
}

func main() {
	app := &cli.Command{
		Name:                  "ghc",
		Version:               version,
		Copyright:             "(c) 2025 David Haukeness, distributed under the GNU General Public License v3.0",
		Usage:                 "Clone GitHub repositories with SSH keys for different organizations",
		UsageText:             "ghc <command> [command options] [arguments...]",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:     "organization",
				Aliases:  []string{"org"},
				Usage:    "Manage GitHub organizations",
				Category: "Configuration",
				Commands: []*cli.Command{
					{
						Name:   "set",
						Usage:  "Sets the SSH key for the specified organization",
						Action: setOrganization,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:    "default",
								Aliases: []string{"D"},
								Usage:   "Set this organization as the default",
							},
						},
						ArgsUsage: "ORG_NAME SSH_KEY_PATH",
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List all organizations in the configuration",
						Action:  listOrganizations,
					},
					{
						Name:      "remove",
						Aliases:   []string{"rm"},
						Usage:     "Remove an organization from the configuration",
						Action:    removeOrganization,
						ArgsUsage: "ORG_NAME",
					},
				},
			},
			{
				Name:     "clone",
				Category: "Repository Management",
				Usage:    "Clone a GitHub repository using the specified SSH key",
				Action: func(ctx context.Context, c *cli.Command) error {
					repo := c.Args().Get(0)
					if repo == "" {
						return fmt.Errorf("repository name is required")
					}
					log.Printf("Cloning repository: %s\n", repo)
					// Here you would add the logic to clone the repository using the SSH key
					return nil
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		writeError(err)
		os.Exit(1)
	}
}

func writeError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
