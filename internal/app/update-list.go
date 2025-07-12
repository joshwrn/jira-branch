package app

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshwrn/jira-branch/internal/git_utils"
	"github.com/joshwrn/jira-branch/internal/jira"
	"github.com/zalando/go-keyring"
)

func (m *model) updateTableSize() {
	if m.width > 0 && m.height > 0 {
		keyWidth := 10
		typeWidth := 10
		statusWidth := 25
		createdWidth := 15
		summaryWidth := max(25, m.width-keyWidth-typeWidth-statusWidth-createdWidth-12)

		columns := []table.Column{
			{Title: "Key", Width: keyWidth},
			{Title: "Type", Width: typeWidth},
			{Title: "Summary", Width: summaryWidth},
			{Title: "Status", Width: statusWidth},
			{Title: "Created", Width: createdWidth},
		}
		m.list.SetColumns(columns)
		m.list.SetWidth(m.width - 2)
		height := m.height - 3
		if m.showSearch {
			height = height - 1
		}
		m.list.SetHeight(height)
	}
}

type ticketsMsg struct {
	tickets                   []jira.JiraTicketsMsg
	shouldOverwriteAllTickets bool
	err                       error
}

func updateList(m model, msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.isLoading = true
			tickets, err := jira.GetJiraTickets(m.credentials)
			return m, tea.Batch(
				func() tea.Msg {
					return ticketsMsg{
						tickets:                   tickets,
						err:                       err,
						shouldOverwriteAllTickets: true,
					}
				},
				m.spinner.Tick,
			)
		case "S":
			m.view = "credentials"
			m.credentialInputs = CreateCredentialInputs(m.width)
			m.currentField = 0
			m.isLoggedIn = false
			keyring.Delete("jira-cli", "credentials")
			return m, textinput.Blink
		case "enter":
			if m.view == "list" && len(m.tickets) > 0 {
				selectedRow := m.list.Cursor()
				if selectedRow < len(m.tickets) {
					selectedTicket := m.tickets[selectedRow]
					selected_branch := git_utils.FormatBranchName(selectedTicket)
					m.view = "form"

					m.form = createForm(&m, selected_branch)
					return m, m.form.Init()
				}
			}
		case "/":
			m.searchInput = createSearchInput(m.width)
			m.searchInput.SetValue(m.search)
			m.searchInput.Focus()
			m.showSearch = true
			m.updateTableSize()
			return m, nil
		}
	case credentialsNeededMsg:
		m.view = "credentials"
		m.credentialInputs = CreateCredentialInputs(m.width)
		m.currentField = 0
		return m, textinput.Blink
	case ticketsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.isLoading = false
			return m, nil
		}
		if msg.shouldOverwriteAllTickets {
			m.allTickets = msg.tickets
		}
		m.tickets = msg.tickets

		filterTickets(&m)

		m.isLoading = false
		m.isLoggedIn = true
		return m, nil
	}

	updatedTable, cmd := m.list.Update(msg)
	m.list = updatedTable
	return m, cmd

}
