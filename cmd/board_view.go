package cmd

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/allbin/yt/internal/state"
	"github.com/allbin/yt/internal/tui"
	"github.com/spf13/cobra"
)

var boardViewSprint string

var boardViewCmd = &cobra.Command{
	Use:   "view [name]",
	Short: "Open interactive board viewer",
	Long: `Open a full-screen interactive viewer for an agile board.

Shows board columns, swimlanes, and issue cards in a navigable grid.
Supports changing issue state, opening issue details, and refreshing.

Defaults to the current sprint unless --sprint is specified.`,
	Example: `  # open board viewer
  yt board view HållKoll

  # specific sprint
  yt board view HållKoll --sprint 2025-02`,
	Args: cobra.ExactArgs(1),
	RunE: runBoardView,
}

func init() {
	boardCmd.AddCommand(boardViewCmd)
	boardViewCmd.Flags().StringVar(&boardViewSprint, "sprint", "", "sprint name (default: current)")
}

func runBoardView(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	viewer := tui.NewBoardViewer(client, args[0], boardViewSprint, state.Load())
	_, err = tea.NewProgram(viewer, tea.WithAltScreen()).Run()
	return err
}
