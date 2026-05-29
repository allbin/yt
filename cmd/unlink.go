package cmd

import (
	"fmt"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var unlinkCmd = &cobra.Command{
	Use:   "unlink <id> <relation> <target-id>",
	Short: "Remove a link between issues",
	Long: `Remove an existing link from an issue to a target issue.

The relation is matched the same way as "yt link" (kebab, spaced, or squashed
form). The matching link must already exist or the command errors.`,
	Example: `  # remove a subtask link
  yt unlink AX-804 subtask-of AX-332

  # remove a relation
  yt unlink AX-1 relates AX-2

  # JSON output (the source issue's remaining links)
  yt unlink AX-804 subtask-of AX-332 --json`,
	Args:              cobra.ExactArgs(3),
	RunE:              runUnlink,
	ValidArgsFunction: completeRelation,
}

func init() {
	rootCmd.AddCommand(unlinkCmd)
}

func runUnlink(cmd *cobra.Command, args []string) error {
	sourceID, alias, target := args[0], args[1], args[2]

	client, err := apiFactory()
	if err != nil {
		return err
	}

	rel, err := resolveRelation(client, alias)
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(sourceID)
	if err != nil {
		return err
	}

	link, matched := youtrack.FindLink(issue.Links, rel, target)
	if link == nil {
		return fmt.Errorf("no %q link from %s to %s", rel.Phrase, sourceID, target)
	}

	// The DELETE endpoint requires the linked issue's internal id, not idReadable.
	targetRef := matched.ID
	if targetRef == "" {
		targetRef = target
	}
	if err := client.RemoveLink(sourceID, link.ID, targetRef); err != nil {
		return err
	}

	issue, err = client.GetIssue(sourceID)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issue.Links)
	}

	arrow := format.StyleDim.Render("✕")
	_, err = fmt.Fprintf(w, "%s %s %s %s %s\n",
		format.StyleID.Render(sourceID), rel.Phrase, arrow, target, format.StyleDim.Render("(removed)"))
	return err
}
