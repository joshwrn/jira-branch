# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Jira Branch is a Go CLI tool that integrates with Jira to help developers create Git branches based on their assigned Jira tickets. It uses Bubble Tea (Charm) for the terminal UI and provides an interactive interface for selecting tickets and creating appropriately named branches.

## Architecture

### Core Components

- **main.go**: Entry point that loads environment variables and initializes the application
- **internal/app/**: Contains the main Bubble Tea application model and UI logic
  - Uses Model-View-Update (MVU) pattern with different views: credentials, list, search, form
  - Manages application state including authentication, ticket loading, and branch creation
- **internal/jira/**: Jira API integration for authentication, ticket fetching, and status updates
  - Handles credential storage via OS keyring (using zalando/go-keyring)
  - Implements JQL queries to fetch user's assigned, non-done tickets
- **internal/git_utils/**: Git operations for branch creation and checkout
  - Formats branch names based on ticket type (feature/ or bugfix/ prefix)
  - Handles both new branch creation and existing branch checkout
- **internal/gui/**: Shared UI components and styling using Lipgloss
- **internal/utils/**: Utilities for logging (zerolog), file operations, and configuration

### Key Features

- Interactive ticket selection with search functionality
- Automatic branch naming: `feature/TICKET-123-description` or `bugfix/TICKET-123-description`
- Credential management via OS keyring
- Project-specific filtering via `jira-branch.config.json`
- Option to mark tickets as "In Progress" when creating branches

## Development Commands

### Prerequisites
- Go 1.21+ (currently using Go 1.24.4)
- Jira account with API token

### Setup
```bash
go mod tidy
```

### Running
```bash
go run main.go
```

### Building
For cross-platform builds:
```powershell
.\build-all.ps1
```

For single platform:
```bash
go build -o jira-branch
```

### Environment Configuration
Create a `.env` file for development:
```
JIRA_API_TOKEN=""
JIRA_USERNAME=""
JIRA_URL=""
DEV="true"  # Enables info level logs
```

## Configuration

The tool uses `jira-branch.config.json` in the project root to filter tickets by project:
```json
{
  "projectKey": "YOUR_PROJECT_KEY"
}
```

## Dependencies

Key external dependencies:
- **Bubble Tea**: Terminal UI framework
- **Lipgloss**: Styling and layout
- **Bubbles**: Pre-built UI components (tables, spinners, text inputs)
- **Huh**: Forms and prompts
- **go-keyring**: Secure credential storage
- **zerolog**: Structured logging
- **godotenv**: Environment variable loading

## Testing

No test files are currently present in the codebase. When adding tests, follow Go conventions with `*_test.go` files.