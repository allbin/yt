package boarddata

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
)

// Result is the normalized output regardless of fetch path.
type Result struct {
	Board       *youtrack.Agile
	Issues      []youtrack.Issue
	SprintBoard *youtrack.SprintBoard
}

// Fetcher abstracts the two board data fetch paths.
type Fetcher struct {
	client youtrack.API
}

// New creates a Fetcher with the given API client.
func New(client youtrack.API) *Fetcher {
	return &Fetcher{client: client}
}

// Load fetches board metadata + issues.
func (f *Fetcher) Load(boardName, sprintName string) (Result, error) {
	board, err := f.client.GetBoardForView(boardName)
	if err != nil {
		return Result{}, err
	}
	r, err := f.fetchIssues(board, sprintName)
	if err != nil {
		return Result{Board: board}, err
	}
	r.Board = board
	return r, nil
}

// Refresh re-fetches issues for an already-loaded board.
func (f *Fetcher) Refresh(board *youtrack.Agile, sprintName string) (Result, error) {
	return f.fetchIssues(board, sprintName)
}

// SetStateAndRefresh applies a state change then refreshes.
func (f *Fetcher) SetStateAndRefresh(issueID, state string, board *youtrack.Agile, sprintName string) (Result, error) {
	if err := f.client.SetIssueState(issueID, state); err != nil {
		return Result{}, err
	}
	return f.fetchIssues(board, sprintName)
}

func (f *Fetcher) fetchIssues(board *youtrack.Agile, sprintName string) (Result, error) {
	if swimlanesEnabled(board) {
		if sprintID := resolveSprintID(board, sprintName); sprintID != "" {
			sb, err := f.client.GetSprintBoard(board.ID, sprintID)
			if err != nil {
				return Result{}, err
			}
			return Result{Issues: flattenSprintBoard(sb), SprintBoard: sb}, nil
		}
	}

	sprint := resolveSprint(board, sprintName)
	query := fmt.Sprintf("Board %s: {%s}", board.Name, sprint)
	issues, err := f.client.ListIssues(query, 0)
	if err != nil {
		return Result{}, err
	}
	return Result{Issues: issues}, nil
}

func swimlanesEnabled(board *youtrack.Agile) bool {
	return board.SwimlaneSettings != nil && board.SwimlaneSettings.Enabled
}

func resolveSprintID(board *youtrack.Agile, sprintName string) string {
	if sprintName == "" {
		if board.CurrentSprint != nil {
			return board.CurrentSprint.ID
		}
		return ""
	}
	for _, s := range board.Sprints {
		if strings.EqualFold(s.Name, sprintName) {
			return s.ID
		}
	}
	return ""
}

func resolveSprint(board *youtrack.Agile, sprintName string) string {
	if sprintName != "" {
		return sprintName
	}
	if board.CurrentSprint != nil {
		return board.CurrentSprint.Name
	}
	return ""
}

func flattenSprintBoard(sb *youtrack.SprintBoard) []youtrack.Issue {
	var out []youtrack.Issue
	for _, col := range sb.Columns {
		for _, cell := range col.Cells {
			out = append(out, cell.Issues...)
		}
	}
	return out
}
