package gui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

func CreateBranchInput(selected_branch string, width int) textinput.Model {
	ti := textinput.New()
	ti.Prompt = "Confirm or edit branch name: \n\n"
	ti.SetValue(selected_branch)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = width

	return ti
}

func CreateBaseCredentialInput(width int) textinput.Model {
	ti := textinput.New()
	ti.CharLimit = 200
	ti.Width = width
	ti.PlaceholderStyle = FaintWhiteText
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	return ti
}

func CreateCredentialInputs(width int) []textinput.Model {
	inputs := make([]textinput.Model, 3)

	inputs[0] = CreateBaseCredentialInput(width)
	inputs[0].Placeholder = "https://your-company.atlassian.net"
	inputs[0].Prompt = "Jira URL: "
	inputs[0].Focus()

	inputs[1] = CreateBaseCredentialInput(width)
	inputs[1].Placeholder = "your-email@company.com"
	inputs[1].Prompt = "Email: "

	inputs[2] = CreateBaseCredentialInput(width)
	inputs[2].Placeholder = "Your JIRA API token"
	inputs[2].Prompt = "API Token: "
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '•'

	return inputs
}

var FaintWhiteText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("7")).
	Faint(true)

var ErrorText = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9")).
	Bold(true)
