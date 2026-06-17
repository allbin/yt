package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	if mock.stateSet != "Open" {
		t.Errorf("stateSet = %q, want %q", mock.stateSet, "Open")
	}
	want := "Severity Critical Subsystem API"
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

func TestRunIssueUpdateSummary(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "New title"},
	}
	run := setupTest(t, mock)

	out, err := run("issue", "update", "PROJ-123", "-S", "New title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updatedFields["summary"] != "New title" {
		t.Errorf("summary = %q, want %q", mock.updatedFields["summary"], "New title")
	}
	if mock.command != "" {
		t.Errorf("command should be empty, got %q", mock.command)
	}
	if !strings.Contains(out, "PROJ-123") {
		t.Errorf("output missing issue ID: %s", out)
	}
}

func TestRunIssueUpdateDescription(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "-d", "Updated body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updatedFields["description"] != "Updated body" {
		t.Errorf("description = %q, want %q", mock.updatedFields["description"], "Updated body")
	}
	if mock.command != "" {
		t.Errorf("command should be empty, got %q", mock.command)
	}
}

func TestRunIssueUpdateSummaryAndDescription(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "New title"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "-S", "New title", "-d", "New body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updatedFields["summary"] != "New title" {
		t.Errorf("summary = %q, want %q", mock.updatedFields["summary"], "New title")
	}
	if mock.updatedFields["description"] != "New body" {
		t.Errorf("description = %q, want %q", mock.updatedFields["description"], "New body")
	}
}

func TestRunIssueUpdateStateOnly(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "Test"},
	}
	run := setupTest(t, mock)

	out, err := run("issue", "update", "PROJ-123", "-s", "In Progress")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.stateSet != "In Progress" {
		t.Errorf("stateSet = %q, want %q", mock.stateSet, "In Progress")
	}
	if mock.command != "" {
		t.Errorf("command should be empty, got %q", mock.command)
	}
	if !strings.Contains(out, "PROJ-123") {
		t.Errorf("output missing issue ID: %s", out)
	}
}

func TestRunIssueUpdateCombinedRESTAndCommand(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-123", Summary: "New title"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "PROJ-123", "-S", "New title", "-s", "In Progress", "-a", "me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.updatedFields["summary"] != "New title" {
		t.Errorf("summary = %q, want %q", mock.updatedFields["summary"], "New title")
	}
	if mock.stateSet != "In Progress" {
		t.Errorf("stateSet = %q, want %q", mock.stateSet, "In Progress")
	}
	if mock.command != "Assignee me" {
		t.Errorf("command = %q, want %q", mock.command, "Assignee me")
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

func TestRunIssueShowsBoardMembership(t *testing.T) {
	run := setupTest(t, &mockAPI{
		issue:       &youtrack.Issue{IDReadable: "AX-812", Summary: "On a board"},
		issueBoards: []youtrack.BoardMembership{{Board: "AllTix", Sprint: "2025-06"}},
	})

	out, err := run("issue", "AX-812")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "AllTix") || !strings.Contains(out, "2025-06") {
		t.Errorf("output missing board membership: %s", out)
	}
}

func TestRunIssueBoardLookupErrorIgnored(t *testing.T) {
	run := setupTest(t, &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "AX-812", Summary: "x"},
		sprintErr: fmt.Errorf("boom"),
	})

	out, err := run("issue", "AX-812")
	if err != nil {
		t.Fatalf("board lookup failure should not fail the command: %v", err)
	}
	if !strings.Contains(out, "AX-812") {
		t.Errorf("output missing issue: %s", out)
	}
}

func TestRunIssueCreateWithBoard(t *testing.T) {
	mock := &mockAPI{
		issue:       &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Sub"},
		board:       testBoard(),
		issueBoards: []youtrack.BoardMembership{{Board: "AllTix", Sprint: "2025-06"}},
	}
	run := setupTest(t, mock)

	out, err := run("issue", "create", "-p", "AX", "-s", "Sub", "--board", "AllTix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S2|PROJ-999" {
		t.Errorf("addedToSprint = %v, want [A1|S2|PROJ-999]", mock.addedToSprint)
	}
	if !strings.Contains(out, "AllTix") {
		t.Errorf("output missing board: %s", out)
	}
}

func TestRunIssueCreateLikeMirrorsBoards(t *testing.T) {
	mock := &mockAPI{
		issue:       &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Sub"},
		board:       testBoard(),
		issueBoards: []youtrack.BoardMembership{{Board: "AllTix", Sprint: "2025-06"}},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "AX", "-s", "Sub", "--like", "AX-332")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S2|PROJ-999" {
		t.Errorf("addedToSprint = %v, want [A1|S2|PROJ-999]", mock.addedToSprint)
	}
}

