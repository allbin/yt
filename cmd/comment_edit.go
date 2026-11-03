package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var (
	commentEditMessage string
	commentDeleteYes   bool
)

// commentTarget describes what a comment subcommand works on: "issue" or
// "article", with the ID check and examples that go with it.
type commentTarget struct {
	noun    string
	example string
	checkID func(string) (string, error)
}

var (
	issueCommentTarget   = commentTarget{"issue", "PROJ-123", func(id string) (string, error) { return id, nil }}
	articleCommentTarget = commentTarget{"article", "PROJ-A-12", articleIDArg}
)

func newCommentEditCmd(t commentTarget) *cobra.Command {
	c := &cobra.Command{
		Use:   "edit <id> <commentID>",
		Short: "Replace the text of a comment on an " + t.noun,
		Long: `Replace the whole text of a comment on an ` + t.noun + `. The comment ID is the
"id" in "yt ` + t.noun + ` comments --json", or the one "yt ` + t.noun + ` comment" printed.
YouTrack lets you edit only comments you may change, usually your own.

The message accepts "@path" to read from a file or "-" to read from stdin.`,
		Example: `  yt ` + t.noun + ` comment edit ` + t.example + ` 4-3431 -m "Corrected: deploy on Friday."

  # from a file
  yt ` + t.noun + ` comment edit ` + t.example + ` 4-3431 -m @comment.md --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := t.checkID(args[0])
			if err != nil {
				return err
			}
			return runCommentEdit(cmd, id, args[1])
		},
	}
	c.Flags().StringVarP(&commentEditMessage, "message", "m", "", "new comment text (@file or - for stdin) (required)")
	_ = c.MarkFlagRequired("message")
	return c
}

func newCommentDeleteCmd(t commentTarget) *cobra.Command {
	c := &cobra.Command{
		Use:   "delete <id> <commentID>",
		Short: "Delete a comment on an " + t.noun,
		Long: `Delete a comment on an ` + t.noun + ` permanently. It cannot be undone, so --yes is
required; without it the command names the comment and changes nothing.`,
		Example: `  yt ` + t.noun + ` comment delete ` + t.example + ` 4-3431 --yes`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := t.checkID(args[0])
			if err != nil {
				return err
			}
			return runCommentDelete(cmd, id, args[1])
		},
	}
	c.Flags().BoolVar(&commentDeleteYes, "yes", false, "confirm the permanent delete")
	return c
}

func init() {
	commentCmd.AddCommand(newCommentEditCmd(issueCommentTarget), newCommentDeleteCmd(issueCommentTarget))
	articleCommentCmd.AddCommand(newCommentEditCmd(articleCommentTarget), newCommentDeleteCmd(articleCommentTarget))
}

func runCommentEdit(cmd *cobra.Command, id, commentID string) error {
	text, err := readTextArg(commentEditMessage, cmd.InOrStdin())
	if err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("comment text cannot be empty; use delete to remove a comment")
	}
	client, err := apiFactory()
	if err != nil {
		return err
	}
	comment, err := client.UpdateComment(id, commentID, text)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, comment)
	}
	_, err = fmt.Fprintf(w, "Comment %s on %s updated\n", commentID, id)
	return err
}

func runCommentDelete(cmd *cobra.Command, id, commentID string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}
	// Read first: a wrong ID errors before anything is deleted, and the
	// refusal without --yes can quote the comment.
	c, err := client.GetComment(id, commentID)
	if err != nil {
		return err
	}
	if !commentDeleteYes {
		return fmt.Errorf("deleting comment %s by %s (%q) cannot be undone; rerun with --yes",
			commentID, c.View().Author, format.FirstLine(c.Text, 60))
	}
	if err := client.DeleteComment(id, commentID); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Comment %s on %s deleted\n", commentID, id)
	return err
}
