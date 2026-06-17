package cmd

import (
	"github.com/spf13/cobra"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Inspect agile board sprints",
	Long: `Inspect the sprints of an agile board.

Use subcommands to list a board's sprints.`,
	Example: `  # list a board's sprints
  yt sprint list AllTix`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(sprintCmd)
}
