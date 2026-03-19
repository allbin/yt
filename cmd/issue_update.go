package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var (
	updateState    string
	updateAssignee string
	updatePriority string
	updateType     string
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a YouTrack issue",
	Long: `Update fields on a YouTrack issue by executing a command string.
Supports setting state, assignee, priority, and type. Multiple flags
can be combined in a single invocation.

After a successful update the issue is fetched and displayed.`,
	Example: `  # set state
  yt issue update PROJ-123 -s "In Progress"

  # set assignee and priority
  yt issue update PROJ-123 -a me -p Critical

  # set type
  yt issue update PROJ-123 -t Bug

  # combine all fields
  yt issue update PROJ-123 -s Open -a john -p Normal -t Task`,
	Args: cobra.ExactArgs(1),
	RunE: runIssueUpdate,
}

func init() {
	issueCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&updateState, "state", "s", "", "set issue state")
	updateCmd.Flags().StringVarP(&updateAssignee, "assignee", "a", "", "set assignee (supports 'me')")
	updateCmd.Flags().StringVarP(&updatePriority, "priority", "p", "", "set priority")
	updateCmd.Flags().StringVarP(&updateType, "type", "t", "", "set issue type")
}

func runIssueUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]

	client, err := newClient()
	if err != nil {
		return err
	}

	assignee, err := resolveAssignee(client, updateAssignee)
	if err != nil {
		return err
	}

	command := buildCommand(updateState, assignee, updatePriority, updateType)
	if command == "" {
		return fmt.Errorf("no fields to update; use --state, --assignee, --priority, or --type")
	}

	if err := client.UpdateIssue(id, command); err != nil {
		return err
	}

	issue, err := client.GetIssue(id)
	if err != nil {
		return err
	}

	if jsonOutput {
		return format.JSON(os.Stdout, issue)
	}
	return format.Issue(os.Stdout, issue)
}

// buildCommand constructs a YouTrack command string from field values.
// Multi-word values are wrapped in braces.
func buildCommand(state, assignee, priority, typ string) string {
	var parts []string
	if state != "" {
		parts = append(parts, "State "+braceWrap(state))
	}
	if assignee != "" {
		parts = append(parts, "Assignee "+braceWrap(assignee))
	}
	if priority != "" {
		parts = append(parts, "Priority "+braceWrap(priority))
	}
	if typ != "" {
		parts = append(parts, "Type "+braceWrap(typ))
	}
	return strings.Join(parts, " ")
}

// braceWrap wraps s in braces if it contains spaces.
func braceWrap(s string) string {
	if strings.Contains(s, " ") {
		return "{" + s + "}"
	}
	return s
}
