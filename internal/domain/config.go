// Package domain contains the core domain models for the GHC application.
package domain

import (
	"fmt"
	"os"
	"regexp"
)

// Config holds the configuration details for the application.
// It contains a list of organizations and their associated SSH keys.
type Config struct {
	Organizations []*Organization `json:"organizations" koanf:"organizations"` // List of organizations and their SSH keys
}

// RemoveOrganization removes an organization from the Config by its name.
// It searches for the organization in the Config's Organizations slice.
// If the organization is not found, it returns an ErrOrganizationNotFound error.
// If the organization to be removed is marked as the default and there are
// other organizations present, it returns an ErrCantRemoveDefault error.
// Otherwise, it removes the organization from the slice and updates the Config.
//
// Parameters:
//   - name: The name of the organization to be removed.
//
// Returns:
//   - error: An error if the organization is not found or cannot be removed,
//     otherwise nil.
func (c *Config) RemoveOrganization(name string) error {
	idxToRemove := -1
	orgToRemove := &Organization{}
	for idx, org := range c.Organizations {
		if org.Name == name {
			idxToRemove = idx
			orgToRemove = org
			break
		}
	}

	if idxToRemove == -1 {
		return fmt.Errorf("%w: %s", ErrOrganizationNotFound, name)
	}

	if orgToRemove.IsDefault && len(c.Organizations) > 1 {
		return ErrCantRemoveDefault
	}

	// this is a go idiom to remove an element from a slice
	// it creates a new slice with the elements before and after the one to remove
	// and appends them together
	// e.g. slice = append(slice[:i], slice[i+1:]...)
	c.Organizations = append(c.Organizations[:idxToRemove], c.Organizations[idxToRemove+1:]...)
	return nil
}

// SetOrganization sets or updates an organization in the configuration.
// If the `isDefault` flag is true, it unsets the default status of all other organizations
// and sets the specified organization as the default. If the organization already exists,
// it updates its SSH key path and default status. If it does not exist, it adds a new
// organization with the provided details.
//
// If there is only one organization in the configuration after the operation, it is
// automatically set as the default regardless of the `isDefault` flag.
//
// Parameters:
//   - name: The name of the organization.
//   - sshKeyPath: The file path to the SSH key associated with the organization.
//   - isDefault: A boolean indicating whether the organization should be set as the default.
//
// Returns:
//   - error: Returns an error if any issue occurs during the operation, otherwise nil.
func (c *Config) SetOrganization(name, sshKeyPath string, isDefault bool) error {
	// if the default flag is set, unset all other organizations
	if isDefault {
		for _, org := range c.Organizations {
			// don't need to check first if the organization is default
			// because we are going to set it to false
			// in Go, setting a bool to false, when it is already false, is a no-op
			// checking would create a CPU branch check when we don't need one.
			org.IsDefault = false
		}
	}

	// check if the organization already exists
	exists := false
	for _, org := range c.Organizations {
		if org.Name == name {
			// update the SSH key path
			org.SSHKeyPath = sshKeyPath
			org.IsDefault = isDefault
			exists = true
			break
		}
	}

	// if the organization does not exist, add it
	if !exists {
		newOrg := &Organization{
			Name:       name,
			SSHKeyPath: sshKeyPath,
			IsDefault:  isDefault,
		}
		c.Organizations = append(c.Organizations, newOrg)
	}

	// if there is only one organization, set it as default regardless of the flag
	if len(c.Organizations) == 1 {
		c.Organizations[0].IsDefault = true
	}

	return nil
}

// Validate checks the configuration for validity. It ensures that:
//   - The Organizations slice is not empty; otherwise, it returns ErrNoOrganizations.
//   - There are no duplicate organization names; otherwise, it returns an error
//     wrapping ErrDuplicateOrganization with the duplicate name.
//   - Each organization in the Organizations slice is valid by calling its Validate method.
//
// If any validation fails, an appropriate error is returned.
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

// Validate checks the validity of the Organization object.
// It performs the following validations:
//  1. Ensures the organization name is not empty. Returns ErrEmptyOrganizationName if empty.
//  2. Validates the organization name against a specific pattern unless it is "default".
//     Returns ErrInvalidOrgName if the name does not match the pattern.
//  3. Ensures the SSH key path is not empty. Returns ErrEmptySSHKeyPath if empty.
//  4. Checks if the SSH key path exists and has the correct file permissions (0600).
//     Returns an appropriate error if the file does not exist or has incorrect permissions.
//
// Returns an error if any of the validations fail, otherwise returns nil.
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
