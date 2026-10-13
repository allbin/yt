package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestRunBoardList(t *testing.T) {
	run := setupTest(t, &mockAPI{
		boards: []youtrack.Agile{
			{
				Name:     "Sprint Board",
				Projects: []youtrack.Project{{ShortName: "PROJ"}},
				CurrentSprint: &youtrack.Sprint{
					Name: "2025-01",
				},
			},
		},
	})

	out, err := run("board", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Sprint Board") {
		t.Errorf("output missing board name: %s", out)
	}
	if !strings.Contains(out, "PROJ") {
		t.Errorf("output missing project: %s", out)
	}
}

func TestRunBoardListEmpty(t *testing.T) {
	run := setupTest(t, &mockAPI{})

	out, err := run("board", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No boards found") {
		t.Errorf("expected empty message: %s", out)
	}
}

func TestRunBoardJSONSprintBoard(t *testing.T) {
	run := setupTest(t, &mockAPI{
		board:  testBoard(),
		issues: []youtrack.Issue{{IDReadable: "AX-1", Summary: "One"}},
	})

	out, err := run("board", "AllTix", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got boardView
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got.Name != "AllTix" || got.Kind != "sprints" || got.Sprint == nil || got.Sprint.Name != "2025-06" {
		t.Errorf("got %+v", got)
	}
	if len(got.Issues) != 1 || got.Issues[0].IDReadable != "AX-1" {
		t.Errorf("issues = %+v", got.Issues)
	}
	if !strings.Contains(got.Membership, "sprints") {
		t.Errorf("membership = %q", got.Membership)
	}
}

func TestRunBoardJSONQueryBoard(t *testing.T) {
	run := setupTest(t, &mockAPI{board: queryBoard()})

	out, err := run("board", "AllTix Webticket", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got boardView
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if got.Kind != "query" || got.Query != "Subsystem: Webticket" || got.Sprint != nil {
		t.Errorf("got %+v", got)
	}
	if got.Issues == nil {
		t.Error("issues should be an empty list, not null")
	}
	if !strings.Contains(out, `"membership": "issues matching the query`) {
		t.Errorf("membership missing: %s", out)
	}
}

func TestRunBoardSprintlessRejectsSprint(t *testing.T) {
	run := setupTest(t, &mockAPI{board: queryBoard()})

	_, err := run("board", "AllTix Webticket", "--sprint", "x")
	if err == nil || !strings.Contains(err.Error(), "has no sprints") {
		t.Fatalf("err = %v, want no-sprints rejection", err)
	}
}

func TestRunBoardTextSprintless(t *testing.T) {
	run := setupTest(t, &mockAPI{board: queryBoard()})

	out, err := run("board", "AllTix Webticket")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Subsystem: Webticket") || strings.Contains(out, "First sprint") {
		t.Errorf("header should name the query, not the implicit sprint: %s", out)
	}
}

func TestRunBoardListShowsKind(t *testing.T) {
	run := setupTest(t, &mockAPI{boards: []youtrack.Agile{*testBoard(), *queryBoard()}})

	out, err := run("board", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sprints") || !strings.Contains(out, "query") {
		t.Errorf("kinds missing: %s", out)
	}
}
