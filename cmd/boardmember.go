package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
)

// sprintChange records the outcome of adding or removing one issue from a
// board's sprint. Changed is false when the operation was a no-op (the issue
// was already in the desired state).
type sprintChange struct {
	Issue   string `json:"issue"`
	Board   string `json:"board"`
	Sprint  string `json:"sprint"`
	Changed bool   `json:"changed"`
}

// loadBoardSprint resolves a board by name and the target sprint (the named
// sprint, or the board's current sprint when name is empty).
func loadBoardSprint(client youtrack.API, boardName, sprintName string) (*youtrack.Agile, *youtrack.Sprint, error) {
	board, err := client.GetBoardByName(boardName)
	if err != nil {
		return nil, nil, err
	}
	sprint, err := resolveSprint(board, sprintName)
	if err != nil {
		return nil, nil, err
	}
	return board, sprint, nil
}

// addIssuesToSprint adds each issue to the sprint, skipping issues already
// present. It returns one sprintChange per issue, in input order.
func addIssuesToSprint(client youtrack.API, board *youtrack.Agile, sprint *youtrack.Sprint, issues []string) ([]sprintChange, error) {
	present, err := sprintMembers(client, board, sprint)
	if err != nil {
		return nil, err
	}

	changes := make([]sprintChange, 0, len(issues))
	for _, id := range issues {
		change := sprintChange{Issue: id, Board: board.Name, Sprint: sprint.Name}
		if present[normalizeID(id)] {
			changes = append(changes, change)
			continue
		}
		if err := client.AddIssueToSprint(board.ID, sprint.ID, id); err != nil {
			return nil, err
		}
		change.Changed = true
		present[normalizeID(id)] = true
		changes = append(changes, change)
	}
	return changes, nil
}

// removeIssuesFromSprint removes each issue from the sprint, skipping issues
// not present. It returns one sprintChange per issue, in input order.
func removeIssuesFromSprint(client youtrack.API, board *youtrack.Agile, sprint *youtrack.Sprint, issues []string) ([]sprintChange, error) {
	present, err := sprintMembers(client, board, sprint)
	if err != nil {
		return nil, err
	}

	changes := make([]sprintChange, 0, len(issues))
	for _, id := range issues {
		change := sprintChange{Issue: id, Board: board.Name, Sprint: sprint.Name}
		if !present[normalizeID(id)] {
			changes = append(changes, change)
			continue
		}
		if err := client.RemoveIssueFromSprint(board.ID, sprint.ID, id); err != nil {
			return nil, err
		}
		change.Changed = true
		present[normalizeID(id)] = false
		changes = append(changes, change)
	}
	return changes, nil
}

// placeOnBoard adds a single issue to the named board's sprint (the current
// sprint when sprintName is empty).
func placeOnBoard(client youtrack.API, issueID, boardName, sprintName string) error {
	board, sprint, err := loadBoardSprint(client, boardName, sprintName)
	if err != nil {
		return err
	}
	_, err = addIssuesToSprint(client, board, sprint, []string{issueID})
	return err
}

// mirrorBoards copies fromID's board and sprint memberships onto toID.
func mirrorBoards(client youtrack.API, fromID, toID string) error {
	memberships, err := client.IssueBoards(fromID)
	if err != nil {
		return err
	}
	if len(memberships) == 0 {
		return fmt.Errorf("%s is not on any board", fromID)
	}
	for _, m := range memberships {
		if err := placeOnBoard(client, toID, m.Board, m.Sprint); err != nil {
			return err
		}
	}
	return nil
}

// sprintMembers returns the sprint's current issues as a normalized-ID set.
func sprintMembers(client youtrack.API, board *youtrack.Agile, sprint *youtrack.Sprint) (map[string]bool, error) {
	ids, err := client.ListSprintIssues(board.ID, sprint.ID)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[normalizeID(id)] = true
	}
	return set, nil
}

func normalizeID(id string) string {
	return strings.ToUpper(id)
}
