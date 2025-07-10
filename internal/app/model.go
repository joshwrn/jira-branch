package app

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/huh"
	"github.com/joshwrn/jira-branch/internal/jira"
)

type model struct {
	isLoading  bool
	isLoggedIn bool
	err        error
	width      int
	height     int
	view       string

	spinner spinner.Model
	table   table.Model
	form    *huh.Form

	isSubmittingForm bool

	formBranchName             *string
	formShouldMarkAsInProgress *bool

	credentialInputs []textinput.Model
	currentField     int
	credentials      jira.Credentials

	allTickets []jira.JiraTicketsMsg
	tickets    []jira.JiraTicketsMsg

	showSearch  bool
	search      string
	searchInput textinput.Model
}
