package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"

	"jira_cli/internal/jira"

	"github.com/charmbracelet/lipgloss"
)

type model struct {
	choices   []string
	cursor    int
	selected  map[int]struct{}
	isLoading bool
	spinner   spinner.Model
	quitting  bool
	err       error
}

func initialModel() model {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		choices: []string{},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected:  make(map[int]struct{}),
		isLoading: true,
		spinner:   sp,
		quitting:  false,
		err:       nil,
	}
}

type errMsg error

type jiraTicketsMsg []string

func fetchJiraTickets() tea.Msg {
	j, err := jira.GetJiraTickets()
	if err != nil {
		return errMsg(err)
	}

	newChoices := []string{}
	for _, issue := range j.Issues {
		newChoices = append(newChoices, issue.Key+" - "+issue.Fields.Summary)
	}

	return jiraTicketsMsg(newChoices)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, fetchJiraTickets)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case jiraTicketsMsg:
		m.choices = msg
		m.isLoading = false

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	if m.isLoading {
		return fmt.Sprintf("\n\n   %s Loading Tickets\n\n", m.spinner.View())
	}

	body := "Choose a ticket.\n\n"

	for i, choice := range m.choices {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		body += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	body += "\nPress q to quit.\n"

	return body
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found. Make sure you have set the required environment variables.")
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
