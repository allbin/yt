package cmd

import (
	"fmt"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var commentMessage string

var commentCmd = &cobra.Command{
	Use:   "comment <id>",
	Short: "Add a comment to an issue",
	Long: `Post a new comment on a YouTrack issue. The comment text is provided
via the --message flag.`,
	Example: `  # add a comment
  yt issue comment PROJ-123 -m "Looks good, merging."

  # JSON output of created comment
  yt issue comment PROJ-123 -m "Done" --json`,
	Args: cobra.ExactArgs(1),
	RunE: runIssueComment,
}

func init() {
	issueCmd.AddCommand(commentCmd)
	commentCmd.Flags().StringVarP(&commentMessage, "message", "m", "", "comment text (required)")
	_ = commentCmd.MarkFlagRequired("message")
}

func runIssueComment(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	comment, err := client.AddComment(args[0], commentMessage)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, comment)
	}
	_, err = fmt.Fprintf(w, "Comment %s added to %s\n", comment.ID, args[0])
	return err
}
