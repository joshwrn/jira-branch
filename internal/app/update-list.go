package app

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshwrn/jira-branch/internal/git_utils"
	"github.com/joshwrn/jira-branch/internal/jira"
	"github.com/joshwrn/jira-branch/internal/utils"
	"github.com/zalando/go-keyring"
)

func (m *model) updateTableSize() {
	if m.width > 0 && m.height > 0 {
		keyWidth := 10
		typeWidth := 10
		statusWidth := 30
		createdWidth := 15
		summaryWidth := max(20, m.width-keyWidth-typeWidth-statusWidth-createdWidth-12)

		columns := []table.Column{
			{Title: "Key", Width: keyWidth},
			{Title: "Type", Width: typeWidth},
			{Title: "Summary", Width: summaryWidth},
			{Title: "Status", Width: statusWidth},
			{Title: "Created", Width: createdWidth},
		}
		m.table.SetColumns(columns)
		m.table.SetWidth(m.width - 2)
		m.table.SetHeight(m.height - 2)
	}
}

type ticketsMsg struct {
	tickets []jira.JiraTicketsMsg
	err     error
}

func updateList(m model, msg tea.Msg) (model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.isLoading = true
			return m, tea.Batch(func() tea.Msg {
				tickets, err := jira.GetJiraTickets(m.credentials)
				return ticketsMsg{tickets: tickets, err: err}
			}, m.spinner.Tick), true
		case "S":
			m.view = "credentials"
			m.credentialInputs = CreateCredentialInputs(m.width)
			m.currentField = 0
			m.isLoggedIn = false
			keyring.Delete("jira-cli", "credentials")
			return m, textinput.Blink, true
		case "enter":
			if m.view == "list" && len(m.tickets) > 0 {
				selectedRow := m.table.Cursor()
				if selectedRow < len(m.tickets) {
					selectedTicket := m.tickets[selectedRow]
					selected_branch := git_utils.FormatBranchName(selectedTicket)
					m.view = "form"

					m.form = createForm(&m, selected_branch)
					return m, m.form.Init(), true
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
