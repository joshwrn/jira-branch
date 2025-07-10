package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshwrn/jira-branch/internal/jira"
	"github.com/joshwrn/jira-branch/internal/utils"
)

func keyIncludesSearch(search string, ticket jira.JiraTicketsMsg) bool {
	keyMap := map[string]string{
		"key":     ticket.Key,
		"summary": ticket.Summary,
		"type":    ticket.Type,
		"status":  ticket.Status,
	}
	for _, value := range keyMap {
		if strings.Contains(strings.ToLower(value), strings.ToLower(search)) {
			return true
		}
	}
	return false
}

func filterTickets(m *model) {
	searchTerm := m.searchInput.Value()
	filteredTickets := []jira.JiraTicketsMsg{}

	for _, ticket := range m.allTickets {
		if keyIncludesSearch(searchTerm, ticket) {
			filteredTickets = append(filteredTickets, ticket)
		}
	}

	utils.Log.Info().Msgf("filteredTickets: %v", len(filteredTickets))

	m.tickets = filteredTickets
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
	m.updateTableSize()
}

func updateSearch(m model, msg tea.Msg) (model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.showSearch = false
			m.updateTableSize()
			m.search = m.searchInput.Value()
			m.table.GotoTop()
			return m, nil, true
		case "esc":
			m.showSearch = false
			m.searchInput.SetValue("")
			m.search = ""
			filterTickets(&m)
			m.table.GotoTop()
			return m, nil, true
		}
	}

	return m, nil, false
}
