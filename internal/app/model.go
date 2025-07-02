package app

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/joshwrn/jira-branch/internal/jira"
)

type model struct {
	table      table.Model
	isLoading  bool
	isLoggedIn bool
	spinner    spinner.Model
	err        error
	width      int
	height     int
	view       string
	input      textinput.Model
	tickets    []jira.JiraTicketsMsg

	// Credential input fields
	credentialInputs []textinput.Model
	currentField     int
	credentials      jira.Credentials
}
