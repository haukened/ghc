// Package configfile provides functionality for managing configuration files
// for the GHC CLI application. It includes loading, writing, and managing
// organization-specific SSH key configurations.
package configfile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/knadh/koanf"
	kjson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"

	"ghc/internal/domain"
	"ghc/internal/utils"
)

// DefaultConfigPath is the default path to the configuration file.
const DefaultConfigPath = "$HOME/.config/ghc/ghc.conf"

var (
	ErrConfigNotFound  = errors.New("config file not found")
	ErrHomeDirNotFound = errors.New("home directory not found")
)

// defaultConfigPath is the active path to the configuration file.
var defaultConfigPath = DefaultConfigPath

// LoadConfig loads the configuration from the default path.
// It returns the configuration or an error if the file is not found or invalid.
func LoadConfig() (*domain.Config, error) {
	if !homeDirExists() {
		return nil, ErrHomeDirNotFound
	}

	// Expand the default config path to the user's home directory
	configPath := utils.ExpandPath(defaultConfigPath)

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, ErrConfigNotFound
	}

	k := koanf.New(".")
	if err := k.Load(file.Provider(configPath), kjson.Parser()); err != nil {
		return nil, err
	}

	var cfg domain.Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// WriteConfig writes the provided configuration to the default config path.
// It creates the necessary directories if they do not exist.
func WriteConfig(cfg *domain.Config) error {
	if !homeDirExists() {
		return ErrHomeDirNotFound
	}
	// Expand the default config path to the user's home directory
	configPath := utils.ExpandPath(defaultConfigPath)

	// ensure the config directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Open the config file for writing
	file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode the config to JSON and write it to the file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	return nil
}

// SetDefaultConfigPath sets the default configuration path for testing purposes.
func SetDefaultConfigPath(path string) {
	defaultConfigPath = path
}

func homeDirExists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	return homeDir != ""
}
