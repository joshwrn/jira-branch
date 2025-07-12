package app

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshwrn/jira-branch/internal/git_utils"
	"github.com/joshwrn/jira-branch/internal/gui"
)

func viewForm(m model) string {
	if m.isSubmittingForm {
		return gui.CreateLoadingView(&gui.LoadingView{
			Text:    "Creating branch and updating Jira...",
			Width:   m.width,
			Height:  m.height,
			Spinner: m.spinner,
		})
	}

	nameLen := len(*m.formBranchName)
	formWidth := m.width - m.width/4 - 5
	if nameLen > formWidth-8 {
		return lipgloss.NewStyle().
			Width(m.width).
			PaddingTop(2).
			PaddingLeft(2).
			Render(m.form.View())
	}

	formView := lipgloss.NewStyle().
		Width(formWidth).
		PaddingTop(2).
		PaddingLeft(2).
		Render(m.form.View())

	sidebar := createSidebar(&m)

	return lipgloss.JoinHorizontal(lipgloss.Top, formView, sidebar)
}

func createForm(m *model, initialBranchName string) *huh.Form {
	branchName := initialBranchName

	shouldMarkAsInProgress := true
	status := m.list.SelectedRow()[3]
	isInProgress := strings.EqualFold(status, "In Progress")

	if isInProgress {
		shouldMarkAsInProgress = false
	}

	m.formBranchName = &branchName
	m.formShouldMarkAsInProgress = &shouldMarkAsInProgress

	inputField := huh.NewInput().
		Title("Branch name").
		Value(m.formBranchName).Validate(func(value string) error {
		if len(value) == 0 {
			return errors.New("branch name is required")
		}
		if git_utils.BranchNameRegex.MatchString(value) {
			return errors.New("branch name can only contain letters, numbers, '-', '_', '/' and '.'")
		}
		return nil
	})

	fields := []huh.Field{inputField}

	if !isInProgress {
		confirmField := huh.NewConfirm().
			Title("Mark as in progress?").
			Value(m.formShouldMarkAsInProgress).
			Affirmative("Yes").
			Negative("No").Inline(true)
		fields = append(fields, confirmField)
	}

	form := huh.NewForm(
		huh.NewGroup(fields...).WithTheme(customTheme()),
	)

	return form
}

func createSidebar(m *model) string {
	b := strings.Builder{}

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("4")).
		Render(m.list.SelectedRow()[2]))
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("5")).
		Render(m.list.SelectedRow()[1]))
	b.WriteString(" - ")
	b.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Render(m.list.SelectedRow()[3]))
	b.WriteString("\n\n")

	b.WriteString(m.list.SelectedRow()[4])

	sidebar := lipgloss.NewStyle().
		Width(m.width/4).
		Height(m.height-3).
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(1, 3).
		BorderForeground(lipgloss.Color("8")).
		Render(b.String())

	return sidebar
}

var catppuccin = huh.ThemeCatppuccin()

func customTheme() *huh.Theme {
	// Focused field styling
	focused := catppuccin.Focused
	focused.Title = focused.Title.Foreground(lipgloss.Color("4"))
	focused.Description = focused.Description.Foreground(lipgloss.Color("7"))
	focused.TextInput.Prompt = focused.TextInput.Prompt.Foreground(lipgloss.Color("4"))
	focused.TextInput.Placeholder = focused.TextInput.Placeholder.
		Foreground(lipgloss.Color("7")).Faint(true)
	focused.TextInput.Text = focused.TextInput.Text.
		Foreground(lipgloss.Color("7"))
	focused.TextInput.Cursor = focused.TextInput.Cursor.
		Foreground(lipgloss.Color("5"))
	focused.TextInput.CursorText = focused.TextInput.CursorText.
		Foreground(lipgloss.Color("5"))

	focused.FocusedButton = focused.FocusedButton.Background(lipgloss.Color("4"))
	focused.BlurredButton = focused.BlurredButton.Background(lipgloss.Color("0"))

	focused.Base = focused.Base.
		BorderForeground(lipgloss.Color("7"))

	// Blurred field styling
	blurred := catppuccin.Blurred
	blurred.Title = blurred.Title.Foreground(lipgloss.Color("4"))
	blurred.Description = blurred.Description.Foreground(lipgloss.Color("7"))
	blurred.TextInput.Prompt = blurred.TextInput.Prompt.Foreground(lipgloss.Color("4"))
	blurred.TextInput.Placeholder = blurred.TextInput.Placeholder.
		Foreground(lipgloss.Color("7")).Faint(true)
	blurred.TextInput.Text = blurred.TextInput.Text.
		Foreground(lipgloss.Color("7"))
	blurred.TextInput.Cursor = blurred.TextInput.Cursor.
		Foreground(lipgloss.Color("5"))
	blurred.TextInput.CursorText = blurred.TextInput.CursorText.
		Foreground(lipgloss.Color("5"))

	blurred.FocusedButton = blurred.FocusedButton.Background(lipgloss.Color("4"))
	blurred.BlurredButton = blurred.BlurredButton.Background(lipgloss.Color("0"))

	blurred.Base = blurred.Base.
		BorderForeground(lipgloss.Color("0"))

	// Help styling
	help := catppuccin.Help
	help.Ellipsis = help.Ellipsis.Foreground(lipgloss.Color("7")).Faint(true)
	help.ShortKey = help.ShortKey.Foreground(lipgloss.Color("7"))
	help.ShortDesc = help.ShortDesc.Foreground(lipgloss.Color("7")).Faint(true)
	help.ShortSeparator = help.ShortSeparator.Foreground(lipgloss.Color("7")).Faint(true)

	return &huh.Theme{
		Form:           catppuccin.Form,
		Group:          catppuccin.Group,
		FieldSeparator: catppuccin.FieldSeparator,
		Blurred:        blurred,
		Focused:        focused,
		Help:           help,
	}
}
