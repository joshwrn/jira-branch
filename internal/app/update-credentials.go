package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshwrn/jira-branch/internal/jira"
)

type credentialsNeededMsg struct{}

func validateAndStoreCredentials(credentials jira.Credentials) tea.Cmd {
	return func() tea.Msg {
		if err := jira.ValidateCredentials(credentials); err != nil {
			return errMsg(err)
		}

		if err := jira.StoreCredentials(credentials); err != nil {
			return errMsg(fmt.Errorf("failed to store credentials: %v", err))
		}

		return credentials
	}
}

func updateCredentials(m model, msg tea.Msg) (model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.currentField == len(m.credentialInputs)-1 {
				jiraURL := strings.TrimSpace(m.credentialInputs[0].Value())
				if !strings.HasPrefix(jiraURL, "https://") {
					jiraURL = "https://" + jiraURL
				}
				jiraURL = strings.TrimSuffix(jiraURL, "/")

				m.credentials = jira.Credentials{
					JiraURL:  jiraURL,
					Email:    strings.TrimSpace(m.credentialInputs[1].Value()),
					APIToken: strings.TrimSpace(m.credentialInputs[2].Value()),
				}

				if m.credentials.JiraURL == "" || m.credentials.Email == "" || m.credentials.APIToken == "" {
					m.err = errMsg(fmt.Errorf("all fields are required"))
					return m, nil, false
				}

				m.isLoading = true
				m.err = nil
				return m, validateAndStoreCredentials(m.credentials), true
			}

			if s == "up" || s == "shift+tab" {
				m.currentField--
			} else {
				m.currentField++
			}

			if m.currentField > len(m.credentialInputs)-1 {
				m.currentField = 0
			} else if m.currentField < 0 {
				m.currentField = len(m.credentialInputs) - 1
			}

			for i := 0; i < len(m.credentialInputs); i++ {
				if i == m.currentField {
					m.credentialInputs[i].Focus()
				} else {
					m.credentialInputs[i].Blur()
				}
			}

			return m, textinput.Blink, true
		}
	}

	return m, nil, false
}
