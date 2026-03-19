package cmd

import (
	"os"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List YouTrack issues",
	Long: `List YouTrack issues with optional filters. Filters are combined into a
YouTrack search query. Use --query for arbitrary YouTrack query syntax.`,
	Example: `  # list issues in a project
  yt issue list -p PROJ

  # list open issues assigned to me
  yt issue list -p PROJ -s Open -a me

  # arbitrary YouTrack query
  yt issue list -q "tag: {Ready for QA} sort by: updated desc"

  # combine filters with raw query
  yt issue list -p PROJ -q "created: today"

  # output as JSON, limit to 5 results
  yt issue list -p PROJ -n 5 --json`,
	RunE: runIssueList,
}

var (
	listProject  string
	listState    string
	listAssignee string
	listQuery    string
	listLimit    int
)

func init() {
	issueCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listProject, "project", "p", "", "filter by project")
	listCmd.Flags().StringVarP(&listState, "state", "s", "", "filter by state")
	listCmd.Flags().StringVarP(&listAssignee, "assignee", "a", "", "filter by assignee")
	listCmd.Flags().StringVarP(&listQuery, "query", "q", "", "raw YouTrack query")
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 20, "max results")
}

func runIssueList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	query := youtrack.BuildQuery(listProject, listState, listAssignee, listQuery)

	issues, err := client.ListIssues(query, listLimit)
	if err != nil {
		return err
	}

	if jsonOutput {
		return format.JSON(os.Stdout, issues)
	}
	return format.IssueList(os.Stdout, issues)
}

