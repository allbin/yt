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
or --like to mirror another issue's board and sprint -- handy for putting a
subtask on the same board as its parent.`,
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

  # put a subtask on the same board+sprint as its parent
  yt issue create -p AX -s "Subtask" --like AX-332

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
	if createLike != "" {
		if err := mirrorBoards(client, createLike, issue.IDReadable); err != nil {
			return err
		}
		placed = true
	}
	if createBoard != "" {
		if err := placeOnBoard(client, issue.IDReadable, createBoard, createSprint); err != nil {
			return err
		}
		placed = true
	}

	if command != "" || placed {
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
