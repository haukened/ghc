package clone

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"ghc/internal/configfile"
	"ghc/internal/sshconfig"
	"ghc/internal/utils"

	"github.com/urfave/cli/v3"
)

var (
	ErrInvalidArgs          = errors.New("exactly one argument is required")
	ErrEmptyRepoURL         = errors.New("repository URL is required")
	ErrInvalidRepoURLFormat = errors.New("invalid GitHub SSH URL format")
	ErrOrgNameNotFound      = errors.New("organization name not found in the URL")
)

// You can override this variable at build time using -ldflags:
// go build -ldflags="-X 'ghc/internal/clone.sshHostName=github.mycompany.com'" ./cmd/ghc
//
// Note: the package path in -X must match the actual package where the variable is defined (here: main)
var sshHostName = "github.com"

// This can also be overridden at build time using -ldflags:
// go build -ldflags="-X 'ghc/internal/clone.defaultSSHConfigPath=/custom/path'" ./cmd/ghc
var defaultSSHConfigPath = "$HOME/.config/ghc/ssh_configs/"

// cloneRepo clones a Git repository using the provided context and command.
// It validates the repository URL, retrieves the SSH key for the organization,
// creates the necessary SSH config file, and then runs the clone command.
func CloneRepo(ctx context.Context, c *cli.Command) error {
	// Step 0: Check nargs and args
	if c.NArg() != 1 {
		return fmt.Errorf("cloneRepo: %w", ErrInvalidArgs)
	}

	repoURL := c.Args().First()
	if repoURL == "" {
		return fmt.Errorf("cloneRepo: %w", ErrEmptyRepoURL)
	}

	// Step 1: Parse the repository URL
	orgName, err := parseGitSSHRepoUrl(repoURL)
	if err != nil {
		return fmt.Errorf("cloneRepo: %w", err)
	}
	if orgName == "" {
		return fmt.Errorf("cloneRepo: %w", ErrOrgNameNotFound)
	}

	// Step 2: Get the SSH key for that organization
	config, err := configfile.LoadConfig()
	if err != nil {
		return fmt.Errorf("cloneRepo: %w", err)
	}

	// Returns the SSH key path for the organization
	sshKeyPath, err := config.GetKeyPathForOrg(orgName)
	if err != nil {
		return fmt.Errorf("cloneRepo: %w", err)
	}

	// Step 3: Resolve the ghc config path
	expandedSSHConfigPath := utils.ExpandPath(defaultSSHConfigPath)

	// Step 4: Ensure the SSH config directory exists
	err = os.MkdirAll(expandedSSHConfigPath, 0700)
	if err != nil {
		return fmt.Errorf("cloneRepo: %w", err)
	}

	// Step 5: Create the SSH config file
	configPath, err := sshconfig.CreateSSHConfigFile(sshHostName, sshKeyPath, expandedSSHConfigPath)
	if err != nil {
		return fmt.Errorf("cloneRepo: %w", err)
	}

	// Step 6: Clone the repository using the SSH config file
	runner := &defaultRunner{}
	return cloneRepoUsingConfigFile(configPath, repoURL, runner)
}

// returns the GitHub User/Org and an error if it's not a GitHub SSH URL
func parseGitSSHRepoUrl(url string) (string, error) {
	// format = git@github.com:haukened/ghc.git
	// quote the metacharacters in the SSH host name
	pattern := fmt.Sprintf(`^git@%s:([^/]+)/[^/]+(?:\.git)?$`, regexp.QuoteMeta(sshHostName))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(url)
	if len(matches) != 2 {
		return "", ErrInvalidRepoURLFormat
	}
	return matches[1], nil
}

// buildCloneCommand constructs an exec.Cmd to clone a Git repository using a custom SSH config file.
func buildCloneCommand(configPath, cloneURI string) *exec.Cmd {
	return exec.Command("git", "clone", "--config", fmt.Sprintf("core.sshCommand=ssh -F %s", configPath), cloneURI)
}

// cloneRepoUsingConfigFile validates the SSH config and clone URL, and runs the Git clone command using the provided CommandRunner.
// It returns an error if validation fails or the clone command fails to run.
func cloneRepoUsingConfigFile(configPath, cloneURI string, runner CommandRunner) error {
	if !fileExists(configPath) {
		return fmt.Errorf("%w: ssh config file %s does not exist", os.ErrNotExist, configPath)
	}

	validGitSSH := regexp.MustCompile(`^git@[^:]+:[^/]+/[^/]+(?:\.git)?$`)
	if !validGitSSH.MatchString(cloneURI) {
		return ErrInvalidRepoURLFormat
	}

	cmd := buildCloneCommand(configPath, cloneURI)
	return runner.Run(cmd)
}

// CommandRunner is an interface that defines how a command should be executed.
// This is useful for testing to avoid running real system commands.
type CommandRunner interface {
	Run(cmd *exec.Cmd) error
}

type defaultRunner struct{}

// Run executes the given command, streaming its output to stdout and stderr.
func (r *defaultRunner) Run(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var fileExists = func(path string) bool {
	// fileExists checks whether the specified file path exists on the filesystem.
	// This function can be overridden in tests.
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
