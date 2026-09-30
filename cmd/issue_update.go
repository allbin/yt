package cmd

import (
	"fmt"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
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
	updateBoard       string
	updateSprint      string
)

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a YouTrack issue",
	Long: `Update fields on a YouTrack issue.

Use --field to set any custom field by name, as "Name=Value". Values with
spaces need no escaping beyond shell quoting. Repeat --field with the same name
to set several values on a multi-value field; the list replaces the current
values. An empty value clears the field. Values are checked against the
field's allowed values before anything is written; "yt project fields PROJ"
lists them.

--assignee, --priority, --type and --subsystem are shorthands for --field on
those fields. Assignee accepts "me", a login or a name.

The description accepts "@path" to read from a file or "-" to read from stdin,
which avoids shell mangling of multi-line text.

Use --board (with optional --sprint) to also place the issue on an agile board.

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

  # set a multi-word value
  yt issue update PROJ-123 -t "User Story"

  # set subsystem
  yt issue update PROJ-123 --subsystem "Management UI"

  # set several values on a multi-value field
  yt issue update PROJ-123 --field "Subsystem=API" --field "Subsystem=Management UI"

  # clear a field
  yt issue update PROJ-123 --field "Subsystem="

  # set arbitrary custom field
  yt issue update PROJ-123 --field "Severity=Critical"

  # add tags
  yt issue update PROJ-123 --tag tech-debt --tag scheduler

  # remove a tag
  yt issue update PROJ-123 --remove-tag obsolete

  # read the description from a file
  yt issue update PROJ-123 -d @notes.md

  # place the issue on a board's current sprint
  yt issue update AX-812 --board AllTix

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
	updateCmd.Flags().StringArrayVar(&updateFields, "field", nil, `set custom field as "Name=Value" (repeatable)`)
	updateCmd.Flags().StringVar(&updateBoard, "board", "", "add the issue to this agile board")
	updateCmd.Flags().StringVar(&updateSprint, "sprint", "", "sprint for --board (default: current)")

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
		description, err := readTextArg(updateDescription, cmd.InOrStdin())
		if err != nil {
			return err
		}
		restFields["description"] = description
	}

	stateChanged := cmd.Flags().Changed("state")

	rawFields, err := parseFieldArgs(updateFields)
	if err != nil {
		return err
	}
	fieldArgs := append(flagFieldArgs(cmd.Flags(),
		[2]string{"assignee", "Assignee"},
		[2]string{"priority", "Priority"},
		[2]string{"type", "Type"},
		[2]string{"subsystem", "Subsystem"},
	), rawFields...)

	command := tagCommand(updateTags, updateRemoveTags)
	boardChanged := cmd.Flags().Changed("board")

	if len(restFields) == 0 && !stateChanged && len(fieldArgs) == 0 && command == "" && !boardChanged {
		return fmt.Errorf("no fields to update; use --summary, --description, --state, --assignee, --priority, --type, --subsystem, --tag, --field, --remove-tag, or --board")
	}

	// Resolved before any write, so a bad value changes nothing.
	var updates []youtrack.FieldUpdate
	if len(fieldArgs) > 0 {
		schema, err := client.ListIssueFields(id)
		if err != nil {
			return err
		}
		if updates, err = fieldUpdates(client, schema, fieldArgs); err != nil {
			return err
		}
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

	if len(updates) > 0 {
		if err := client.SetIssueFields(id, updates); err != nil {
			return err
		}
	}

	if command != "" {
		if err := client.UpdateIssue(id, command); err != nil {
			return err
		}
	}

	if boardChanged {
		if err := placeOnBoard(client, id, updateBoard, updateSprint); err != nil {
			return err
		}
	}

	issue, err := client.GetIssue(id)
	if err != nil {
		return err
	}
	if boardChanged {
		// Best-effort: show resulting board membership; ignore lookup failures.
		if boards, err := client.IssueBoards(id); err == nil {
			issue.Boards = boards
		}
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issue)
	}
	return format.Issue(w, issue)
}
