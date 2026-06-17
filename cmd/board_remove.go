package cmd

import (
	"fmt"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var boardRemoveSprint string

var boardRemoveCmd = &cobra.Command{
	Use:     "remove <board> <issue>...",
	Aliases: []string{"rm"},
	Short:   "Remove issue(s) from a board's sprint",
	Long: `Remove one or more issues from an agile board's sprint.

The board is matched by name (case-insensitive). Without --sprint the board's
current sprint is used.

Removing is idempotent: an issue not on the sprint is reported as such.`,
	Example: `  # remove an issue from the current sprint
  yt board remove AllTix AX-812

  # remove from a specific sprint
  yt board remove AllTix AX-812 --sprint 2025-06

  # JSON output
  yt board remove AllTix AX-812 --json`,
	Args: cobra.MinimumNArgs(2),
	RunE: runBoardRemove,
}

func init() {
	boardCmd.AddCommand(boardRemoveCmd)
	boardRemoveCmd.Flags().StringVar(&boardRemoveSprint, "sprint", "", "sprint name (default: current)")
}

func runBoardRemove(cmd *cobra.Command, args []string) error {
	boardName, issues := args[0], args[1:]

	client, err := apiFactory()
	if err != nil {
		return err
	}

	board, sprint, err := loadBoardSprint(client, boardName, boardRemoveSprint)
	if err != nil {
		return err
	}

	changes, err := removeIssuesFromSprint(client, board, sprint, issues)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, changes)
	}

	cross := format.StyleDim.Render("✕")
	for _, c := range changes {
		note := format.StyleDim.Render("(removed)")
		if !c.Changed {
			note = format.StyleDim.Render("(not on board)")
		}
		if _, err := fmt.Fprintf(w, "%s %s %s %s %s %s\n",
			format.StyleID.Render(c.Issue), cross, c.Board,
			format.StyleDim.Render("/"), c.Sprint, note); err != nil {
			return err
		}
	}
	return nil
}
