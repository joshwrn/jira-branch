package git_utils

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/joshwrn/jira-branch/internal/jira"

	tea "github.com/charmbracelet/bubbletea"
)

var BranchNameRegex = regexp.MustCompile(`[^a-zA-Z0-9-_./]`)

type errMsg error

func FormatBranchName(ticket jira.JiraTicketsMsg) string {
	prefix := "feature/"
	if ticket.Type == "Bug" {
		prefix = "bugfix/"
	}

	branchName := prefix + ticket.Key + "-" + strings.ToLower(ticket.Summary)
	branchName = strings.ReplaceAll(branchName, " ", "_")
	branchName = BranchNameRegex.ReplaceAllString(branchName, "")

	return branchName
}

func CheckoutBranch(branchName string) tea.Cmd {
	return func() tea.Msg {
		checkCmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branchName)
		branchExists := checkCmd.Run() == nil

		var cmd *exec.Cmd

		if branchExists {
			cmd = exec.Command("git", "checkout", branchName)
		} else {
			cmd = exec.Command("git", "checkout", "-b", branchName)
		}

		err := cmd.Run()

		if err != nil {
			return errMsg(fmt.Errorf("failed to checkout branch %s: %v", branchName, err))
		}

		return tea.Quit()
	}
}
