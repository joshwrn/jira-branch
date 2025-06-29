package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"jira_cli/internal/jira"

	"github.com/charmbracelet/lipgloss"
)

type jiraTicketsMsg struct {
	Key     string
	Summary string
}
type model struct {
	table     table.Model
	cursor    int
	selected  map[int]struct{}
	isLoading bool
	spinner   spinner.Model
	err       error
	width     int
	height    int
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type errMsg error

func fetchJiraTickets() tea.Cmd {
	return func() tea.Msg {
		j, err := jira.GetJiraTickets()
		if err != nil {
			return errMsg(err)
		}

		newChoices := []jiraTicketsMsg{}
		for _, issue := range j.Issues {
			newChoices = append(newChoices, jiraTicketsMsg{
				Key:     issue.Key,
				Summary: issue.Fields.Summary,
			})
		}

		return newChoices
	}
}

func (m *model) updateTableSize() {
	if m.width > 0 && m.height > 0 {
		availableHeight := m.height
		if availableHeight < 5 {
			availableHeight = 5
		}

		totalBorderWidth := 8
		availableWidth := m.width - totalBorderWidth

		if availableWidth > 0 {
			keyWidth := max(8, availableWidth*5/100)
			summaryWidth := max(20, availableWidth*80/100)

			cols := m.table.Columns()
			if len(cols) > 0 {
				cols[0].Width = keyWidth
				cols[1].Width = summaryWidth
				m.table.SetColumns(cols)
			}
		}

		m.table.SetHeight(availableHeight)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		fetchJiraTickets(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.isLoading && m.err == nil {
			m.updateTableSize()
		}

	case tea.KeyMsg:
		if m.isLoading {
			return m, nil
		}

		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if len(m.table.SelectedRow()) > 0 {
				return m, tea.Batch(
					tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
				)
			}
		case "r":
			m.isLoading = true
			m.err = nil
			return m, tea.Batch(
				m.spinner.Tick,
				fetchJiraTickets(),
			)
		}

	case []jiraTicketsMsg:
		rows := []table.Row{}
		for _, choice := range msg {
			rows = append(rows, table.Row{choice.Key, choice.Summary})
		}

		t := table.New(
			table.WithColumns([]table.Column{
				{Title: "Select a ticket", Width: 10},
				{Title: "", Width: 50},
			}),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(10),
		)

		s := table.DefaultStyles()

		s.Header = s.Header.
			Bold(false).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			Height(0).
			Padding(0).
			Foreground(lipgloss.Color("7")).
			Bold(true)

		s.Selected = s.Selected.
			UnsetForeground().
			Foreground(lipgloss.Color("7")).
			Background(lipgloss.Color("0")).
			Bold(true)
		t.SetStyles(s)

		m.table = t
		m.isLoading = false
		m.updateTableSize()
		return m, nil

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

	if !m.isLoading {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd

}

func (m model) View() string {
	if m.isLoading {
		loadingText := fmt.Sprintf("%s %s", m.spinner.View(), "Loading Jira tickets...")

		return fmt.Sprintf("\n%s\n\nPress q to quit", loadingText)
	}

	if m.err != nil {
		errorText := fmt.Sprintf("Error loading data: %v", m.err)
		helpText := "Press 'r' to retry or 'q' to quit"

		if m.width > 0 {
			errorPadding := max(0, (m.width-len(errorText))/2)
			helpPadding := max(0, (m.width-len(helpText))/2)
			errorText = strings.Repeat(" ", errorPadding) + errorText
			helpText = strings.Repeat(" ", helpPadding) + helpText
		}

		return fmt.Sprintf("\n%s\n\n%s\n", errorText, helpText)
	}

	return m.table.View()
}

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found. Make sure you have set the required environment variables.")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	emptyTable := table.New(
		table.WithColumns([]table.Column{}),
		table.WithRows([]table.Row{}),
		table.WithHeight(10),
	)

	m := model{
		table:     emptyTable,
		spinner:   s,
		isLoading: true,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
