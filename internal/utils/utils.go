// Package utils provides utility functions for common operations.

package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func GenerateTestSSHKey(t *testing.T) (privateKeyPath, publicKeyPath string) {
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

// ExpandPath replaces the tilde (~) in a file path with the user's home directory
// and expands any environment variables present in the path.
//
// Parameters:
//   - path: A string representing the file path to be expanded.
//
// Returns:
//   - A string with the tilde replaced by the user's home directory and
//     any environment variables expanded.
func ExpandPath(path string) string {
	// first replace any instance of ~ with the $HOME directory
	path = strings.Replace(path, "~", "$HOME", -1)
	// then expand the path
	return os.ExpandEnv(path)
}
