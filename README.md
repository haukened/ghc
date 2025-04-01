# ghc

`ghc` is a command-line utility designed to help developers seamlessly switch between multiple GitHub accounts, such as personal and enterprise accounts, especially when dealing with conflicting SSH keys. This tool simplifies the management of SSH keys and GitHub organization configurations, ensuring a smooth workflow.

## Why?

Have you ever had a `~/.ssh/config` file that looked like this?

```
Host github-work
    Hostname github.com
    User git
    PreferredAuthentications publickey
    IdentityFile ~/.ssh/your-work-key

Host github-personal
    Hostname github.com
    User git
    PreferredAuthentications publickey
    IdentityFile ~/.ssh/your-regular-key
```

and then had to manually edit every clone command that you get from GitHub to have `github-work` or `github-personal` ?


## Purpose
Managing multiple GitHub accounts can be challenging, particularly when SSH keys conflict or when switching between accounts frequently. `ghc` addresses this problem by providing a straightforward way to configure and manage GitHub organizations and their associated SSH keys. With `ghc`, you can:

- Configure SSH keys for different GitHub organizations.
- Set a default organization for streamlined operations.
- Remove or list organizations as needed.

## Organization Commands
The following commands are available for managing GitHub organizations:

### `organization set` | `org set`
Sets the SSH key for a specified organization. If the `--default` flag is provided, the organization is marked as the default.

**Usage:**
```bash
ghc org set <organization_name> <ssh_key_path> [--default]
```

**Example:**
```bash
# Set the SSH key for the "my-org" organization
ghc org set my-org ~/.ssh/my_org_key

# Set the SSH key for your personal stuff and mark it as default
ghc org set GITHUB_USERNAME ~/.ssh/my_personal_key --default
```

### `organization remove` | `org rm`
Removes a specified organization from the configuration.

**Usage:**
```bash
ghc org rm <organization_name>
```

**Example:**
```bash
# Remove the "my-org" organization
ghc org rm my-org
```

### `organization list` | `org ls`
Lists all configured organizations and their details.

**Usage:**
```bash
ghc org ls
```

**Example:**
```bash
# List all organizations
ghc org ls
```