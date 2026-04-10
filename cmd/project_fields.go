package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var projectFieldsCmd = &cobra.Command{
	Use:   "fields <project>",
	Short: "List custom fields for a project",
	Long: `List all custom fields configured on a YouTrack project, including
their types and allowed values.

Useful for discovering which fields can be set with --field or --subsystem
on issue create and update commands.`,
	Example: `  # list fields for a project
  yt project fields PROJ

  # output as JSON
  yt project fields PROJ --json`,
	Args:              cobra.ExactArgs(1),
	RunE:              runProjectFields,
	ValidArgsFunction: completeProjectNames,
}

func init() {
	projectCmd.AddCommand(projectFieldsCmd)
}

func runProjectFields(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	fields, err := client.ListProjectFields(args[0])
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, fields)
	}
	return format.ProjectFields(w, fields)
}
