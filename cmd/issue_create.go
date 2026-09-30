package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var (
	createProject     string
	createSummary     string
	createDescription string
	createSubsystem   string
	createType        string
	createPriority    string
	createAssignee    string
	createLinks       []string
	createTags        []string
	createFields      []string
	createBoard       string
	createSprint      string
	createLike        string
	createParent      string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new YouTrack issue",
	Long: `Create a new issue in the specified YouTrack project. Requires a project
short name and summary. Optionally accepts a description.

The created issue is displayed after creation.

Use --field "Name=Value" to set any custom field; --type, --priority,
--assignee and --subsystem are shorthands for those fields. Repeat --field
with the same name to set several values on a multi-value field. Values are
checked against the project's allowed values ("yt project fields PROJ") and
sent with the create request, so a bad value creates nothing.

Use --tag to tag the issue; a tag that does not exist yet is created. Use
--link "relation=ID" to link it to other issues ("yt link types" lists the
relations).

Tags, links and board placement are applied after the issue exists. If one of
them fails, the created issue is still printed and the error names its ID:
finish it with "yt issue update" or "yt link" rather than creating it again.

The description accepts "@path" to read from a file or "-" to read from stdin,
which avoids shell mangling of multi-line text.

Use --board (with optional --sprint) to place the new issue on an agile board,
or --like to mirror another issue's board and sprint.

Use --parent <id> to create the issue as a subtask of another issue: it adds
a "subtask of" link to the parent and places the new issue on the parent's
board and sprint in one step -- the common "subtask on the parent's board"
workflow. When --board is also given it overrides the parent's board.`,
	Example: `  # create a minimal issue
  yt issue create -p PROJ -s "Fix login bug"

  # create with description
  yt issue create -p PROJ -s "Add dark mode" -d "Support system-level dark mode preference"

  # read the description from a file
  yt issue create -p PROJ -s "Big writeup" -d @notes.md

  # read the description from stdin
  cat notes.md | yt issue create -p PROJ -s "Big writeup" -d -

  # set type and subsystem (quote multi-word values)
  yt issue create -p PROJ -s "Fix API auth" --type "User Story" --subsystem "Management UI"

  # create with custom field
  yt issue create -p PROJ -s "Critical outage" --field "Severity=Critical"

  # several values on a multi-value field
  yt issue create -p PROJ -s "Shared fix" --field "Subsystem=API" --field "Subsystem=Management UI"

  # link to other issues on creation
  yt issue create -p PROJ -s "Follow-up" --link depends-on=PROJ-12 --link relates=PROJ-7

  # create with tags
  yt issue create -p PROJ -s "Fix stale state" -t tech-debt -t scheduler

  # place the new issue on a board's current sprint
  yt issue create -p AX -s "Subtask" --board AllTix

  # mirror another issue's board+sprint without linking
  yt issue create -p AX -s "Subtask" --like AX-332

  # create a subtask: link to the parent AND share its board+sprint
  yt issue create -p AX -s "Subtask" --parent AX-332

  # output as JSON
  yt issue create -p PROJ -s "New feature" --json`,
	RunE: runIssueCreate,
}

func init() {
	issueCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createProject, "project", "p", "", "project short name (required)")
	createCmd.Flags().StringVarP(&createSummary, "summary", "s", "", "issue summary (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "issue description (@file or - for stdin)")
	createCmd.Flags().StringVar(&createSubsystem, "subsystem", "", "set subsystem")
	createCmd.Flags().StringVar(&createType, "type", "", "set issue type")
	createCmd.Flags().StringVar(&createPriority, "priority", "", "set priority")
	createCmd.Flags().StringVar(&createAssignee, "assignee", "", "set assignee (supports 'me')")
	createCmd.Flags().StringSliceVarP(&createTags, "tag", "t", nil, "add tag (repeatable)")
	createCmd.Flags().StringArrayVar(&createFields, "field", nil, `set custom field as "Name=Value" (repeatable)`)
	createCmd.Flags().StringArrayVar(&createLinks, "link", nil, `link to an issue as "relation=ID", e.g. depends-on=AX-3 (repeatable)`)
	createCmd.Flags().StringVar(&createBoard, "board", "", "add the issue to this agile board")
	createCmd.Flags().StringVar(&createSprint, "sprint", "", "sprint for --board (default: current)")
	createCmd.Flags().StringVar(&createLike, "like", "", "mirror another issue's board and sprint")
	createCmd.Flags().StringVar(&createParent, "parent", "", "make the issue a subtask of this issue and share its board")
	createCmd.MarkFlagsMutuallyExclusive("parent", "like")
	_ = createCmd.MarkFlagRequired("project")
	_ = createCmd.MarkFlagRequired("summary")

	_ = createCmd.RegisterFlagCompletionFunc("subsystem", completeProjectFieldValues("Subsystem"))
	_ = createCmd.RegisterFlagCompletionFunc("type", completeProjectFieldValues("Type"))
	_ = createCmd.RegisterFlagCompletionFunc("priority", completeProjectFieldValues("Priority"))
	_ = createCmd.RegisterFlagCompletionFunc("field", completeFieldFlag(false))
}

