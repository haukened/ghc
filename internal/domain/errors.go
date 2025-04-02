package domain

import "errors"

var (
	ErrCantRemoveDefault     = errors.New("cannot remove the default organization")
	ErrDuplicateOrganization = errors.New("duplicate organization name found")
	ErrEmptyOrganizationName = errors.New("organization name cannot be empty")
	ErrEmptySSHKeyPath       = errors.New("SSH key path cannot be empty")
	ErrInvalidOrgName        = errors.New("invalid organization name")
	ErrNoOrganizations       = errors.New("no organizations found in the configuration")
	ErrOrganizationNotFound  = errors.New("organization not found")
	ErrOrgNotFound           = errors.New("organization not found")
)
