package app

import (
	"errors"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/joshwrn/jira-branch/internal/git_utils"
)

func CreateForm(m *model, initialBranchName string) *huh.Form {
	branchName := initialBranchName
	shouldMarkAsInProgress := true
	m.formBranchName = &branchName
	m.formShouldMarkAsInProgress = &shouldMarkAsInProgress

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Branch name").
				Value(m.formBranchName).Validate(func(value string) error {
				if len(value) == 0 {
					return errors.New("branch name is required")
				}
				if git_utils.BranchNameRegex.MatchString(value) {
					return errors.New("branch name can only contain letters, numbers, '-', '_', '/' and '.'")
				}
				return nil
			}),
			huh.NewConfirm().
				Title("Mark as in progress?").
				Value(m.formShouldMarkAsInProgress).
				Affirmative("Yes").
				Negative("No").Inline(true),
		).WithTheme(CustomTheme()),
	)

	return form
}

var catppuccin = huh.ThemeCatppuccin()

func CustomTheme() *huh.Theme {
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
