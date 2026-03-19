package cmd

import (
	"os"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var (
	createProject     string
	createSummary     string
	createDescription string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new YouTrack issue",
	Long: `Create a new issue in the specified YouTrack project. Requires a project
short name and summary. Optionally accepts a description.

The created issue is displayed after creation.`,
	Example: `  # create a minimal issue
  yt issue create -p PROJ -s "Fix login bug"

  # create with description
  yt issue create -p PROJ -s "Add dark mode" -d "Support system-level dark mode preference"

  # output as JSON
  yt issue create -p PROJ -s "New feature" --json`,
	RunE: runIssueCreate,
}

func init() {
	issueCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createProject, "project", "p", "", "project short name (required)")
	createCmd.Flags().StringVarP(&createSummary, "summary", "s", "", "issue summary (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "issue description")
	_ = createCmd.MarkFlagRequired("project")
	_ = createCmd.MarkFlagRequired("summary")
}

func runIssueCreate(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	issue, err := client.CreateIssue(createProject, createSummary, createDescription)
	if err != nil {
		return err
	}

	if jsonOutput {
		return format.JSON(os.Stdout, issue)
	}
	return format.Issue(os.Stdout, issue)
}
