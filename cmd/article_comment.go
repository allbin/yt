package cmd

import (
	"github.com/spf13/cobra"
)

var articleCommentCmd = &cobra.Command{
	Use:   "comment <id>",
	Short: "Add a comment to an article",
	Long: `Post a new comment on a knowledge base article. The comment text is provided
via the --message flag.

The message accepts "@path" to read from a file or "-" to read from stdin.`,
	Example: `  # add a comment
  yt article comment PROJ-A-12 -m "Action items moved to PROJ-123."

  # JSON output of created comment
  yt article comment PROJ-A-12 -m @comment.md --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := articleIDArg(args[0])
		if err != nil {
			return err
		}
		return runIssueComment(cmd, []string{id})
	},
}

var articleCommentsCmd = &cobra.Command{
	Use:     "comments <id>",
	Short:   "List comments on an article",
	Long:    `List all comments on a knowledge base article, showing author, timestamp, and text.`,
	Example: `  yt article comments PROJ-A-12 --json`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := articleIDArg(args[0])
		if err != nil {
			return err
		}
		return runIssueComments(cmd, []string{id})
	},
}

func init() {
	articleCmd.AddCommand(articleCommentCmd, articleCommentsCmd)
	// Shares commentMessage with "yt issue comment": both run runIssueComment.
	articleCommentCmd.Flags().StringVarP(&commentMessage, "message", "m", "", "comment text (@file or - for stdin) (required)")
	_ = articleCommentCmd.MarkFlagRequired("message")
}
