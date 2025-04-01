// Package domain contains the core domain models for the GHC application.
package domain

import (
	"fmt"
	"os"
	"regexp"
)

var (
	ErrEmptyOrganizationName = fmt.Errorf("organization name cannot be empty")
	ErrEmptySSHKeyPath       = fmt.Errorf("SSH key path cannot be empty")
	ErrNoOrganizations       = fmt.Errorf("no organizations found in the configuration")
	ErrDuplicateOrganization = fmt.Errorf("duplicate organization name found")
	ErrInvalidOrgName        = fmt.Errorf("invalid organization name")
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
	// check if the organization name is empty
	if o.Name == "" {
		return ErrEmptyOrganizationName
	}
	// check if the organization name matches the requirements | default
	reg := regexp.MustCompile(`^[a-z0-9](?:[a-z0-9\-]{0,37}[a-z0-9])?$`)
	if o.Name != "default" {
		if !reg.MatchString(o.Name) {
			return ErrInvalidOrgName
		}
	}
	// check if the SSH key path is empty
	if o.SSHKeyPath == "" {
		return ErrEmptySSHKeyPath
	}
	// check if the SSH key path is valid
	if fileInfo, err := os.Stat(o.SSHKeyPath); err == nil {
		// check permissions are secure and correct
		if fileInfo.Mode().Perm() != 0600 {
			return fmt.Errorf("%w: %s has incorrect permissions: %v", os.ErrPermission, o.SSHKeyPath, fileInfo.Mode().Perm())
		}
	} else if os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", os.ErrNotExist, o.SSHKeyPath)
	}

	return nil
}
