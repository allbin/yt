package cmd

import (
	"fmt"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var (
	createProject     string
	createSummary     string
	createDescription string
	createSubsystem   string
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

Use --subsystem or --field to set custom fields on the new issue.

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

  # create with subsystem
  yt issue create -p PROJ -s "Fix API auth" --subsystem API

  # create with custom field
  yt issue create -p PROJ -s "Critical outage" --field "Severity=Critical"

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
	createCmd.Flags().StringSliceVarP(&createTags, "tag", "t", nil, "add tag (repeatable)")
	createCmd.Flags().StringSliceVar(&createFields, "field", nil, `set custom field as "Name=Value" (repeatable)`)
	createCmd.Flags().StringVar(&createBoard, "board", "", "add the issue to this agile board")
	createCmd.Flags().StringVar(&createSprint, "sprint", "", "sprint for --board (default: current)")
	createCmd.Flags().StringVar(&createLike, "like", "", "mirror another issue's board and sprint")
	createCmd.Flags().StringVar(&createParent, "parent", "", "make the issue a subtask of this issue and share its board")
	createCmd.MarkFlagsMutuallyExclusive("parent", "like")
	_ = createCmd.MarkFlagRequired("project")
	_ = createCmd.MarkFlagRequired("summary")

	_ = createCmd.RegisterFlagCompletionFunc("subsystem", completeProjectFieldValues("Subsystem"))
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

	issue, err := client.CreateIssue(createProject, createSummary, description, createTags)
	if err != nil {
		return err
	}

	fields := createFields
	if cmd.Flags().Changed("subsystem") {
		fields = append(fields, "Subsystem="+createSubsystem)
	}

	command, err := buildCommand("", "", "", "", nil, nil, fields)
	if err != nil {
		return err
	}
	if command != "" {
		if err := client.UpdateIssue(issue.IDReadable, command); err != nil {
			return fmt.Errorf("set fields on %s: %w", issue.IDReadable, err)
		}
	}

	placed := false
	if createParent != "" {
		if err := linkAsSubtask(client, issue.IDReadable, createParent); err != nil {
			return err
		}
		// Share the parent's board unless an explicit --board overrides it.
		// A parent on no board is fine: the subtask link still stands.
		if createBoard == "" {
			n, err := mirrorBoards(client, createParent, issue.IDReadable)
			if err != nil {
				return err
			}
			if n > 0 {
				placed = true
			}
		}
	}
	if createLike != "" {
		n, err := mirrorBoards(client, createLike, issue.IDReadable)
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("%s is not on any board", createLike)
		}
		placed = true
	}
	if createBoard != "" {
		if err := placeOnBoard(client, issue.IDReadable, createBoard, createSprint); err != nil {
			return err
		}
		placed = true
	}

	if command != "" || placed || createParent != "" {
		issue, err = client.GetIssue(issue.IDReadable)
		if err != nil {
			return err
		}
	}
	if placed {
		// Best-effort: show resulting board membership; ignore lookup failures.
		if boards, err := client.IssueBoards(issue.IDReadable); err == nil {
			issue.Boards = boards
		}
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issue)
	}
	return format.Issue(w, issue)
}
