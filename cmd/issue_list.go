package cmd

import (
	"fmt"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List YouTrack issues",
	Long: `List YouTrack issues with optional filters. Filters are combined into a
YouTrack search query. Use --query for arbitrary YouTrack query syntax.

--json prints every issue in full: description, attachments and all custom
fields. For long lists add --compact, which prints one small object per
issue (id, summary, state, type, priority, assignee, tags, updated) and asks
the server for nothing else.`,
	Example: `  # list issues in a project
  yt issue list -p PROJ

  # list open issues assigned to me
  yt issue list -p PROJ -s Open -a me

  # arbitrary YouTrack query
  yt issue list -q "tag: {Ready for QA} sort by: updated desc"

  # combine filters with raw query
  yt issue list -p PROJ -q "created: today"

  # output as JSON, limit to 5 results
  yt issue list -p PROJ -n 5 --json

  # a whole project as compact JSON
  yt issue list -p PROJ -n 500 --json --compact`,
	RunE: runIssueList,
}

var (
	listProject  string
	listState    string
	listAssignee string
	listQuery    string
	listLimit    int
	listCompact  bool
)

func init() {
	issueCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&listProject, "project", "p", "", "filter by project")
	listCmd.Flags().StringVarP(&listState, "state", "s", "", "filter by state")
	listCmd.Flags().StringVarP(&listAssignee, "assignee", "a", "", "filter by assignee")
	listCmd.Flags().StringVarP(&listQuery, "query", "q", "", "raw YouTrack query")
	listCmd.Flags().IntVarP(&listLimit, "limit", "n", 20, "max results")
	listCmd.Flags().BoolVar(&listCompact, "compact", false, "with --json: one small object per issue, no description")
}

func runIssueList(cmd *cobra.Command, args []string) error {
	if listCompact && !jsonOutput {
		return fmt.Errorf("--compact is a JSON form; add --json")
	}

	client, err := apiFactory()
	if err != nil {
		return err
	}

	assignee, err := resolveAssignee(client, listAssignee)
	if err != nil {
		return err
	}

	query := youtrack.BuildQuery(listProject, listState, assignee, listQuery)
	w := cmd.OutOrStdout()

	if listCompact {
		compact, err := client.ListIssuesCompact(query, listLimit)
		if err != nil {
			return err
		}
		return format.JSON(w, compact)
	}

	issues, err := client.ListIssues(query, listLimit)
	if err != nil {
		return err
	}
	if jsonOutput {
		return format.JSON(w, issues)
	}
	return format.IssueList(w, issues)
}