func TestRunIssueCreateParentLinksAndMirrors(t *testing.T) {
	mock := &mockAPI{
		issue:       &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Sub"},
		board:       testBoard(),
		linkTypes:   mockLinkTypes,
		issueBoards: []youtrack.BoardMembership{{Board: "AllTix", Sprint: "2025-06"}},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "AX", "-s", "Sub", "--parent", "AX-332")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// subtask-of link from the new issue to the parent.
	if len(mock.createdLinks) != 1 || mock.createdLinks[0] != "PROJ-999|subtask of|AX-332" {
		t.Errorf("createdLinks = %v, want [PROJ-999|subtask of|AX-332]", mock.createdLinks)
	}
	// placed on the parent's board+sprint.
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S2|PROJ-999" {
		t.Errorf("addedToSprint = %v, want [A1|S2|PROJ-999]", mock.addedToSprint)
	}
}

func TestRunIssueCreateParentNotOnBoardStillLinks(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Sub"},
		board:     testBoard(),
		linkTypes: mockLinkTypes,
		// issueBoards empty: parent is on no board.
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "AX", "-s", "Sub", "--parent", "AX-332")
	if err != nil {
		t.Fatalf("--parent should not fail when parent is on no board: %v", err)
	}
	if len(mock.createdLinks) != 1 {
		t.Errorf("expected subtask link, got %v", mock.createdLinks)
	}
	if len(mock.addedToSprint) != 0 {
		t.Errorf("expected no board placement, got %v", mock.addedToSprint)
	}
}

func TestRunIssueCreateParentBoardOverride(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Sub"},
		board:     testBoard(),
		linkTypes: mockLinkTypes,
		// issueBoards would say the parent is on a different sprint, but --board wins.
		issueBoards: []youtrack.BoardMembership{{Board: "AllTix", Sprint: "2025-05"}},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "AX", "-s", "Sub", "--parent", "AX-332", "--board", "AllTix", "--sprint", "2025-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.createdLinks) != 1 {
		t.Errorf("expected subtask link, got %v", mock.createdLinks)
	}
	// Explicit --board/--sprint used; parent mirror skipped (no duplicate add).
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S1|PROJ-999" {
		t.Errorf("addedToSprint = %v, want [A1|S1|PROJ-999]", mock.addedToSprint)
	}
}

func TestRunIssueCreateParentLikeMutuallyExclusive(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Sub"},
		linkTypes: mockLinkTypes,
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "AX", "-s", "Sub", "--parent", "AX-332", "--like", "AX-1")
	if err == nil {
		t.Fatal("expected error for mutually exclusive --parent and --like")
	}
	if len(mock.createdLinks) != 0 {
		t.Errorf("should not create link when flags conflict, got %v", mock.createdLinks)
	}
}

func TestRunIssueCreateLikeNotOnBoard(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Sub"},
		board: testBoard(),
		// issueBoards empty: parent is on no board.
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "AX", "-s", "Sub", "--like", "AX-332")
	if err == nil {
		t.Fatal("expected error when --like target is on no board")
	}
	if !strings.Contains(err.Error(), "not on any board") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunIssueUpdateWithBoard(t *testing.T) {
	mock := &mockAPI{
		issue:       &youtrack.Issue{IDReadable: "AX-812", Summary: "x"},
		board:       testBoard(),
		issueBoards: []youtrack.BoardMembership{{Board: "AllTix", Sprint: "2025-06"}},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "AX-812", "--board", "AllTix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.addedToSprint) != 1 || mock.addedToSprint[0] != "A1|S2|AX-812" {
		t.Errorf("addedToSprint = %v, want [A1|S2|AX-812]", mock.addedToSprint)
	}
}

func TestRunIssueCommentFromStdin(t *testing.T) {
	mock := &mockAPI{}
	run := setupTest(t, mock)

	rootCmd.SetIn(strings.NewReader("piped comment body"))
	t.Cleanup(func() { rootCmd.SetIn(nil) })

	out, err := run("issue", "comment", "PROJ-123", "-m", "-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "mock-comment-1") {
		t.Errorf("output missing comment ID: %s", out)
	}
	if mock.addedComment != "piped comment body" {
		t.Errorf("addedComment = %q, want piped comment body", mock.addedComment)
	}
}

func TestRunIssueCreateDescriptionFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "desc.md")
	if err := os.WriteFile(file, []byte("multi\nline\nbody"), 0o600); err != nil {
		t.Fatal(err)
	}
	mock := &mockAPI{issue: &youtrack.Issue{IDReadable: "PROJ-999", Summary: "x"}}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "PROJ", "-s", "x", "-d", "@"+file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.createdDescription != "multi\nline\nbody" {
		t.Errorf("createdDescription = %q, want multi-line body", mock.createdDescription)
	}
}
