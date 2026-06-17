package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var sprintListCmd = &cobra.Command{
	Use:   "list <board>",
	Short: "List a board's sprints",
	Long: `List the sprints of an agile board, marking the current sprint.

The board is matched by name (case-insensitive), like "yt board".`,
	Example: `  yt sprint list AllTix
  yt sprint list AllTix --json`,
	Args: cobra.ExactArgs(1),
	RunE: runSprintList,
}

func init() {
	sprintCmd.AddCommand(sprintListCmd)
}

func runSprintList(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	board, err := client.GetBoardByName(args[0])
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, board.Sprints)
	}
	return format.SprintList(w, board)
}
