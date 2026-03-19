package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestBoardList(t *testing.T) {
	boards := []youtrack.Agile{
		{
			ID:   "1",
			Name: "TestBoard",
			Projects: []youtrack.Project{
				{ShortName: "TB"},
			},
			CurrentSprint: &youtrack.Sprint{ID: "s1", Name: "2026-01"},
		},
		{
			ID:   "2",
			Name: "NoSprint",
			Projects: []youtrack.Project{
				{ShortName: "NS"},
			},
		},
	}

	var buf bytes.Buffer
	if err := BoardList(&buf, boards); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, "TestBoard") {
		t.Error("missing board name")
	}
	if !strings.Contains(out, "2026-01") {
		t.Error("missing sprint name")
	}
	if !strings.Contains(out, "NoSprint") {
		t.Error("missing board without sprint")
	}
}

func TestBoardListEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := BoardList(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No boards found") {
		t.Error("expected empty message")
	}
}

func TestSprintIssues(t *testing.T) {
	issues := []youtrack.Issue{
		{
			IDReadable: "T-1",
			Summary:    "First",
			CustomFields: []youtrack.CustomField{
				{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
			},
		},
	}

	var buf bytes.Buffer
	if err := SprintIssues(&buf, "TestBoard", "2026-01", issues); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, "TestBoard") {
		t.Error("missing board name")
	}
	if !strings.Contains(out, "2026-01") {
		t.Error("missing sprint name")
	}
	if !strings.Contains(out, "T-1") {
		t.Error("missing issue ID")
	}
}

func TestSprintIssuesEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := SprintIssues(&buf, "TestBoard", "2026-01", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No issues") {
		t.Error("expected empty message")
	}
}

func TestFormatProjects(t *testing.T) {
	tests := []struct {
		name     string
		projects []youtrack.Project
		want     string
	}{
		{"empty", nil, ""},
		{"one", []youtrack.Project{{ShortName: "A"}}, "A"},
		{"three", []youtrack.Project{{ShortName: "A"}, {ShortName: "B"}, {ShortName: "C"}}, "A, B, C"},
		{"four_truncated", []youtrack.Project{{ShortName: "A"}, {ShortName: "B"}, {ShortName: "C"}, {ShortName: "D"}}, "A, B, C, +1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatProjects(tt.projects); got != tt.want {
				t.Errorf("formatProjects() = %q, want %q", got, tt.want)
			}
		})
	}
}
