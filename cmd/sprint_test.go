package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func testBoard() *youtrack.Agile {
	// Noon-UTC millis so the rendered date is timezone-stable in tests.
	start := int64(1748779200000) // 2025-06-01 12:00 UTC
	finish := int64(1749988800000)
	return &youtrack.Agile{
		ID:            "A1",
		Name:          "AllTix",
		Projects:      []youtrack.Project{{ShortName: "AX"}},
		CurrentSprint: &youtrack.Sprint{ID: "S2", Name: "2025-06"},
		Sprints: []youtrack.Sprint{
			{ID: "S1", Name: "2025-05"},
			{ID: "S2", Name: "2025-06", Start: &start, Finish: &finish},
		},
	}
}

func TestRunSprintList(t *testing.T) {
	run := setupTest(t, &mockAPI{board: testBoard()})

	out, err := run("sprint", "list", "AllTix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "2025-05") || !strings.Contains(out, "2025-06") {
		t.Errorf("output missing sprints: %s", out)
	}
	if !strings.Contains(out, "current") {
		t.Errorf("output missing current marker: %s", out)
	}
	if !strings.Contains(out, "→") {
		t.Errorf("output missing sprint date range: %s", out)
	}
}

func TestRunSprintListJSON(t *testing.T) {
	run := setupTest(t, &mockAPI{board: testBoard()})

	out, err := run("sprint", "list", "AllTix", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []youtrack.Sprint
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(got) != 2 {
		t.Errorf("got %d sprints, want 2", len(got))
	}
}

func TestRunBoardAddCurrentSprint(t *testing.T) {
	mock := &mockAPI{board: testBoard()}
	run := setupTest(t, mock)

	out, err := run("board", "add", "AllTix", "AX-812")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S2|AX-812" {
		t.Errorf("addedToSprint = %v, want [A1|S2|AX-812]", mock.addedToSprint)
	}
	if !strings.Contains(out, "AX-812") || !strings.Contains(out, "AllTix") || !strings.Contains(out, "2025-06") {
		t.Errorf("output missing details: %s", out)
	}
}

