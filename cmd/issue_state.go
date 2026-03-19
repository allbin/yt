package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/tui"
	"github.com/spf13/cobra"
)

var stateCmd = &cobra.Command{
	Use:   "state [id]",
	Short: "Interactively set issue state",
	Long: `Open an interactive picker to change the state of a YouTrack issue.

Shows all available states for the issue's project with the current state
marked. Navigate with arrow keys or j/k, select with Enter, cancel with
Esc or q.

If no ID is given, attempts to detect it from the current git branch name.`,
	Example: `  # pick state interactively
  yt issue state PROJ-123

  # auto-detect from current branch
  yt issue state`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIssueState,
}

func init() {
	issueCmd.AddCommand(stateCmd)
}

func runIssueState(cmd *cobra.Command, args []string) error {
	id := issueIDFromArgs(args)
	if id == "" {
		return cmd.Help()
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(id)
	if err != nil {
		return err
	}

	states, err := client.GetIssueStates(id)
	if err != nil {
		return err
	}

	currentState := issue.Field("State")
	picker := tui.NewStatePicker(issue.IDReadable, issue.Summary, currentState, states)

	finalModel, err := tea.NewProgram(picker).Run()
	if err != nil {
		return err
	}

	picker, ok := finalModel.(tui.StatePicker)
	if !ok {
		return fmt.Errorf("unexpected state from picker")
	}
	result := picker.Result()
	if result.Cancelled || result.State == currentState {
		return nil
	}

	if err := client.SetIssueState(id, result.State); err != nil {
		return err
	}

	if jsonOutput {
		issue, err = client.GetIssue(id)
		if err != nil {
			return err
		}
		return format.JSON(os.Stdout, issue)
	}

	from := lipgloss.NewStyle().Foreground(format.StateColor(currentState)).Render(currentState)
	to := lipgloss.NewStyle().Foreground(format.StateColor(result.State)).Render(result.State)
	_, err = fmt.Fprintf(os.Stdout, "%s %s %s %s\n", issue.IDReadable, from, format.StyleDim.Render("→"), to)
	return err
}
