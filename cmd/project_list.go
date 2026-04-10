package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:     "projects",
	Short:   "List YouTrack projects",
	Long:    "List all YouTrack projects with their short names.",
	Example: `  yt projects`,
	RunE:    runProjectList,
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}

func runProjectList(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	projects, err := client.ListProjects()
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, projects)
	}
	return format.ProjectList(w, projects)
}
