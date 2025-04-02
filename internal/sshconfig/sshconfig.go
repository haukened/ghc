package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// createSSHConfigFile creates an SSH config file with a single host entry.
// The file is created in a temporary directory and is always named "config".
// Parameters:
// - sshKeyPath: The path to the SSH key file.
// - configDir: The directory where the SSH config file will be created.
// Returns the path to the created SSH config file.
func CreateSSHConfigFile(sshHostName, sshKeyPath, configDir string) (string, error) {
	// create the file content
	sshConfig := fmt.Sprintf(`Host %s
	User git
	IdentityFile %s`, sshHostName, sshKeyPath)

	// Check if the SSH key path ends with ".pub"
	// If it does, add the "IdentitiesOnly yes" line
	if strings.HasSuffix(sshKeyPath, ".pub") {
		sshConfig += "\n\tIdentitiesOnly yes\n"
	}

	// create the file path
	sshConfigFilePath := filepath.Join(configDir, generateUUID())

	// create the file
	err := os.WriteFile(sshConfigFilePath, []byte(sshConfig), 0600)

	return sshConfigFilePath, err
}

// extract the UUID generation logic into a variable
// This allows for easier testing and mocking of UUID generation.
var generateUUID = func() string {
	return uuid.New().String()
}
