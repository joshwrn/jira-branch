package app

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshwrn/jira-branch/internal/git_utils"
	"github.com/joshwrn/jira-branch/internal/gui"
	"github.com/joshwrn/jira-branch/internal/utils"
)

func updateList(m model, msg tea.Msg) (model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.view == "list" && len(m.tickets) > 0 {
				selectedRow := m.table.Cursor()
				if selectedRow < len(m.tickets) {
					selectedTicket := m.tickets[selectedRow]
					selected_branch := git_utils.FormatBranchName(selectedTicket)
					m.view = "input"
					m.input = gui.CreateBranchInput(selected_branch, m.width)
					return m, textinput.Blink, true
				}
			}
		}
	case credentialsNeededMsg:
		m.view = "credentials"
		m.credentialInputs = CreateCredentialInputs(m.width)
		m.currentField = 0
		return m, textinput.Blink, true
	case ticketsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.isLoading = false
			return m, nil, true
		}

		m.tickets = msg.tickets

		columns := []table.Column{
			{Title: "Key", Width: 0},
			{Title: "Type", Width: 0},
			{Title: "Summary", Width: 0},
			{Title: "Status", Width: 0},
			{Title: "Created", Width: 0},
		}

		rows := []table.Row{}
		for _, ticket := range m.tickets {
			rows = append(rows, table.Row{ticket.Key, ticket.Type, ticket.Summary, ticket.Status, utils.FormatRelativeTime(ticket.Created)})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			BorderBottom(true).
			Bold(false)
		s.Selected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Background(lipgloss.Color("0")).
			Bold(false)

		t.SetStyles(s)

		m.table = t
		m.isLoading = false
		m.isLoggedIn = true
		m.updateTableSize()
		return m, nil, true
	}

	return m, nil, false
}
