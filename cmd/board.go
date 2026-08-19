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
	Long: `Show issues on an agile board's sprint. Looks up the board by name
(case-insensitive). Defaults to the current sprint.

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

  # JSON output
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

// boardFilter narrows a board's sprint issues.
type boardFilter struct {
	sprint   string
	state    string
	assignee string
	query    string
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

// showBoard writes one sprint's issues, honouring --json.
func showBoard(cmd *cobra.Command, name string, f boardFilter) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	board, err := client.GetBoardByName(name)
	if err != nil {
		return err
	}

	sprint, err := resolveSprint(board, f.sprint)
	if err != nil {
		return err
	}

	assignee, err := resolveAssignee(client, f.assignee)
	if err != nil {
		return err
	}

	boardPart := fmt.Sprintf("Board %s: {%s}", board.Name, sprint.Name)
	if f.query != "" {
		boardPart += " " + f.query
	}
	query := youtrack.BuildQuery("", f.state, assignee, boardPart)

	issues, err := client.ListIssues(query, 0)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issues)
	}
	return format.SprintIssues(w, board.Name, sprint.Name, issues)
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
