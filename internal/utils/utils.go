// Package utils provides utility functions for common operations.

package utils

import (
	"os"
	"strings"
)

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
