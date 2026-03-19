package cmd

import (
	"os"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var projectListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List YouTrack projects",
	Long:    "List all YouTrack projects with their short names.",
	Example: `  yt project list`,
	RunE:    runProjectList,
}

func init() {
	projectCmd.AddCommand(projectListCmd)
}

func runProjectList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	projects, err := client.ListProjects()
	if err != nil {
		return err
	}

	if jsonOutput {
		return format.JSON(os.Stdout, projects)
	}
	return format.ProjectList(os.Stdout, projects)
}
