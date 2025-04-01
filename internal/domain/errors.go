package domain

import "fmt"

var (
	ErrEmptyOrganizationName = fmt.Errorf("organization name cannot be empty")
	ErrEmptySSHKeyPath       = fmt.Errorf("SSH key path cannot be empty")
	ErrNoOrganizations       = fmt.Errorf("no organizations found in the configuration")
	ErrDuplicateOrganization = fmt.Errorf("duplicate organization name found")
	ErrInvalidOrgName        = fmt.Errorf("invalid organization name")
	ErrCantRemoveDefault     = fmt.Errorf("cannot remove the default organization")
	ErrOrganizationNotFound  = fmt.Errorf("organization not found")
)
