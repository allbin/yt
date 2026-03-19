package cmd

import (
	"os"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue [id]",
	Short: "Show or manage YouTrack issues",
	Long: `Fetch a single YouTrack issue by its readable ID (e.g. PROJ-123) and display
its summary, state, assignee, priority, type, subsystems, tags, and description.

Use subcommands to list and filter issues.`,
	Example: `  # show an issue as formatted text
  yt issue PROJ-123

  # show an issue as JSON
  yt issue PROJ-123 --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIssue,
}

func init() {
	rootCmd.AddCommand(issueCmd)
}

func runIssue(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(args[0])
	if err != nil {
		return err
	}

	if jsonOutput {
		return format.JSON(os.Stdout, issue)
	}
	return format.Issue(os.Stdout, issue)
}
