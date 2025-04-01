// Package domain contains the core domain models for the GHC application.
package domain

import (
	"fmt"
	"os"
)

var (
	ErrEmptyOrganizationName = fmt.Errorf("organization name cannot be empty")
	ErrEmptySSHKeyPath       = fmt.Errorf("SSH key path cannot be empty")
	ErrNoOrganizations       = fmt.Errorf("no organizations found in the configuration")
	ErrDuplicateOrganization = fmt.Errorf("duplicate organization name found")
)

// Config holds the configuration details for the application.
// It contains a list of organizations and their associated SSH keys.
type Config struct {
	Organizations []*Organization `json:"organizations" koanf:"organizations"` // List of organizations and their SSH keys
}

func (c *Config) Validate() error {
	if len(c.Organizations) == 0 {
		return ErrNoOrganizations
	}
	nameSet := make(map[string]struct{})
	for _, org := range c.Organizations {
		if _, exists := nameSet[org.Name]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateOrganization, org.Name)
		}
		nameSet[org.Name] = struct{}{}
		if err := org.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Organization represents a GitHub organization and its associated SSH key.
// The IsDefault field indicates if this is the default organization.
type Organization struct {
	Name       string `json:"name" koanf:"name"`                 // Name of the organization
	SSHKeyPath string `json:"ssh_key_path" koanf:"ssh_key_path"` // Path to the SSH key for the organization
	IsDefault  bool   `json:"is_default" koanf:"is_default"`     // Indicates if this is the default organization
}

func (o *Organization) Validate() error {
	if o.Name == "" {
		return ErrEmptyOrganizationName
	}
	if o.SSHKeyPath == "" {
		return ErrEmptySSHKeyPath
	}
	if fileInfo, err := os.Stat(o.SSHKeyPath); err == nil {
		if fileInfo.Mode().Perm() != 0600 {
			return fmt.Errorf("%w: %s has incorrect permissions: %v", os.ErrPermission, o.SSHKeyPath, fileInfo.Mode().Perm())
		}
	} else if os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", os.ErrNotExist, o.SSHKeyPath)
	}

	return nil
}
