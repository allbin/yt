package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue [id]",
	Short: "Show or manage YouTrack issues",
	Long: `Fetch a single YouTrack issue by its readable ID (e.g. PROJ-123) and display
its summary, state, assignee, priority, type, subsystems, tags, description,
board membership, and the knowledge base articles that mention it.

If no ID is given, attempts to detect it from the current git branch name.

Use subcommands to list and filter issues.`,
	Example: `  # show an issue as formatted text
  yt issue PROJ-123

  # show an issue as JSON
  yt issue PROJ-123 --json

  # auto-detect from current branch (e.g. proj-123-some-slug)
  yt issue`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIssue,
}

func init() {
	rootCmd.AddCommand(issueCmd)
}

func runIssue(cmd *cobra.Command, args []string) error {
	id, client, err := issueClient(args)
	if err != nil {
		return err
	}
	if id == "" {
		return cmd.Help()
	}

	issue, err := client.GetIssue(id)
	if err != nil {
		return err
	}

	// Board/sprint membership is supplementary; ignore lookup failures so a
	// transient board error never blocks displaying the issue itself.
	if boards, err := client.IssueBoards(id); err == nil {
		issue.Boards = boards
	}
	// Same for the knowledge base articles that mention the issue.
	if articles, err := client.ArticlesMentioning(issue.IDReadable); err == nil {
		issue.Articles = articles
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issue)
	}
	return format.Issue(w, issue)
}
