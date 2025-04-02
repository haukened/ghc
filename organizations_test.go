package main

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"ghc/internal/configfile"
	"ghc/internal/domain"
	"ghc/internal/utils"

	"github.com/urfave/cli/v3"
)

func TestOrganization(t *testing.T) {
	// test setup
	privateKey, _ := utils.GenerateTestSSHKey(t)
	tempDir := t.TempDir()

	// create an empty org
	setConfigPath := filepath.Join(tempDir, "config.json")

	// create a test config file with bad content
	badConfigPath := filepath.Join(tempDir, "bad_config.json")
	utils.WriteConfigFileForTest(t, badConfigPath, []byte("foo"))

	// create a test config for the remove tests
	rmConfigPath := filepath.Join(tempDir, "rm_config.json")
	rmConfig := &domain.Config{
		Organizations: []*domain.Organization{
			{Name: "org1", SSHKeyPath: "/path/to/key1", IsDefault: false},
			{Name: "org2", SSHKeyPath: "/path/to/key2", IsDefault: true},
		},
	}
	rmBytes, err := rmConfig.JSON()
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}
	utils.WriteConfigFileForTest(t, rmConfigPath, rmBytes)

	// create a test config for the list tests
	lsConfigPath := filepath.Join(tempDir, "ls_config.json")
	lsConfig := &domain.Config{
		Organizations: []*domain.Organization{
			{Name: "org1", SSHKeyPath: "/path/to/key1", IsDefault: false},
			{Name: "org2", SSHKeyPath: "/path/to/key2", IsDefault: true},
		},
	}
	lsBytes, err := lsConfig.JSON()
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}
	utils.WriteConfigFileForTest(t, lsConfigPath, lsBytes)

	// create an org that will remain empty
	emptyOrgConfigPath := filepath.Join(tempDir, "empty_org_config.json")
	emptyOrgConfig := &domain.Config{
		Organizations: []*domain.Organization{},
	}
	emptyOrgBytes, err := emptyOrgConfig.JSON()
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}
	utils.WriteConfigFileForTest(t, emptyOrgConfigPath, emptyOrgBytes)

	// mock urfave/cli/v3
	cmd := &cli.Command{
		Name: "organization",
		Commands: []*cli.Command{
			{
				Name:   "set",
				Action: setOrganization,
			},
			{
				Name:   "list",
				Action: listOrganizations,
			},
			{
				Name:   "remove",
				Action: removeOrganization,
			},
		},
	}

	// create a series of test cases
	tests := []struct {
		name         string
		configPath   string
		args         []string
		expected     error
		expectedType any
	}{
		{
			name:       "set valid args",
			configPath: setConfigPath,
			args:       []string{"org", "set", "org1", privateKey},
			expected:   nil,
		},
		{
			name:       "set bad nargs",
			configPath: setConfigPath,
			args:       []string{"org", "set", "org1"},
			expected:   ErrNumArguments,
		},
		{
			name:       "set empty key path",
			configPath: setConfigPath,
			args:       []string{"org", "set", "org1", " "},
			expected:   ErrNumArguments, // assuming empty key path is treated as invalid, the key path error will by wrapped in this
		},
		{
			name:       "set empty org name",
			configPath: setConfigPath,
			args:       []string{"org", "set", " ", privateKey},
			expected:   ErrNumArguments, // assuming empty org name is treated as invalid, the org name error will by wrapped in this
		},
		{
			name:       "set invalid org name",
			configPath: setConfigPath,
			args:       []string{"org", "set", "!-invalid!", privateKey},
			expected:   domain.ErrInvalidOrgName,
		},
		{
			name:         "set bad config file",
			configPath:   badConfigPath,
			args:         []string{"org", "set", "org1", privateKey},
			expectedType: &json.SyntaxError{},
		},
		{
			name:         "rm bad config file",
			configPath:   badConfigPath,
			args:         []string{"org", "remove", "org1"},
			expectedType: &json.SyntaxError{},
		},
		{
			name:         "ls bad config file",
			configPath:   badConfigPath,
			args:         []string{"org", "list"},
			expectedType: &json.SyntaxError{},
		},
		{
			name:       "remove default org",
			configPath: rmConfigPath,
			args:       []string{"org", "remove", "org2"},
			expected:   domain.ErrCantRemoveDefault,
		},
		{
			name:       "remove valid",
			configPath: rmConfigPath,
			args:       []string{"org", "remove", "org1"},
			expected:   nil,
		},
		{
			name:       "remove bad nargs",
			configPath: rmConfigPath,
			args:       []string{"org", "remove"},
			expected:   ErrNumArguments,
		},
		{
			name:       "list org valid",
			configPath: lsConfigPath,
			args:       []string{"org", "list"},
			expected:   nil,
		},
		{
			name:       "list no organizations",
			configPath: emptyOrgConfigPath,
			args:       []string{"org", "list"},
			expected:   domain.ErrNoOrganizations,
		},
	}

	// run the tests
	for _, test := range tests {
		t.Logf("running test: %s", test.name)
		configfile.SetDefaultConfigPath(test.configPath)
		err := cmd.Run(t.Context(), test.args)

		// use Errors.As for json.SyntaxError because it is a different type of error
		if test.expectedType != nil {
			if !errors.As(err, &test.expectedType) {
				t.Errorf("expected error of type %T, got %v", test.expectedType, err)
			}
			continue
		}

		// use errors.Is for other errors
		if !errors.Is(err, test.expected) {
			t.Errorf("expected %v, got %v", test.expected, err)
		}
	}
}
