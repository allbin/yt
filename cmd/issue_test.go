package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestRunIssue(t *testing.T) {
	run := setupTest(t, &mockAPI{
		issue: &youtrack.Issue{
			IDReadable: "PROJ-123",
			Summary:    "Fix login bug",
		},
	})

	out, err := run("issue", "PROJ-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PROJ-123") {
		t.Errorf("output missing issue ID: %s", out)
	}
	if !strings.Contains(out, "Fix login bug") {
		t.Errorf("output missing summary: %s", out)
	}
}

func TestRunIssueJSON(t *testing.T) {
	run := setupTest(t, &mockAPI{
		issue: &youtrack.Issue{
			IDReadable: "PROJ-123",
			Summary:    "Fix login bug",
		},
	})

	out, err := run("issue", "PROJ-123", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got youtrack.Issue
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if got.IDReadable != "PROJ-123" {
		t.Errorf("got ID %q, want PROJ-123", got.IDReadable)
	}
}

func TestRunIssueList(t *testing.T) {
	run := setupTest(t, &mockAPI{
		issues: []youtrack.Issue{
			{IDReadable: "PROJ-1", Summary: "First issue"},
			{IDReadable: "PROJ-2", Summary: "Second issue"},
		},
	})

	out, err := run("issue", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PROJ-1") {
		t.Errorf("output missing PROJ-1: %s", out)
	}
	if !strings.Contains(out, "PROJ-2") {
		t.Errorf("output missing PROJ-2: %s", out)
	}
}

func TestRunIssueComments(t *testing.T) {
	run := setupTest(t, &mockAPI{
		comments: []youtrack.Comment{
			{ID: "c-1", Text: "Looks good"},
			{ID: "c-2", Text: "Merged"},
		},
	})

	out, err := run("issue", "comments", "PROJ-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Looks good") {
		t.Errorf("output missing comment text: %s", out)
	}
}

func TestRunIssueComment(t *testing.T) {
	run := setupTest(t, &mockAPI{})

	out, err := run("issue", "comment", "PROJ-123", "-m", "Ship it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "mock-comment-1") {
		t.Errorf("output missing comment ID: %s", out)
	}
	if !strings.Contains(out, "PROJ-123") {
		t.Errorf("output missing issue ID: %s", out)
	}
}

func TestRunIssueCreateWithTags(t *testing.T) {
	run := setupTest(t, &mockAPI{})

	out, err := run("issue", "create", "-p", "PROJ", "-s", "Tagged issue", "-t", "tech-debt", "-t", "scheduler")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PROJ-999") {
		t.Errorf("output missing issue ID: %s", out)
	}
}

func TestRunIssueCreateWithTagsJSON(t *testing.T) {
	run := setupTest(t, &mockAPI{})

	out, err := run("issue", "create", "-p", "PROJ", "-s", "Tagged issue", "-t", "tech-debt", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got youtrack.Issue
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(got.Tags) != 1 || got.Tags[0].Name != "tech-debt" {
		t.Errorf("got tags %v, want [{tech-debt}]", got.Tags)
	}
}

func TestRunIssueUpdateWithTags(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	out, err := run("issue", "update", "PROJ-123", "--tag", "tech-debt", "--tag", "scheduler")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PROJ-123") {
		t.Errorf("output missing issue ID: %s", out)
	}
}

func TestRunIssueUpdateWithRemoveTag(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	out, err := run("issue", "update", "PROJ-123", "--remove-tag", "obsolete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PROJ-123") {
		t.Errorf("output missing issue ID: %s", out)
	}
}
