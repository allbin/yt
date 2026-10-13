package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var (
	boardSprint   string
	boardState    string
	boardAssignee string
	boardQuery    string
)

var boardCmd = &cobra.Command{
	Use:   "board [name]",
	Short: "Show board issues or list boards",
	Long: `Show the issues on an agile board. Looks up the board by name
(case-insensitive).

A board is one of four kinds, by what puts an issue on it: "sprints" (the
issue is assigned to one of the board's sprints; shown for the current sprint
unless --sprint names another), "query" (the issue matches the board's query),
"projects" (every issue of the board's projects) or "manual" (added by hand).
The last three have no sprints. --json prints the kind, what membership means
for it, the sprint shown, and the issues.

Filters are handled server-side by YouTrack, so assignee supports "me",
login names, and full names.

Use subcommands to list available boards.`,
	Example: `  # show current sprint issues
  yt board HållKoll

  # specific sprint
  yt board HållKoll --sprint 2025-02

  # filter by state
  yt board HållKoll -s "In Progress"

  # issues assigned to me
  yt board HållKoll -a me

  # combine filters with extra query
  yt board HållKoll -a me -q "sort by: Priority"

  # JSON output: kind, membership, sprint and issues
  yt board HållKoll --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBoard,
}

func init() {
	rootCmd.AddCommand(boardCmd)
	boardCmd.Flags().StringVar(&boardSprint, "sprint", "", "sprint name (default: current)")
	boardCmd.Flags().StringVarP(&boardState, "state", "s", "", "filter by state")
	boardCmd.Flags().StringVarP(&boardAssignee, "assignee", "a", "", "filter by assignee (supports 'me')")
	boardCmd.Flags().StringVarP(&boardQuery, "query", "q", "", "additional YouTrack query")
}

// boardFilter narrows a board's issues.
type boardFilter struct {
	sprint   string
	state    string
	assignee string
	query    string
}

// boardView is the JSON form of "yt board <name>": the board's kind and what
// membership means for it, ahead of the issues shown.
type boardView struct {
	Name        string           `json:"name"`
	Kind        string           `json:"kind"`
	Membership  string           `json:"membership"`
	Query       string           `json:"query,omitempty"`
	ColumnField string           `json:"columnField,omitempty"`
	Projects    []string         `json:"projects"`
	Sprint      *youtrack.Sprint `json:"sprint"`
	Issues      []youtrack.Issue `json:"issues"`
}

func runBoard(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	return showBoard(cmd, args[0], boardFilter{
		sprint:   boardSprint,
		state:    boardState,
		assignee: boardAssignee,
		query:    boardQuery,
	})
}

// showBoard writes a board's issues, honouring --json. On a sprint-based
// board these are one sprint's issues; a sprintless board has only the one
// set.
func showBoard(cmd *cobra.Command, name string, f boardFilter) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	board, err := client.GetBoardByName(name)
	if err != nil {
		return err
	}

	sprint, err := shownSprint(board, f.sprint)
	if err != nil {
		return err
	}

	assignee, err := resolveAssignee(client, f.assignee)
	if err != nil {
		return err
	}

	sprintName := ""
	if sprint != nil {
		sprintName = sprint.Name
	}
	boardPart := board.SearchQuery(sprintName)
	if f.query != "" {
		boardPart += " " + f.query
	}
	query := youtrack.BuildQuery("", f.state, assignee, boardPart)

	issues, err := client.ListIssues(query, 0)
	if err != nil {
		return err
	}
	if issues == nil {
		issues = []youtrack.Issue{}
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, boardView{
			Name:        board.Name,
			Kind:        board.Kind(),
			Membership:  board.Membership(),
			Query:       board.Query(),
			ColumnField: board.ColumnField(),
			Projects:    board.ProjectNames(),
			Sprint:      sprint,
			Issues:      issues,
		})
	}
	return format.BoardIssues(w, board, sprint, issues)
}

// shownSprint is the sprint a board view lists: the named or current sprint
// of a sprint-based board, none for a sprintless one, which cannot take
// --sprint.
func shownSprint(board *youtrack.Agile, name string) (*youtrack.Sprint, error) {
	if !board.Sprintless() {
		return resolveSprint(board, name)
	}
	if name != "" {
		return nil, fmt.Errorf("board %q has no sprints; it shows %s", board.Name, board.Membership())
	}
	return nil, nil
}

// resolveSprint returns the named sprint on the board, or the board's current
// sprint when name is empty. Matching is case-insensitive.
func resolveSprint(board *youtrack.Agile, name string) (*youtrack.Sprint, error) {
	if name != "" {
		for i := range board.Sprints {
			if strings.EqualFold(board.Sprints[i].Name, name) {
				return &board.Sprints[i], nil
			}
		}
		if board.CurrentSprint != nil && strings.EqualFold(board.CurrentSprint.Name, name) {
			return board.CurrentSprint, nil
		}
		return nil, fmt.Errorf("sprint %q not found on board %q", name, board.Name)
	}
	if board.CurrentSprint == nil {
		return nil, fmt.Errorf("board %q has no current sprint", board.Name)
	}
	return board.CurrentSprint, nil
}
