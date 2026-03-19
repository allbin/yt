package cmd

import (
	"os"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List agile boards",
	Long:  "List all agile boards with their projects and current sprint.",
	Example: `  yt board list
  yt board list --json`,
	RunE: runBoardList,
}

func init() {
	boardCmd.AddCommand(boardListCmd)
}

func runBoardList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	boards, err := client.ListBoards()
	if err != nil {
		return err
	}

	if jsonOutput {
		return format.JSON(os.Stdout, boards)
	}
	return format.BoardList(os.Stdout, boards)
}
