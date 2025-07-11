package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
			return ticketsMsg{
				tickets:                   tickets,
				err:                       err,
				shouldOverwriteAllTickets: true,
			}
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
		if m.showSearch {
			return updateSearch(m, msg)
		} else {
			return updateList(m, msg)
		}
	case "credentials":
		return updateCredentials(m, msg)
	case "form":
		return updateForm(m, msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.view == "credentials" {
		return viewCredentials(m)
	}

	if m.err != nil {
		b := strings.Builder{}
		bw := b.WriteString
		bw(gui.ErrorText.Render(fmt.Sprintf("❌ %v", m.err)))
		bw("\n\n")
		bw(gui.CreateHelpItems([]gui.HelpItem{
			{Key: "q/ctrl+c", Desc: "Quit"},
		}))
		return b.String()
	}

	if !m.isLoggedIn && m.isLoading {
		return gui.CreateLoadingView(&gui.LoadingView{
			Text:    "Validating credentials...",
			Width:   m.width,
			Height:  m.height,
			Spinner: m.spinner,
		})
	}

	if m.isLoading {
		return gui.CreateLoadingView(&gui.LoadingView{
			Text:    "Loading Jira tickets...",
			Width:   m.width,
			Height:  m.height,
			Spinner: m.spinner,
		})
	}

	if m.view == "form" {
		return viewForm(m)
	}

	return viewList(m)
}

func Run() {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	m := model{
		list:             table.New(),
		spinner:          s,
		isLoading:        true,
		isLoggedIn:       false,
		view:             "list",
		tickets:          []jira.JiraTicketsMsg{},
		credentialInputs: []textinput.Model{},
		currentField:     0,
		showSearch:       false,
		searchInput:      textinput.New(),
		search:           "",
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