func TestRunBoardAddSpecificSprint(t *testing.T) {
	mock := &mockAPI{board: testBoard()}
	run := setupTest(t, mock)

	_, err := run("board", "add", "AllTix", "AX-1", "--sprint", "2025-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S1|AX-1" {
		t.Errorf("addedToSprint = %v, want [A1|S1|AX-1]", mock.addedToSprint)
	}
}

func TestRunBoardAddAlreadyOnBoard(t *testing.T) {
	mock := &mockAPI{
		board:        testBoard(),
		sprintIssues: map[string][]string{"A1|S2": {"AX-812"}},
	}
	run := setupTest(t, mock)

	out, err := run("board", "add", "AllTix", "ax-812") // case-insensitive
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 0 {
		t.Errorf("expected no add, got %v", mock.addedToSprint)
	}
	if !strings.Contains(out, "already on board") {
		t.Errorf("output missing 'already on board': %s", out)
	}
}

func TestRunBoardAddMultiple(t *testing.T) {
	mock := &mockAPI{
		board:        testBoard(),
		sprintIssues: map[string][]string{"A1|S2": {"AX-1"}},
	}
	run := setupTest(t, mock)

	_, err := run("board", "add", "AllTix", "AX-1", "AX-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// AX-1 already present, only AX-2 added.
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S2|AX-2" {
		t.Errorf("addedToSprint = %v, want [A1|S2|AX-2]", mock.addedToSprint)
	}
}

func TestRunBoardAddJSON(t *testing.T) {
	mock := &mockAPI{board: testBoard()}
	run := setupTest(t, mock)

	out, err := run("board", "add", "AllTix", "AX-812", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []sprintChange
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(got) != 1 || !got[0].Changed || got[0].Sprint != "2025-06" {
		t.Errorf("got %+v", got)
	}
}

func TestRunBoardRemove(t *testing.T) {
	mock := &mockAPI{
		board:        testBoard(),
		sprintIssues: map[string][]string{"A1|S2": {"AX-812"}},
	}
	run := setupTest(t, mock)

	out, err := run("board", "remove", "AllTix", "AX-812")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.removedSprint) != 1 || mock.removedSprint[0] != "A1|S2|AX-812" {
		t.Errorf("removedSprint = %v, want [A1|S2|AX-812]", mock.removedSprint)
	}
	if !strings.Contains(out, "removed") {
		t.Errorf("output missing 'removed': %s", out)
	}
}

func TestRunBoardRemoveNotOnBoard(t *testing.T) {
	mock := &mockAPI{board: testBoard()}
	run := setupTest(t, mock)

	out, err := run("board", "remove", "AllTix", "AX-999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.removedSprint) != 0 {
		t.Errorf("expected no remove, got %v", mock.removedSprint)
	}
	if !strings.Contains(out, "not on board") {
		t.Errorf("output missing 'not on board': %s", out)
	}
}

func TestRunBoardAddUnknownSprint(t *testing.T) {
	mock := &mockAPI{board: testBoard()}
	run := setupTest(t, mock)

	_, err := run("board", "add", "AllTix", "AX-1", "--sprint", "nope")
	if err == nil {
		t.Fatal("expected error for unknown sprint")
	}
	if !strings.Contains(err.Error(), "sprint \"nope\" not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

// queryBoard is a sprintless board whose query decides membership, like a
// "Subsystem: Webticket" board. The API still exposes one implicit sprint.
func queryBoard() *youtrack.Agile {
	return &youtrack.Agile{
		ID:              "A5",
		Name:            "AllTix Webticket",
		Projects:        []youtrack.Project{{ShortName: "AX"}},
		CurrentSprint:   &youtrack.Sprint{ID: "S0", Name: "First sprint"},
		SprintsSettings: &youtrack.SprintsSettings{DisableSprints: true, ExplicitQuery: "Subsystem: Webticket"},
	}
}

func TestRunBoardAddQueryBoardAlreadyShown(t *testing.T) {
	mock := &mockAPI{
		board:        queryBoard(),
		sprintIssues: map[string][]string{"A5|S0": {"AX-1050"}},
	}
	run := setupTest(t, mock)

	out, err := run("board", "add", "AllTix Webticket", "AX-1050", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 0 {
		t.Errorf("expected no add on a query board, got %v", mock.addedToSprint)
	}
	var got []sprintChange
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(got) != 1 || got[0].Changed || got[0].Note != "on board by its query, nothing to add" || got[0].Sprint != "" {
		t.Errorf("got %+v", got)
	}
}

func TestRunBoardAddQueryBoardNotMatching(t *testing.T) {
	mock := &mockAPI{board: queryBoard()}
	run := setupTest(t, mock)

	_, err := run("board", "add", "AllTix Webticket", "AX-1061")
	if err == nil {
		t.Fatal("expected error for an issue the board's query does not show")
	}
	for _, want := range []string{"AX-1061", `"Subsystem: Webticket"`, "change the issue"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}
	if len(mock.addedToSprint) != 0 {
		t.Errorf("expected no add on a query board, got %v", mock.addedToSprint)
	}
}

func TestRunBoardRemoveQueryBoard(t *testing.T) {
	mock := &mockAPI{
		board:        queryBoard(),
		sprintIssues: map[string][]string{"A5|S0": {"AX-1050"}},
	}
	run := setupTest(t, mock)

	_, err := run("board", "remove", "AllTix Webticket", "AX-1050")
	if err == nil || !strings.Contains(err.Error(), "no longer matches") {
		t.Fatalf("err = %v, want a removal refusal", err)
	}
	if len(mock.removedSprint) != 0 {
		t.Errorf("expected no remove on a query board, got %v", mock.removedSprint)
	}
}

func TestRunBoardAddSprintlessRejectsSprint(t *testing.T) {
	mock := &mockAPI{board: queryBoard()}
	run := setupTest(t, mock)

	_, err := run("board", "add", "AllTix Webticket", "AX-1050", "--sprint", "First sprint")
	if err == nil || !strings.Contains(err.Error(), "has no sprints") {
		t.Fatalf("err = %v, want no-sprints rejection", err)
	}
}

func TestRunBoardAddManualSprintless(t *testing.T) {
	board := queryBoard()
	board.Name = "AllTix Planning"
	board.SprintsSettings = &youtrack.SprintsSettings{DisableSprints: true, IsExplicit: true}
	mock := &mockAPI{board: board}
	run := setupTest(t, mock)

	out, err := run("board", "add", "AllTix Planning", "AX-1", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A5|S0|AX-1" {
		t.Errorf("addedToSprint = %v, want the implicit sprint", mock.addedToSprint)
	}
	if strings.Contains(out, "First sprint") {
		t.Errorf("implicit sprint leaked into output: %s", out)
	}
}

func TestRunIssueCreateLikeSkipsQueryBoard(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Sub"},
		board: testBoard(),
		issueBoards: []youtrack.BoardMembership{
			{Board: "AllTix Webticket", Kind: youtrack.BoardQuery},
			{Board: "AllTix", Kind: youtrack.BoardSprints, Sprint: "2025-06"},
		},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "AX", "-s", "Sub", "--like", "AX-332")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S2|PROJ-999" {
		t.Errorf("addedToSprint = %v, want only the sprint board", mock.addedToSprint)
	}
}
