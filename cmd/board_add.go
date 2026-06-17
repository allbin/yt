package cmd

import (
	"fmt"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var boardAddSprint string

var boardAddCmd = &cobra.Command{
	Use:   "add <board> <issue>...",
	Short: "Add issue(s) to a board's sprint",
	Long: `Add one or more issues to an agile board by placing them on a sprint.

The board is matched by name (case-insensitive). Without --sprint the board's
current sprint is used.

Adding is idempotent: an issue already on the sprint is reported as such and
left unchanged.`,
	Example: `  # add an issue to the current sprint
  yt board add AllTix AX-812

  # add several issues to a specific sprint
  yt board add AllTix AX-812 AX-813 --sprint 2025-06

  # JSON output
  yt board add AllTix AX-812 --json`,
	Args: cobra.MinimumNArgs(2),
	RunE: runBoardAdd,
}

func init() {
	boardCmd.AddCommand(boardAddCmd)
	boardAddCmd.Flags().StringVar(&boardAddSprint, "sprint", "", "sprint name (default: current)")
}

func runBoardAdd(cmd *cobra.Command, args []string) error {
	boardName, issues := args[0], args[1:]

	client, err := apiFactory()
	if err != nil {
		return err
	}

	board, sprint, err := loadBoardSprint(client, boardName, boardAddSprint)
	if err != nil {
		return err
	}

	changes, err := addIssuesToSprint(client, board, sprint, issues)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, changes)
	}

	arrow := format.StyleDim.Render("→")
	for _, c := range changes {
		note := ""
		if !c.Changed {
			note = " " + format.StyleDim.Render("(already on board)")
		}
		if _, err := fmt.Fprintf(w, "%s %s %s %s %s%s\n",
			format.StyleID.Render(c.Issue), arrow, c.Board,
			format.StyleDim.Render("/"), c.Sprint, note); err != nil {
			return err
		}
	}
	return nil
}