func runIssueCreate(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	description, err := readTextArg(createDescription, cmd.InOrStdin())
	if err != nil {
		return err
	}

	// Everything that can be checked is resolved before creating, so a bad
	// field value or relation leaves no orphan issue.
	updates, err := createFieldUpdates(cmd, client)
	if err != nil {
		return err
	}
	links, err := resolveLinkArgs(client, createLinks)
	if err != nil {
		return err
	}
	command := tagCommand(createTags, nil)

	issue, err := client.CreateIssue(createProject, createSummary, description, updates)
	if err != nil {
		return err
	}

	// From here on the issue exists. A failed step still prints it and names
	// its ID, so the caller finishes it with update instead of creating a
	// duplicate.
	issue, stepErr := completeCreate(client, issue, command, links)

	w := cmd.OutOrStdout()
	if jsonOutput {
		err = format.JSON(w, issue)
	} else {
		err = format.Issue(w, issue)
	}
	if err != nil {
		return err
	}
	if stepErr != nil {
		return fmt.Errorf("%s was created, but a follow-up step failed; finish it with `yt issue update %s`, do not create it again: %w",
			issue.IDReadable, issue.IDReadable, stepErr)
	}
	return nil
}

// createFieldUpdates resolves --field and the field shorthands against the
// project's fields.
func createFieldUpdates(cmd *cobra.Command, client youtrack.API) ([]youtrack.FieldUpdate, error) {
	rawFields, err := parseFieldArgs(createFields)
	if err != nil {
		return nil, err
	}
	args := append(flagFieldArgs(cmd.Flags(),
		[2]string{"type", "Type"},
		[2]string{"priority", "Priority"},
		[2]string{"assignee", "Assignee"},
		[2]string{"subsystem", "Subsystem"},
	), rawFields...)
	if len(args) == 0 {
		return nil, nil
	}
	schema, err := client.ListProjectFields(createProject)
	if err != nil {
		return nil, err
	}
	return fieldUpdates(client, schema, fieldChanges{set: args}, nil)
}

// linkArg is a resolved --link: the relation phrase and its target.
type linkArg struct {
	phrase, target string
}

// resolveLinkArgs parses "relation=ID" values and resolves each relation
// against the instance's link types.
func resolveLinkArgs(client youtrack.API, raw []string) ([]linkArg, error) {
	links := make([]linkArg, 0, len(raw))
	for _, l := range raw {
		alias, target, ok := strings.Cut(l, "=")
		alias, target = strings.TrimSpace(alias), strings.TrimSpace(target)
		if !ok || alias == "" || target == "" {
			return nil, fmt.Errorf("invalid --link %q: expected relation=ID, e.g. depends-on=AX-3", l)
		}
		rel, err := resolveRelation(client, alias)
		if err != nil {
			return nil, err
		}
		links = append(links, linkArg{phrase: rel.Phrase, target: target})
	}
	return links, nil
}

// completeCreate applies tags, links and board placement to a newly created
// issue. It always returns the best-known state of the issue, alongside the
// first step that failed.
func completeCreate(client youtrack.API, created *youtrack.Issue, command string, links []linkArg) (*youtrack.Issue, error) {
	id := created.IDReadable
	placed, err := applyCreateSteps(client, id, command, links)
	if command == "" && !placed && createParent == "" && len(links) == 0 && err == nil {
		return created, nil
	}

	issue, fetchErr := client.GetIssue(id)
	if fetchErr != nil {
		return created, errors.Join(err, fetchErr)
	}
	if placed {
		// Best-effort: show resulting board membership; ignore lookup failures.
		if boards, err := client.IssueBoards(id); err == nil {
			issue.Boards = boards
		}
	}
	return issue, err
}

// applyCreateSteps runs the post-create steps in order, stopping at the first
// failure. placed reports whether the issue landed on a board.
func applyCreateSteps(client youtrack.API, id, command string, links []linkArg) (placed bool, err error) {
	if command != "" {
		if err := client.UpdateIssue(id, command); err != nil {
			return false, fmt.Errorf("add tags: %w", err)
		}
	}
	for _, l := range links {
		if err := client.CreateLink(id, l.phrase, l.target); err != nil {
			return false, fmt.Errorf("link %s %s: %w", l.phrase, l.target, err)
		}
	}

	if createParent != "" {
		if err := linkAsSubtask(client, id, createParent); err != nil {
			return false, err
		}
		// Share the parent's board unless an explicit --board overrides it.
		// A parent on no board is fine: the subtask link still stands.
		if createBoard == "" {
			n, err := mirrorBoards(client, createParent, id)
			if err != nil {
				return false, err
			}
			placed = n > 0
		}
	}
	if createLike != "" {
		n, err := mirrorBoards(client, createLike, id)
		if err != nil {
			return placed, err
		}
		if n == 0 {
			return placed, fmt.Errorf("%s is not on any board", createLike)
		}
		placed = true
	}
	if createBoard != "" {
		if err := placeOnBoard(client, id, createBoard, createSprint); err != nil {
			return placed, err
		}
		placed = true
	}
	return placed, nil
}
