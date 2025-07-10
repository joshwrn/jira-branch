package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joshwrn/jira-branch/internal/jira"
)

func keyIncludesSearch(search string, ticket jira.JiraTicketsMsg) bool {
	keyMap := map[string]string{
		"key":     ticket.Key,
		"summary": ticket.Summary,
		"type":    ticket.Type,
		"status":  ticket.Status,
		"created": ticket.Created,
	}
	for _, value := range keyMap {
		if strings.Contains(strings.ToLower(value), strings.ToLower(search)) {
			return true
		}
	}
	return false
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
			filteredTickets := []jira.JiraTicketsMsg{}
			for _, ticket := range m.allTickets {
				if keyIncludesSearch(m.search, ticket) {
					filteredTickets = append(filteredTickets, ticket)
				}
			}

			return m, func() tea.Msg {
				return ticketsMsg{
					tickets:                   filteredTickets,
					err:                       nil,
					shouldOverwriteAllTickets: false,
				}
			}, true
		case "esc":
			m.showSearch = false
			m.updateTableSize()
			return m, nil, true
		}
	}

	return m, nil, false
}
