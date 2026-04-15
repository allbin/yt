package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var (
	updateState       string
	updateAssignee    string
	updatePriority    string
	updateType        string
	updateSubsystem   string
	updateSummary     string
	updateDescription string
	updateTags        []string
	updateRemoveTags  []string
	updateFields      []string
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a YouTrack issue",
	Long: `Update fields on a YouTrack issue.

Summary and description use the REST API; other fields use the command API.
Both can be combined in a single invocation.

Use --field to set arbitrary custom fields by name.

After a successful update the issue is fetched and displayed.`,
	Example: `  # set state
  yt issue update PROJ-123 -s "In Progress"

  # update summary
  yt issue update PROJ-123 -S "New title"

  # update description
  yt issue update PROJ-123 -d "Updated description"

  # set assignee and priority
  yt issue update PROJ-123 -a me -p Critical

  # set type
  yt issue update PROJ-123 -t Bug

  # set subsystem
  yt issue update PROJ-123 --subsystem API

  # set arbitrary custom field
  yt issue update PROJ-123 --field "Severity=Critical"

  # add tags
  yt issue update PROJ-123 --tag tech-debt --tag scheduler

  # remove a tag
  yt issue update PROJ-123 --remove-tag obsolete

  # combine REST and command fields
  yt issue update PROJ-123 -S "New title" -s "In Progress" -a me`,
	Args: cobra.ExactArgs(1),
	RunE: runIssueUpdate,
}

func init() {
	issueCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(&updateState, "state", "s", "", "set issue state")
	updateCmd.Flags().StringVarP(&updateAssignee, "assignee", "a", "", "set assignee (supports 'me')")
	updateCmd.Flags().StringVarP(&updatePriority, "priority", "p", "", "set priority")
	updateCmd.Flags().StringVarP(&updateType, "type", "t", "", "set issue type")
	updateCmd.Flags().StringVar(&updateSubsystem, "subsystem", "", "set subsystem")
	updateCmd.Flags().StringVarP(&updateSummary, "summary", "S", "", "set issue summary")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "set issue description")
	updateCmd.Flags().StringSliceVar(&updateTags, "tag", nil, "add tag (repeatable)")
	updateCmd.Flags().StringSliceVar(&updateRemoveTags, "remove-tag", nil, "remove tag (repeatable)")
	updateCmd.Flags().StringSliceVar(&updateFields, "field", nil, `set custom field as "Name=Value" (repeatable)`)

	_ = updateCmd.RegisterFlagCompletionFunc("state", completeFieldValues("State"))
	_ = updateCmd.RegisterFlagCompletionFunc("priority", completeFieldValues("Priority"))
	_ = updateCmd.RegisterFlagCompletionFunc("type", completeFieldValues("Type"))
	_ = updateCmd.RegisterFlagCompletionFunc("subsystem", completeFieldValues("Subsystem"))
	_ = updateCmd.RegisterFlagCompletionFunc("field", completeFieldFlag(true))
}

func runIssueUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]

	client, err := apiFactory()
	if err != nil {
		return err
	}

	// REST API: summary and description
	restFields := make(map[string]string)
	if cmd.Flags().Changed("summary") {
		restFields["summary"] = updateSummary
	}
	if cmd.Flags().Changed("description") {
		restFields["description"] = updateDescription
	}

	// REST API: state (uses SetIssueState for reliable field-level update)
	stateChanged := cmd.Flags().Changed("state")

	// Command API: assignee, priority, type, subsystem, tags, fields
	assignee, err := resolveAssignee(client, updateAssignee)
	if err != nil {
		return err
	}

	fields := updateFields
	if cmd.Flags().Changed("subsystem") {
		fields = append(fields, "Subsystem="+updateSubsystem)
	}

	command, err := buildCommand("", assignee, updatePriority, updateType, updateTags, updateRemoveTags, fields)
	if err != nil {
		return err
	}

	if len(restFields) == 0 && !stateChanged && command == "" {
		return fmt.Errorf("no fields to update; use --summary, --description, --state, --assignee, --priority, --type, --subsystem, --tag, --field, or --remove-tag")
	}

	if len(restFields) > 0 {
		if err := client.UpdateIssueFields(id, restFields); err != nil {
			return err
		}
	}

	if stateChanged {
		if err := client.SetIssueState(id, updateState); err != nil {
			return err
		}
	}

	if command != "" {
		if err := client.UpdateIssue(id, command); err != nil {
			return err
		}
	}

	issue, err := client.GetIssue(id)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issue)
	}
	return format.Issue(w, issue)
}

// buildCommand constructs a YouTrack command string from field values.
// Multi-word values are wrapped in braces.
func buildCommand(state, assignee, priority, typ string, tags, removeTags, fields []string) (string, error) {
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
	for _, f := range fields {
		name, value, ok := strings.Cut(f, "=")
		if !ok {
			return "", fmt.Errorf("invalid --field format %q: expected Name=Value", f)
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" {
			return "", fmt.Errorf("invalid --field format %q: name must be non-empty", f)
		}
		parts = append(parts, braceWrap(name)+" "+braceWrap(value))
	}
	for _, t := range tags {
		parts = append(parts, "tag "+braceWrap(t))
	}
	for _, t := range removeTags {
		parts = append(parts, "untag "+braceWrap(t))
	}
	return strings.Join(parts, " "), nil
}

// braceWrap wraps s in braces if it contains spaces.
func braceWrap(s string) string {
	if strings.Contains(s, " ") {
		return "{" + s + "}"
	}
	return s
}
