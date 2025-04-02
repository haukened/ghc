package configfile

import (
	"os"
	"path/filepath"
	"testing"

	"ghc/internal/domain"
)

func TestLoadConfig_FileNotFound(t *testing.T) {
	SetDefaultConfigPath("/nonexistent/path/to/config.json")

	_, err := LoadConfig()
	if err == nil || err != ErrConfigNotFound {
		t.Errorf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "invalid_config_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString("invalid json"); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	SetDefaultConfigPath(tempFile.Name())

	_, err = LoadConfig()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestLoadConfig_KoanfLoadError(t *testing.T) {
	// Set an invalid config path to simulate a koanf load error
	SetDefaultConfigPath("/invalid/path/to/config.json")

	_, err := LoadConfig()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestWriteConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	SetDefaultConfigPath(configPath)

	cfg := &domain.Config{
		Organizations: []*domain.Organization{
			{Name: "org1", SSHKeyPath: "/path/to/key"},
		},
	}

	if err := WriteConfig(cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("expected config file to exist, but it does not")
	}
}

func TestLoadConfig_Success(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	SetDefaultConfigPath(configPath)

	cfg := &domain.Config{
		Organizations: []*domain.Organization{
			{Name: "org1", SSHKeyPath: "/path/to/key"},
		},
	}

	if err := WriteConfig(cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(loadedCfg.Organizations) != len(cfg.Organizations) || loadedCfg.Organizations[0].Name != cfg.Organizations[0].Name {
		t.Errorf("loaded config does not match written config")
	}
}

func TestWriteConfig_MkdirAllError(t *testing.T) {
	// Set an invalid directory path to simulate MkdirAll error
	SetDefaultConfigPath("/invalid/path/to/config.json")

	cfg := &domain.Config{
		Organizations: []*domain.Organization{
			{Name: "org1", SSHKeyPath: "/path/to/key"},
		},
	}

	err := WriteConfig(cfg)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestWriteConfig_OpenFileError(t *testing.T) {
	// Set a directory path instead of a file path to simulate OpenFile error
	tempDir := t.TempDir()
	SetDefaultConfigPath(tempDir)

	cfg := &domain.Config{
		Organizations: []*domain.Organization{
			{Name: "org1", SSHKeyPath: "/path/to/key"},
		},
	}

	err := WriteConfig(cfg)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestWriteConfig_EncoderError(t *testing.T) {
	// Use a read-only file to simulate an encoder error
	tempFile, err := os.CreateTemp("", "readonly_config_*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	tempFile.Close()
	if err := os.Chmod(tempFile.Name(), 0400); err != nil {
		t.Fatalf("failed to set file permissions: %v", err)
	}

	SetDefaultConfigPath(tempFile.Name())

	cfg := &domain.Config{
		Organizations: []*domain.Organization{
			{Name: "org1", SSHKeyPath: "/path/to/key"},
		},
	}

	err = WriteConfig(cfg)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
