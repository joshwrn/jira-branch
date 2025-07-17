# Jira Branch

A CLI tool that allows you to create a branch using the name of one of your latest Jira issues.

![demo](https://github.com/user-attachments/assets/14372f65-b674-4022-b641-f35ebae24f2d)

---

## Installation

**Windows (PowerShell):**
```powershell
iwr https://raw.githubusercontent.com/joshwrn/jira-branch/main/install.ps1 | iex
```

**macOS/Linux (Bash):**
```bash
curl -sSL https://raw.githubusercontent.com/joshwrn/jira-branch/main/install.sh | bash
```

The installer will:
- Detect your OS and architecture automatically
- Download the latest release
- Install to `~/.jira-branch/bin/`
- Add to your PATH (you may need to restart your terminal)
- Optionally, add the alias `jb` to your shell config

### Updating

Run the installation script again to update to the latest version.

---

## Usage

You can use the alias `jb` or the full command `jira-branch`.

```
jb
# or
jira-branch
```

### Setting up your Jira API token

The first time you run `jb`, you'll be prompted to sign in to Jira.

First, enter your Atlassian URL. This will probably be something like `your-company.atlassian.net`.

Then, enter the email address you use to sign in to Jira.

Finally, enter your Jira API token. You can create one in your Atlassian account settings.

> [!WARNING]
> Make sure you do NOT choose the "API token with scopes" option.

[Create an API token](https://id.atlassian.com/manage-profile/security/api-tokens)

---

## Configuration

To only show issues from a specific project, add a file called `jira-branch.config.json` in the root of your project.

If the issues in your project are something like `PRJ-123`, then the project key is `PRJ`.

So, the `jira-branch.config.json` file should look something like this:

```json
{
  "projectKey": "PRJ"
}
```

---

## Development

### Prerequisites

- Go 1.21 or higher
- A Jira account and API token

### Setup

1. Clone the repository
2. Run `go mod tidy`
3. Run `go run main.go`
4. Add the following to your `.env` file:

```
# Makes it easier to test signing in to Jira
JIRA_API_TOKEN=""
JIRA_USERNAME=""
JIRA_URL=""

# Enables info level logs
DEV="true"
```


