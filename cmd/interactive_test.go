package cmd

import (
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

// setupTest writes to a buffer, never a terminal, so these exercise the same
// path an agent or a pipe takes.

func TestIssueViewFallsBackToIssueOutput(t *testing.T) {
	run := setupTest(t, &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-1", Summary: "Fallback summary"},
	})

	out, err := run("issue", "view", "PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PROJ-1") || !strings.Contains(out, "Fallback summary") {
		t.Errorf("want the issue printed instead of a TUI, got: %s", out)
	}
}

func TestIssueViewJSONFallback(t *testing.T) {
	run := setupTest(t, &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-1", Summary: "Fallback summary"},
	})

	out, err := run("issue", "view", "PROJ-1", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"idReadable"`) || !strings.Contains(out, "PROJ-1") {
		t.Errorf("want JSON, got: %s", out)
	}
}

func TestBoardViewFallsBackToSprintIssues(t *testing.T) {
	run := setupTest(t, &mockAPI{
		board: &youtrack.Agile{
			Name:          "Sprint Board",
			CurrentSprint: &youtrack.Sprint{Name: "2025-01"},
		},
		issues: []youtrack.Issue{{IDReadable: "PROJ-7", Summary: "On the board"}},
	})

	out, err := run("board", "view", "Sprint Board")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PROJ-7") || !strings.Contains(out, "On the board") {
		t.Errorf("want the sprint's issues printed instead of a TUI, got: %s", out)
	}
}

func TestBoardViewJSONFallback(t *testing.T) {
	run := setupTest(t, &mockAPI{
		board: &youtrack.Agile{
			Name:          "Sprint Board",
			CurrentSprint: &youtrack.Sprint{Name: "2025-01"},
		},
		issues: []youtrack.Issue{{IDReadable: "PROJ-7", Summary: "On the board"}},
	})

	out, err := run("board", "view", "Sprint Board", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"idReadable"`) || !strings.Contains(out, "PROJ-7") {
		t.Errorf("want JSON, got: %s", out)
	}
}

// The viewer's --sprint flag must survive the fallback.
func TestBoardViewFallbackHonoursSprintFlag(t *testing.T) {
	run := setupTest(t, &mockAPI{
		board: &youtrack.Agile{
			Name:          "Sprint Board",
			CurrentSprint: &youtrack.Sprint{Name: "2025-01"},
			Sprints:       []youtrack.Sprint{{Name: "2025-06"}},
		},
		issues: []youtrack.Issue{{IDReadable: "PROJ-7", Summary: "On the board"}},
	})

	out, err := run("board", "view", "Sprint Board", "--sprint", "2025-06")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "2025-06") {
		t.Errorf("want the requested sprint, got: %s", out)
	}
}
