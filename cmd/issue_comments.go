package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var commentsCmd = &cobra.Command{
	Use:   "comments <id>",
	Short: "List comments on an issue",
	Long: `List all comments on a YouTrack issue, showing author, timestamp,
and text for each comment.`,
	Example: `  # list comments
  yt issue comments PROJ-123

  # JSON output
  yt issue comments PROJ-123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runIssueComments,
}

func init() {
	issueCmd.AddCommand(commentsCmd)
}

func runIssueComments(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	comments, err := client.ListComments(args[0])
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, comments)
	}
	return format.CommentList(w, comments)
}
