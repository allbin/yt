package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func testIssue() *youtrack.Issue {
	desc := "Fix the bug"
	return &youtrack.Issue{
		IDReadable:  "PROJ-1",
		Summary:     "Test issue",
		Description: &desc,
		Tags:        []youtrack.Tag{{Name: "urgent"}},
		CustomFields: []youtrack.CustomField{
			{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
			{Name: "Assignee", Value: json.RawMessage(`{"fullName":"John Doe"}`)},
			{Name: "Priority", Value: json.RawMessage(`{"name":"Major"}`)},
			{Name: "Type", Value: json.RawMessage(`{"name":"Bug"}`)},
			{Name: "Subsystem", Value: json.RawMessage(`[{"name":"API"}]`)},
		},
	}
}

func TestIssueFormat(t *testing.T) {
	var buf bytes.Buffer
	if err := Issue(&buf, testIssue()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	checks := []string{
		"# PROJ-1: Test issue",
		"**State:** Open",
		"**Assignee:** John Doe",
		"**Priority:** Major",
		"**Type:** Bug",
		"**Subsystem:** API",
		"**Tags:** urgent",
		"## Description",
		"Fix the bug",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestIssueFormatNoDescription(t *testing.T) {
	issue := &youtrack.Issue{
		IDReadable: "PROJ-2",
		Summary:    "No desc",
	}
	var buf bytes.Buffer
	if err := Issue(&buf, issue); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(buf.String(), "## Description") {
		t.Error("should not contain Description section when description is nil")
	}
}

func TestIssueList(t *testing.T) {
	issues := []youtrack.Issue{
		{
			IDReadable: "PROJ-1",
			Summary:    "First",
			CustomFields: []youtrack.CustomField{
				{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
				{Name: "Priority", Value: json.RawMessage(`{"name":"Major"}`)},
				{Name: "Assignee", Value: json.RawMessage(`{"fullName":"Alice"}`)},
			},
		},
		{
			IDReadable: "PROJ-2",
			Summary:    "Second",
			CustomFields: []youtrack.CustomField{
				{Name: "State", Value: json.RawMessage(`{"name":"Done"}`)},
			},
		},
	}

	var buf bytes.Buffer
	if err := IssueList(&buf, issues); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	if !strings.Contains(out, "PROJ-1") || !strings.Contains(out, "PROJ-2") {
		t.Error("missing issue IDs in output")
	}
	if !strings.Contains(out, "ID") || !strings.Contains(out, "STATE") {
		t.Error("missing table header")
	}
}

func TestIssueListEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := IssueList(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No issues found") {
		t.Error("expected 'No issues found' for empty list")
	}
}

func TestJSON(t *testing.T) {
	issue := testIssue()
	var buf bytes.Buffer
	if err := JSON(&buf, issue); err != nil {
		t.Fatal(err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("JSON output not valid: %v", err)
	}
	if parsed["idReadable"] != "PROJ-1" {
		t.Errorf("idReadable = %v, want PROJ-1", parsed["idReadable"])
	}
}
