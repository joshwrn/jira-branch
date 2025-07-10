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
			m, cmd, shouldReturn := updateSearch(m, msg)
			if shouldReturn {
				return m, cmd
			}
		} else {
			m, cmd, shouldReturn := updateList(m, msg)
			if shouldReturn {
				return m, cmd
			}
		}
		if !m.isLoading && m.isLoggedIn {
			if m.showSearch {
				m.searchInput, cmd = m.searchInput.Update(msg)
				filterTickets(&m)
			} else {
				m.table, cmd = m.table.Update(msg)
			}
		}

	case "credentials":
		switch msg := msg.(type) {
		case tea.KeyMsg:
			m, cmd, shouldReturn := updateCredentials(m, msg)
			if shouldReturn {
				return m, cmd
			}
		}
		m.credentialInputs[m.currentField], cmd = m.credentialInputs[m.currentField].Update(msg)

	case "form":
		if m.isSubmittingForm {
			return m, nil
		}
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.view = "list"
				return m, nil
			}
		}
		form, formCmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			if m.form.State != huh.StateCompleted {
				return m, formCmd
			}
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
		bw := b.WriteString
		bw(gui.ErrorText.Render(fmt.Sprintf("❌ %v", m.err)))
		bw("\n\n")
		bw(gui.CreateHelpItems([]gui.HelpItem{
			{Key: "q/ctrl+c", Desc: "Quit"},
		}))

		return b.String()
	}

	if !m.isLoggedIn && m.isLoading {
		b := strings.Builder{}
		bw := b.WriteString
		bw(m.spinner.View())
		bw(" ")
		bw("Validating credentials...")
		bw("\n\n")
		bw(gui.CreateHelpItems([]gui.HelpItem{
			{Key: "q/ctrl+c", Desc: "Quit"},
		}))

		return b.String()
	}

	if m.isLoading {
		b := strings.Builder{}
		bw := b.WriteString
		bw(m.spinner.View())
		bw(" ")
		bw("Loading Jira tickets...")
		bw("\n\n")
		b.WriteString(gui.CreateHelpItems([]gui.HelpItem{
			{Key: "q/ctrl+c", Desc: "Quit"},
		}))

		return b.String()
	}

	if m.view == "form" {
		if m.isSubmittingForm {
			b := strings.Builder{}
			bw := b.WriteString
			bw(m.spinner.View())
			bw(" ")
			bw("Creating branch and updating Jira...")
			bw("\n\n")
			b.WriteString(gui.CreateHelpItems([]gui.HelpItem{
				{Key: "q/ctrl+c", Desc: "Quit"},
			}))
			return b.String()
		}

		sidebar := createSidebar(&m)

		formWidth := m.width - m.width/4 - 5
		formView := lipgloss.NewStyle().Width(formWidth).Render(m.form.View())

		return lipgloss.JoinHorizontal(lipgloss.Center, formView, sidebar)
	}

	helper := gui.CreateHelpItems([]gui.HelpItem{
		{Key: "enter", Desc: "Select ticket"},
		{Key: "/", Desc: "Search"},
		{Key: "r", Desc: "Refresh"},
		{Key: "S", Desc: "Sign out"},
		{Key: "q/ctrl+c", Desc: "Quit"},
	})

	b := strings.Builder{}
	bw := b.WriteString

	if m.showSearch {
		bw(m.searchInput.View())
		bw("\n")
		helper = gui.CreateHelpItems([]gui.HelpItem{
			{Key: "esc", Desc: "Clear"},
			{Key: "enter", Desc: "Confirm"},
		})
	}

	if m.search != "" && !m.showSearch {
		helperWidth := lipgloss.Width(helper)
		availableWidth := m.width - helperWidth
		searchText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Render(fmt.Sprintf("/%s", m.search))

		helper = helper +
			lipgloss.NewStyle().
				Width(availableWidth-1).
				Align(lipgloss.Right).
				Render(searchText)
	}

	return b.String() + lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Render(m.table.View()) + "\n" + helper
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
