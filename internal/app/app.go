package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/joshwrn/jira-branch/internal/git_utils"
	"github.com/joshwrn/jira-branch/internal/gui"
	"github.com/joshwrn/jira-branch/internal/jira"

	"github.com/charmbracelet/lipgloss"
)

type errMsg error

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		textinput.Blink,
		func() tea.Msg {
			credentials, err := jira.LoadCredentials()
			if err != nil {
				return credentialsNeededMsg{}
			}
			if err := jira.ValidateCredentials(credentials); err != nil {
				return credentialsNeededMsg{}
			}
			return credentials
		},
	)
}

type credentialsNeededMsg struct{}

type ticketsMsg struct {
	tickets []jira.JiraTicketsMsg
	err     error
}

func fetchTickets(credentials jira.Credentials) tea.Cmd {
	return func() tea.Msg {
		tickets, err := jira.GetJiraTickets(credentials)
		return ticketsMsg{tickets: tickets, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.view {
	case "list":
		m, cmd, shouldReturn := updateList(m, msg)
		if shouldReturn {
			return m, cmd
		}
	case "credentials":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			m, cmd, shouldReturn := updateCredentials(m, msg)
			if shouldReturn {
				return m, cmd
			}
		}
	case "input":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				branchName := m.input.Value()
				return m, git_utils.CheckoutBranch(branchName)
			case "esc":
				m.view = "list"
				return m, nil
			}
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	// global messages
	switch msg := msg.(type) {
	case jira.Credentials:
		m.credentials = msg
		m.isLoggedIn = true
		m.isLoading = true
		m.view = "list"
		return m, fetchTickets(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.isLoading && m.err == nil && m.isLoggedIn {
			m.updateTableSize()
		}

	case tea.KeyMsg:
		if m.isLoggedIn && msg.String() == "q" {
			return m, nil
		}
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		m.isLoading = false
		return m, nil

	case spinner.TickMsg:
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if !m.isLoading && m.isLoggedIn && m.view == "list" {
		m.table, cmd = m.table.Update(msg)
	} else if m.view == "credentials" && len(m.credentialInputs) > 0 {
		m.credentialInputs[m.currentField], cmd = m.credentialInputs[m.currentField].Update(msg)
	} else if m.view == "input" {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.view == "credentials" {
		return viewCredentials(m)
	}

	if !m.isLoggedIn && m.isLoading {
		line1 := fmt.Sprintf(
			"%s Validating credentials...",
			m.spinner.View(),
		)

		line2 := gui.FaintWhiteText.
			Render("Press ctrl+c to quit")

		return fmt.Sprintf(
			"\n%s\n\n%s",
			line1,
			line2,
		)
	}

	if m.err != nil {
		errorText := fmt.Sprintf("Error: %v", m.err)
		helpText := "Press 'r' to retry or 'q' to quit"

		if m.width > 0 {
			errorPadding := max(0, (m.width-len(errorText))/2)
			helpPadding := max(0, (m.width-len(helpText))/2)
			errorText = strings.Repeat(" ", errorPadding) + errorText
			helpText = strings.Repeat(" ", helpPadding) + helpText
		}

		return fmt.Sprintf("\n%s\n\n%s\n", errorText, helpText)
	}

	if m.isLoading {
		loadingText := fmt.Sprintf("%s %s", m.spinner.View(), "Loading Jira tickets...")

		return fmt.Sprintf("\n%s\n\nPress q to quit", loadingText)
	}

	if m.view == "input" {
		faintStyle := gui.FaintWhiteText
		return fmt.Sprintf(
			"%s\n\nenter %s %s esc %s %s q/ctrl+c %s",
			m.input.View(),
			faintStyle.Render("checkout branch"),
			faintStyle.Render("•"),
			faintStyle.Render("go back"),
			faintStyle.Render("•"),
			faintStyle.Render("quit"),
		)
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Render(m.table.View())
}

func Run() {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	m := model{
		table:            table.New(),
		spinner:          s,
		isLoading:        true,
		isLoggedIn:       false,
		view:             "list",
		input:            textinput.New(),
		tickets:          []jira.JiraTicketsMsg{},
		credentialInputs: []textinput.Model{},
		currentField:     0,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
