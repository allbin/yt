package cmd

import "github.com/spf13/cobra"

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Inspect YouTrack project details",
	Long:  "Inspect custom fields and configuration for a YouTrack project.",
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
