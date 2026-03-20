package cmd

import (
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
