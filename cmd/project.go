package cmd

import "github.com/spf13/cobra"

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage YouTrack projects",
	Long:  "List and inspect YouTrack projects.",
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
