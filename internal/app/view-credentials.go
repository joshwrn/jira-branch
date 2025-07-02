package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshwrn/jira-branch/internal/gui"
)

func CreateBaseCredentialInput(width int) textinput.Model {
	ti := textinput.New()
	ti.CharLimit = 200
	ti.Width = width
	ti.PlaceholderStyle = gui.FaintWhiteText
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	return ti
}

func CreateCredentialInputs(width int) []textinput.Model {
	inputs := make([]textinput.Model, 3)

	jira_url := os.Getenv("JIRA_URL")
	jira_email := os.Getenv("JIRA_EMAIL")
	jira_api_token := os.Getenv("JIRA_API_TOKEN")

	inputs[0] = CreateBaseCredentialInput(width)
	inputs[0].Placeholder = "your-company.atlassian.net"
	if jira_url != "" {
		inputs[0].SetValue(jira_url)
	}
	inputs[0].Prompt = "Atlassian URL: "
	inputs[0].Focus()

	inputs[1] = CreateBaseCredentialInput(width)
	inputs[1].Placeholder = "your-email@company.com"
	if jira_email != "" {
		inputs[1].SetValue(jira_email)
	}
	inputs[1].Prompt = "Email: "

	inputs[2] = CreateBaseCredentialInput(width)
	inputs[2].Placeholder = "Your JIRA API token"
	if jira_api_token != "" {
		inputs[2].SetValue(jira_api_token)
	}
	inputs[2].Prompt = "API Token: "
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '•'

	return inputs
}

func viewCredentials(m model) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(lipgloss.
		NewStyle().
		Foreground(lipgloss.Color("7")).
		Render("Generate an API token at: "))
	b.WriteString(
		lipgloss.
			NewStyle().
			Foreground(lipgloss.Color("4")).
			Underline(true).
			Render("https://id.atlassian.com/manage-profile/security/api-tokens"))

	b.WriteString("\n\n")

	b.WriteString(gui.FaintWhiteText.
		Render(`If you choose "API token with scopes" give it the "read:jira-work" scope.`))

	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(gui.ErrorText.Render(fmt.Sprintf("❌ %v", m.err)))
		b.WriteString("\n\n")
	}

	b.WriteString(m.credentialInputs[0].View())
	b.WriteString("\n")
	b.WriteString(m.credentialInputs[1].View())
	b.WriteString("\n")
	b.WriteString(m.credentialInputs[2].View())
	b.WriteString("\n\n")

	b.WriteString(gui.CreateHelpItems([]gui.HelpItem{
		{Key: "tab", Desc: "Navigate"},
		{Key: "enter", Desc: "Submit"},
		{Key: "ctrl+c", Desc: "Quit"},
	}))

	return b.String()
}
