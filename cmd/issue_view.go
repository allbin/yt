package cmd

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/allbin/yt/internal/tui"
	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "Open interactive issue viewer",
	Long: `Open a full-screen interactive viewer for a YouTrack issue.

Shows issue summary, metadata, description, and comments in a scrollable
viewport. Supports changing issue state via an embedded state picker.

If no ID is given, attempts to detect it from the current git branch name.

Without a terminal to draw on, or with --json, prints the issue instead of
opening the viewer -- the same output as "yt issue <ID>".`,
	Example: `  # open viewer for a specific issue
  yt issue view PROJ-123

  # auto-detect from current branch
  yt issue view`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIssueView,
}

func init() {
	issueCmd.AddCommand(viewCmd)
}

func runIssueView(cmd *cobra.Command, args []string) error {
	if !interactive(cmd) {
		return runIssue(cmd, args)
	}

	id := issueIDFromArgs(args)
	if id == "" {
		return cmd.Help()
	}

	client, err := apiFactory()
	if err != nil {
		return err
	}

	viewer := tui.NewIssueViewer(client, id)
	_, err = tea.NewProgram(viewer, tea.WithAltScreen()).Run()
	return err
}
