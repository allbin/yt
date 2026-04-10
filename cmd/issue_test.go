package cmd

import (
	"encoding/json"
	"fmt"
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

func TestRunIssueUpdateWithSubsystem(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "--subsystem", "API")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.command != "Subsystem API" {
		t.Errorf("command = %q, want %q", mock.command, "Subsystem API")
	}
}

func TestRunIssueUpdateWithField(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "--field", "Severity=Critical")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.command != "Severity Critical" {
		t.Errorf("command = %q, want %q", mock.command, "Severity Critical")
	}
}

func TestRunIssueUpdateWithFieldAndSubsystem(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "--subsystem", "API", "--field", "Severity=Critical", "-s", "Open")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "State Open Severity Critical Subsystem API"
	if mock.command != want {
		t.Errorf("command = %q, want %q", mock.command, want)
	}
}

func TestRunIssueUpdateEmptySubsystem(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "--subsystem", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.command != "Subsystem " {
		t.Errorf("command = %q, want %q", mock.command, "Subsystem ")
	}
}

func TestRunIssueUpdateEmptyField(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "--field", "Subsystem=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.command != "Subsystem " {
		t.Errorf("command = %q, want %q", mock.command, "Subsystem ")
	}
}

func TestRunIssueCreateEmptySubsystem(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "PROJ", "-s", "Test", "--subsystem", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.command != "Subsystem " {
		t.Errorf("command = %q, want %q", mock.command, "Subsystem ")
	}
}

func TestRunIssueUpdateNoFlags(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123")
	if err == nil {
		t.Fatal("expected error for no flags")
	}
	if !strings.Contains(err.Error(), "no fields to update") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunIssueUpdateInvalidField(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "--field", "bad-format")
	if err == nil {
		t.Fatal("expected error for invalid field format")
	}
	if !strings.Contains(err.Error(), "invalid --field format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunIssueCreateWithSubsystem(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Test"},
	}
	run := setupTest(t, mock)

	out, err := run("issue", "create", "-p", "PROJ", "-s", "Test", "--subsystem", "API")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.command != "Subsystem API" {
		t.Errorf("command = %q, want %q", mock.command, "Subsystem API")
	}
	if !strings.Contains(out, "PROJ-") {
		t.Errorf("output missing issue ID: %s", out)
	}
}

func TestRunIssueCreateWithField(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "PROJ", "-s", "Test", "--field", "Severity=Critical")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.command != "Severity Critical" {
		t.Errorf("command = %q, want %q", mock.command, "Severity Critical")
	}
}

func TestRunIssueCreateFieldFailure(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Test"},
		updateErr: fmt.Errorf("unknown field"),
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "PROJ", "-s", "Test", "--subsystem", "BadValue")
	if err == nil {
		t.Fatal("expected error when field-setting fails")
	}
	if !strings.Contains(err.Error(), "set fields on PROJ-999") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunIssueCreateNoFieldsSkipsUpdate(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "PROJ", "-s", "Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.command != "" {
		t.Errorf("expected no update command, got %q", mock.command)
	}
}
