# Jira Branch

A CLI tool that allows you to create a branch using the name of one of your latest Jira issues.

![demo](https://github.com/user-attachments/assets/14372f65-b674-4022-b641-f35ebae24f2d)

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


