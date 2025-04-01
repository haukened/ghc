package domain

import (
	"errors"
	"os"
	"testing"

	"github.com/haukened/ghc/internal/utils"
)

func TestConfigValidate(t *testing.T) {
	privateKey, _ := utils.GenerateTestSSHKey(t)

	tests := []struct {
		name    string
		config  Config
		expects error
	}{
		{
			name:    "No organizations",
			config:  Config{},
			expects: ErrNoOrganizations,
		},
		{
			name: "Duplicate organization names",
			config: Config{
				Organizations: []*Organization{
					{Name: "org1", SSHKeyPath: privateKey},
					{Name: "org1", SSHKeyPath: privateKey},
				},
			},
			expects: ErrDuplicateOrganization,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if !errors.Is(err, tt.expects) {
				t.Errorf("expected %v, got %v", tt.expects, err)
			}
		})
	}
}

func TestConfigValidate_InvalidOrganization(t *testing.T) {
	privateKey, _ := utils.GenerateTestSSHKey(t)

	tests := []struct {
		name    string
		config  Config
		expects error
	}{
		{
			name: "Invalid organization with empty name",
			config: Config{
				Organizations: []*Organization{
					{Name: "", SSHKeyPath: privateKey},
				},
			},
			expects: ErrEmptyOrganizationName,
		},
		{
			name: "Invalid organization with empty SSH key path",
			config: Config{
				Organizations: []*Organization{
					{Name: "org1", SSHKeyPath: ""},
				},
			},
			expects: ErrEmptySSHKeyPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if !errors.Is(err, tt.expects) {
				t.Errorf("expected %v, got %v", tt.expects, err)
			}
		})
	}
}

func TestConfigValidate_ValidConfig(t *testing.T) {
	privateKey, _ := utils.GenerateTestSSHKey(t)

	config := Config{
		Organizations: []*Organization{
			{Name: "org1", SSHKeyPath: privateKey},
			{Name: "org2", SSHKeyPath: privateKey},
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestOrganizationValidate_InvalidOrgName(t *testing.T) {
	privateKey, _ := utils.GenerateTestSSHKey(t)

	tests := []struct {
		name    string
		org     Organization
		expects error
	}{
		{
			name:    "Invalid organization name with special characters",
			org:     Organization{Name: "Invalid!Org", SSHKeyPath: privateKey},
			expects: ErrInvalidOrgName,
		},
		{
			name:    "Invalid organization name with spaces",
			org:     Organization{Name: "Invalid Org", SSHKeyPath: privateKey},
			expects: ErrInvalidOrgName,
		},
		{
			name:    "Valid organization name",
			org:     Organization{Name: "valid-org", SSHKeyPath: privateKey},
			expects: nil,
		},
		{
			name:    "Default organization name",
			org:     Organization{Name: "default", SSHKeyPath: privateKey},
			expects: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.org.Validate()
			if !errors.Is(err, tt.expects) {
				t.Errorf("expected %v, got %v", tt.expects, err)
			}
		})
	}
}

func TestOrganizationValidate_FileNotExist(t *testing.T) {
	org := Organization{
		Name:       "org1",
		SSHKeyPath: "/nonexistent/path/to/ssh_key",
	}

	err := org.Validate()
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got %v", err)
	}
}

func TestOrganizationValidate_InvalidPermissions(t *testing.T) {
	// Create a temporary file to act as an SSH key with invalid permissions
	file, err := os.CreateTemp("", "invalid_permissions_ssh_key_*.key")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	// Set incorrect permissions for the SSH key
	if err := os.Chmod(file.Name(), 0644); err != nil {
		t.Fatalf("failed to set file permissions: %v", err)
	}

	org := Organization{
		Name:       "org1",
		SSHKeyPath: file.Name(),
	}

	err = org.Validate()
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("expected os.ErrPermission, got %v", err)
	}
}

func TestConfigRemoveOrganization(t *testing.T) {
	privateKey, _ := utils.GenerateTestSSHKey(t)

	config := Config{
		Organizations: []*Organization{
			{Name: "org1", SSHKeyPath: privateKey, IsDefault: true},
			{Name: "org2", SSHKeyPath: privateKey},
		},
	}

	tests := []struct {
		name    string
		orgName string
		expects error
	}{
		// this one has to go first because it is the default organization
		{
			name:    "Remove default organization when there is more than one",
			orgName: "org1",
			expects: ErrCantRemoveDefault,
		},
		{
			name:    "Remove non-default organization",
			orgName: "org2",
			expects: nil,
		},
		{
			name:    "Remove default organization when its the last one",
			orgName: "org1",
			expects: nil,
		},
		{
			name:    "Remove non-existent organization",
			orgName: "org3",
			expects: ErrOrganizationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.RemoveOrganization(tt.orgName)
			if !errors.Is(err, tt.expects) {
				t.Errorf("expected %v, got %v", tt.expects, err)
			}
		})
	}
}

func TestConfigSetOrganization(t *testing.T) {
	privateKey, _ := utils.GenerateTestSSHKey(t)

	tests := []struct {
		name       string
		config     Config
		orgName    string
		sshKeyPath string
		isDefault  bool
		expects    error
	}{
		{
			name: "Add new organization",
			config: Config{
				Organizations: []*Organization{},
			},
			orgName:    "org1",
			sshKeyPath: privateKey,
			isDefault:  false,
			expects:    nil,
		},
		{
			name: "Update existing organization",
			config: Config{
				Organizations: []*Organization{
					{Name: "org1", SSHKeyPath: "/old/path", IsDefault: false},
				},
			},
			orgName:    "org1",
			sshKeyPath: privateKey,
			isDefault:  true,
			expects:    nil,
		},
		{
			name: "Set organization as default",
			config: Config{
				Organizations: []*Organization{
					{Name: "org1", SSHKeyPath: privateKey, IsDefault: false},
					{Name: "org2", SSHKeyPath: privateKey, IsDefault: false},
				},
			},
			orgName:    "org2",
			sshKeyPath: privateKey,
			isDefault:  true,
			expects:    nil,
		},
		{
			name: "Automatically set single organization as default",
			config: Config{
				Organizations: []*Organization{},
			},
			orgName:    "org1",
			sshKeyPath: privateKey,
			isDefault:  false,
			expects:    nil,
		},
		{
			name: "set org with invalid key",
			config: Config{
				Organizations: []*Organization{},
			},
			orgName:    "org1",
			sshKeyPath: "/invalid/path/to/key",
			isDefault:  false,
			expects:    os.ErrNotExist,
		},
		{
			name: "set existing org with invalid key",
			config: Config{
				Organizations: []*Organization{
					{Name: "org1", SSHKeyPath: privateKey, IsDefault: false},
				},
			},
			orgName:    "org1",
			sshKeyPath: "/invalid/path/to/key",
			isDefault:  false,
			expects:    os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.SetOrganization(tt.orgName, tt.sshKeyPath, tt.isDefault)
			if !errors.Is(err, tt.expects) {
				t.Errorf("expected %v, got %v", tt.expects, err)
			}

			// Additional checks for specific scenarios
			if tt.name == "Set organization as default" {
				for _, org := range tt.config.Organizations {
					if org.Name == tt.orgName && !org.IsDefault {
						t.Errorf("expected organization %s to be default", tt.orgName)
					}
					if org.Name != tt.orgName && org.IsDefault {
						t.Errorf("expected organization %s to not be default", org.Name)
					}
				}
			}
		})
	}
}
