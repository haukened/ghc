// Package main provides the CLI commands for managing organizations
// and their configurations in the GHC application.
//
// This package includes commands to set, remove, and list organizations
// in the configuration file. It interacts with the domain and configfile
// packages to perform these operations.
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/haukened/ghc/internal/configfile"
	"github.com/haukened/ghc/internal/domain"
	"github.com/haukened/ghc/internal/utils"
	"github.com/rodaine/table"
	"github.com/urfave/cli/v3"
)

var (
	ErrNumArguments = fmt.Errorf("incorrect number of arguments")
)

// setOrganization sets the SSH key for the specified organization.
//
// This function requires the organization name and the SSH key path as arguments.
// If the "default" flag is set, the organization is marked as the default.
//
// It performs the following steps:
// 1. Validates the number of arguments and their values.
// 2. Expands the SSH key path to its absolute form.
// 3. Loads the current configuration file.
// 4. Adds or updates the organization in the configuration.
// 5. Writes the updated configuration back to the file.
//
// Returns an error if any of the steps fail.
func setOrganization(ctx context.Context, c *cli.Command) error {
	const nargs = 2
	if c.NArg() != nargs {
		return fmt.Errorf("%s: expected %d, got %d", ErrNumArguments, nargs, c.Args().Len())
	}

	orgName := c.Args().Get(0)
	if orgName == "" {
		return domain.ErrEmptyOrganizationName
	}

	sshKeyPath := c.Args().Get(1)
	if sshKeyPath == "" {
		return domain.ErrEmptySSHKeyPath
	}
	// expand the path to the SSH key
	sshKeyPath = utils.ExpandPath(sshKeyPath)

	// read the current config
	conf, err := configfile.LoadConfig()
	if err != nil {
		if errors.Is(err, configfile.ErrConfigNotFound) {
			// create a new config file
			conf = &domain.Config{
				Organizations: []*domain.Organization{},
			}
		} else {
			return err
		}
	}

	err = conf.SetOrganization(orgName, sshKeyPath, c.Bool("default"))
	if err != nil {
		return err
	}

	// write the configuration back to the file
	err = configfile.WriteConfig(conf)

	// no need to check if the error is nil, because we are going to return it anyway
	return err
}

// removeOrganization removes an organization from the configuration.
//
// This function requires the organization name as an argument.
//
// It performs the following steps:
// 1. Validates the number of arguments and their values.
// 2. Loads the current configuration file.
// 3. Removes the organization from the configuration.
// 4. Writes the updated configuration back to the file.
//
// Returns an error if any of the steps fail.
func removeOrganization(ctx context.Context, c *cli.Command) error {
	const nargs = 1
	if c.NArg() != nargs {
		return fmt.Errorf("%s: expected %d, got %d", ErrNumArguments, nargs, c.Args().Len())
	}

	orgName := c.Args().Get(0)
	if orgName == "" {
		return domain.ErrEmptyOrganizationName
	}

	// read the current config
	conf, err := configfile.LoadConfig()
	if err != nil {
		return err
	}

	// remove the organization from the config
	err = conf.RemoveOrganization(orgName)
	if err != nil {
		return err
	}

	// write the configuration back to the file
	err = configfile.WriteConfig(conf)

	// no need to check if the error is nil, because we are going to return it anyway
	return err
}

// listOrganizations lists all organizations in the configuration.
//
// This function retrieves the current configuration and prints
// the list of organizations to the standard output.
//
// Returns an error if the configuration cannot be loaded.
func listOrganizations(ctx context.Context, c *cli.Command) error {
	// read the current config
	conf, err := configfile.LoadConfig()
	if err != nil {
		return err
	}

	if conf.Organizations == nil {
		return domain.ErrNoOrganizations
	}

	// create formatters
	header := color.New(color.FgGreen, color.Underline).SprintfFunc()

	tbl := table.New("Org Name", "SSH Key Path", "Default")
	tbl.WithHeaderFormatter(header).WithPadding(2)

	// add rows to the table
	for _, org := range conf.Organizations {
		defChar := " "
		if org.IsDefault {
			defChar = "*"
		}
		tbl.AddRow(org.Name, org.SSHKeyPath, defChar)
	}
	fmt.Println("")
	tbl.Print()
	fmt.Println("")
	return nil
}
