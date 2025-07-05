package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// global messages
	switch msg := msg.(type) {
	case jira.Credentials:
		m.credentials = msg
		m.isLoggedIn = true
		m.isLoading = true
		m.view = "list"
		return m, func() tea.Msg {
			tickets, err := jira.GetJiraTickets(msg)
			return ticketsMsg{tickets: tickets, err: err}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.isLoading && m.err == nil && m.isLoggedIn {
			m.updateTableSize()
		}

	case tea.KeyMsg:
		if m.isLoggedIn && msg.String() == "q" && m.view == "list" {
			return m, tea.Quit
		}
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.QuitMsg:
		if m.isSubmittingForm {
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		m.isLoading = false
		m.isSubmittingForm = false
		return m, nil

	case spinner.TickMsg:
		if m.isLoading || m.isSubmittingForm {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	}

	// view specific messages
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
	case "form":
		if m.isSubmittingForm {
			return m, nil
		}
		form, formCmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			if m.form.State == huh.StateCompleted {
				m.isSubmittingForm = true
				return m, tea.Batch(
					m.spinner.Tick,
					func() tea.Msg {
						if *m.formShouldMarkAsInProgress {
							err := jira.MarkAsInProgress(
								m.credentials,
								m.table.SelectedRow()[0],
							)
							if err != nil {
								return errMsg(err)
							}
						}
						checkCmd := git_utils.CheckoutBranch(*m.formBranchName)
						return checkCmd()
					},
				)
			}
			return m, formCmd
		}
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			m.view = "list"
			return m, nil
		}
		return m, cmd
	}

	// view specific updates
	if !m.isLoading && m.isLoggedIn && m.view == "list" {
		m.table, cmd = m.table.Update(msg)
	} else if m.view == "credentials" && len(m.credentialInputs) > 0 {
		m.credentialInputs[m.currentField], cmd = m.credentialInputs[m.currentField].Update(msg)
	} else if m.view == "form" {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.view == "credentials" {
		return viewCredentials(m)
	}

	if m.err != nil {
		b := strings.Builder{}
		b.WriteString(gui.ErrorText.Render(fmt.Sprintf("❌ %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(gui.CreateHelpItems([]gui.HelpItem{
			{Key: "q/ctrl+c", Desc: "Quit"},
		}))

		return b.String()
	}

	if !m.isLoggedIn && m.isLoading {
		b := strings.Builder{}
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString("Validating credentials...")
		b.WriteString("\n\n")
		b.WriteString(gui.CreateHelpItems([]gui.HelpItem{
			{Key: "q/ctrl+c", Desc: "Quit"},
		}))

		return b.String()
	}

	if m.isLoading {
		b := strings.Builder{}
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString("Loading Jira tickets...")
		b.WriteString("\n\n")
		b.WriteString(gui.CreateHelpItems([]gui.HelpItem{
			{Key: "q/ctrl+c", Desc: "Quit"},
		}))

		return b.String()
	}

	if m.view == "form" {
		if m.isSubmittingForm {
			b := strings.Builder{}
			b.WriteString(m.spinner.View())
			b.WriteString(" ")
			b.WriteString("Creating branch and updating Jira...")
			b.WriteString("\n\n")
			b.WriteString(gui.CreateHelpItems([]gui.HelpItem{
				{Key: "q/ctrl+c", Desc: "Quit"},
			}))
			return b.String()
		}
		return m.form.View()
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
