package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var linksCmd = &cobra.Command{
	Use:   "links [id]",
	Short: "List an issue's links",
	Long: `List the links on a YouTrack issue, grouped by relation, showing the
directed phrase and each linked issue's ID and summary.

If no ID is given, attempts to detect it from the current git branch name.`,
	Example: `  # list links for an issue
  yt links AX-804

  # auto-detect from current branch
  yt links

  # JSON output
  yt links AX-804 --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLinks,
}

func init() {
	rootCmd.AddCommand(linksCmd)
}

func runLinks(cmd *cobra.Command, args []string) error {
	id := issueIDFromArgs(args)
	if id == "" {
		return cmd.Help()
	}

	client, err := apiFactory()
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(id)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issue.Links)
	}
	return format.Links(w, issue.Links)
}
