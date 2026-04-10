package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var (
	createProject     string
	createSummary     string
	createDescription string
	createTags        []string
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

  # create with tags
  yt issue create -p PROJ -s "Fix stale state" -t tech-debt -t scheduler

  # output as JSON
  yt issue create -p PROJ -s "New feature" --json`,
	RunE: runIssueCreate,
}

func init() {
	issueCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createProject, "project", "p", "", "project short name (required)")
	createCmd.Flags().StringVarP(&createSummary, "summary", "s", "", "issue summary (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "issue description")
	createCmd.Flags().StringSliceVarP(&createTags, "tag", "t", nil, "add tag (repeatable)")
	_ = createCmd.MarkFlagRequired("project")
	_ = createCmd.MarkFlagRequired("summary")
}

func runIssueCreate(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	issue, err := client.CreateIssue(createProject, createSummary, createDescription, createTags)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issue)
	}
	return format.Issue(w, issue)
}
