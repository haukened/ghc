package domain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"
)

func generateTestSSHKey(t *testing.T) (privateKeyPath, publicKeyPath string) {
	// This function generates a temporary RSA key pair for testing purposes.
	t.Helper()

	// Create a temporary directory for the keys
	tempDir := t.TempDir()
	privateKeyPath = filepath.Join(tempDir, "id_rsa")
	publicKeyPath = privateKeyPath + ".pub"

	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	// Write private key to file, with permissions 0600
	privateKeyFile, err := os.OpenFile(privateKeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatalf("failed to create private key file: %v", err)
	}
	defer privateKeyFile.Close()

	// Encode the private key in PEM format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	}
	if err := pem.Encode(privateKeyFile, privBlock); err != nil {
		t.Fatalf("failed to write PEM private key: %v", err)
	}

	// Generate public key in OpenSSH format
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to create SSH public key: %v", err)
	}
	pubBytes := ssh.MarshalAuthorizedKey(pub)

	// Write public key
	if err := os.WriteFile(publicKeyPath, pubBytes, 0644); err != nil {
		t.Fatalf("failed to write public key: %v", err)
	}

	// return the paths of the generated keys
	return privateKeyPath, publicKeyPath
}

func TestConfigValidate(t *testing.T) {
	privateKey, _ := generateTestSSHKey(t)

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
	privateKey, _ := generateTestSSHKey(t)

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
	privateKey, _ := generateTestSSHKey(t)

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

func TestOrganizationValidate_NameValidation(t *testing.T) {
	privateKey, _ := generateTestSSHKey(t)

	tests := []struct {
		name    string
		org     Organization
		expects error
	}{
		{
			name:    "Empty organization name",
			org:     Organization{},
			expects: ErrEmptyOrganizationName,
		},
		{
			name:    "Valid organization name",
			org:     Organization{Name: "valid-org", SSHKeyPath: privateKey},
			expects: nil,
		},
		{
			name:    "Invalid organization name",
			org:     Organization{Name: "Invalid!Org", SSHKeyPath: privateKey},
			expects: ErrInvalidOrgName,
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
