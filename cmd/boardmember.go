package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
)

// sprintChange records the outcome of adding or removing one issue from a
// board. Changed is false when nothing was written, and Note then says why:
// the issue was already in the wanted state, or the board's query or projects
// decide membership. Sprint is empty on a sprintless board.
type sprintChange struct {
	Issue   string `json:"issue"`
	Board   string `json:"board"`
	Sprint  string `json:"sprint,omitempty"`
	Changed bool   `json:"changed"`
	Note    string `json:"note,omitempty"`
}

// loadBoardSprint resolves a board by name, id or URL and the sprint to act on: the
// named sprint, or the board's current one. A sprintless board has one
// implicit sprint, which the API still needs, and takes no --sprint.
func loadBoardSprint(client youtrack.API, boardRef, sprintName string) (*youtrack.Agile, *youtrack.Sprint, error) {
	board, err := client.GetBoard(boardRef)
	if err != nil {
		return nil, nil, err
	}
	if board.Sprintless() && sprintName != "" {
		return nil, nil, fmt.Errorf("board %q has no sprints; it shows %s", board.Name, board.Membership())
	}
	sprint, err := resolveSprint(board, sprintName)
	if err != nil {
		return nil, nil, err
	}
	return board, sprint, nil
}

// automaticNote says why an issue on a query or projects board needed no
// adding.
func automaticNote(board *youtrack.Agile) string {
	if board.Kind() == youtrack.BoardProjects {
		return "on board as an issue of its project, nothing to add"
	}
	return "on board by its query, nothing to add"
}

// addIssuesToSprint adds each issue to the board's sprint, skipping issues
// already present. On a query or projects board nothing can be added: an
// issue the board already shows is reported as such, one it does not show is
// an error naming what would put it there. It returns one sprintChange per
// issue, in input order.
func addIssuesToSprint(client youtrack.API, board *youtrack.Agile, sprint *youtrack.Sprint, issues []string) ([]sprintChange, error) {
	present, err := sprintMembers(client, board, sprint)
	if err != nil {
		return nil, err
	}

	changes := make([]sprintChange, 0, len(issues))
	for _, id := range issues {
		change := newSprintChange(id, board, sprint)
		switch {
		case present[normalizeID(id)] && board.Automatic():
			change.Note = automaticNote(board)
		case present[normalizeID(id)]:
			change.Note = "already on board"
		case board.Automatic():
			return nil, fmt.Errorf("%s is not on board %q, which shows %s; change the issue to match instead of adding it",
				id, board.Name, board.Membership())
		default:
			if err := client.AddIssueToSprint(board.ID, sprint.ID, id); err != nil {
				return nil, err
			}
			change.Changed = true
			present[normalizeID(id)] = true
		}
		changes = append(changes, change)
	}
	return changes, nil
}

// removeIssuesFromSprint removes each issue from the board's sprint, skipping
// issues not present. A query or projects board takes no removals. It
// returns one sprintChange per issue, in input order.
func removeIssuesFromSprint(client youtrack.API, board *youtrack.Agile, sprint *youtrack.Sprint, issues []string) ([]sprintChange, error) {
	if board.Automatic() {
		return nil, fmt.Errorf("board %q shows %s; change an issue so it no longer matches instead of removing it",
			board.Name, board.Membership())
	}
	present, err := sprintMembers(client, board, sprint)
	if err != nil {
		return nil, err
	}

	changes := make([]sprintChange, 0, len(issues))
	for _, id := range issues {
		change := newSprintChange(id, board, sprint)
		if !present[normalizeID(id)] {
			change.Note = "not on board"
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

func newSprintChange(id string, board *youtrack.Agile, sprint *youtrack.Sprint) sprintChange {
	c := sprintChange{Issue: id, Board: board.Name}
	if !board.Sprintless() {
		c.Sprint = sprint.Name
	}
	return c
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

// mirrorBoards copies fromID's board and sprint memberships onto toID and
// returns how many were applied. Query and projects boards are skipped: the
// new issue is on them exactly when it matches, which nothing here changes. A
// zero count is not an error here; callers that require at least one board
// decide how to treat it.
func mirrorBoards(client youtrack.API, fromID, toID string) (int, error) {
	memberships, err := client.IssueBoards(fromID)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, m := range memberships {
		if m.Kind == youtrack.BoardQuery || m.Kind == youtrack.BoardProjects {
			continue
		}
		if err := placeOnBoard(client, toID, m.Board, m.Sprint); err != nil {
			return 0, err
		}
		n++
	}
	return n, nil
}

// linkAsSubtask makes childID a subtask of parentID. The standard YouTrack
// "subtask of" relation is resolved against the instance's link types.
func linkAsSubtask(client youtrack.API, childID, parentID string) error {
	rel, err := resolveRelation(client, "subtask-of")
	if err != nil {
		return err
	}
	return client.CreateLink(childID, rel.Phrase, parentID)
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
