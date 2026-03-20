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

func runBoard(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	client, err := apiFactory()
	if err != nil {
		return err
	}

	board, err := client.GetBoardByName(args[0])
	if err != nil {
		return err
	}

	sprintName, err := resolveSprintName(board)
	if err != nil {
		return err
	}

	assignee, err := resolveAssignee(client, boardAssignee)
	if err != nil {
		return err
	}

	boardPart := fmt.Sprintf("Board %s: {%s}", board.Name, sprintName)
	if boardQuery != "" {
		boardPart += " " + boardQuery
	}
	query := youtrack.BuildQuery("", boardState, assignee, boardPart)

	issues, err := client.ListIssues(query, 0)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issues)
	}
	return format.SprintIssues(w, board.Name, sprintName, issues)
}

func resolveSprintName(board *youtrack.Agile) (string, error) {
	if boardSprint != "" {
		for _, s := range board.Sprints {
			if strings.EqualFold(s.Name, boardSprint) {
				return s.Name, nil
			}
		}
		return "", fmt.Errorf("sprint %q not found on board %q", boardSprint, board.Name)
	}
	if board.CurrentSprint == nil {
		return "", fmt.Errorf("board %q has no current sprint", board.Name)
	}
	return board.CurrentSprint.Name, nil
}
