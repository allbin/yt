package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

var mockLinkTypes = []youtrack.LinkType{
	{ID: "105-0", Name: "Relates", SourceToTarget: "relates to", TargetToSource: "", Directed: false},
	{ID: "105-1", Name: "Depend", SourceToTarget: "is required for", TargetToSource: "depends on", Directed: true},
	{ID: "105-2", Name: "Duplicate", SourceToTarget: "is duplicated by", TargetToSource: "duplicates", Directed: true},
	{ID: "105-3", Name: "Subtask", SourceToTarget: "parent for", TargetToSource: "subtask of", Directed: true},
}

func subtaskLink() youtrack.IssueLink {
	return youtrack.IssueLink{
		ID:        "105-3t",
		Direction: youtrack.DirInward,
		LinkType:  mockLinkTypes[3],
		Issues:    []youtrack.LinkedIssue{{ID: "2-6261", IDReadable: "AX-332", Summary: "Story: students"}},
	}
}

func TestRunLinkCreate(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "AX-804", Summary: "Child"},
		linkTypes: mockLinkTypes,
	}
	run := setupTest(t, mock)

	out, err := run("link", "AX-804", "subtask-of", "AX-332")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.createdLinks) != 1 || mock.createdLinks[0] != "AX-804|subtask of|AX-332" {
		t.Errorf("createdLinks = %v, want [AX-804|subtask of|AX-332]", mock.createdLinks)
	}
	if !strings.Contains(out, "AX-804") || !strings.Contains(out, "AX-332") {
		t.Errorf("output missing ids: %s", out)
	}
}

func TestRunLinkMultipleTargets(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "AX-1", Summary: "Hub"},
		linkTypes: mockLinkTypes,
	}
	run := setupTest(t, mock)

	_, err := run("link", "AX-1", "relates", "AX-2", "AX-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.createdLinks) != 2 {
		t.Fatalf("createdLinks = %v, want 2", mock.createdLinks)
	}
	if mock.createdLinks[0] != "AX-1|relates to|AX-2" || mock.createdLinks[1] != "AX-1|relates to|AX-3" {
		t.Errorf("createdLinks = %v", mock.createdLinks)
	}
}

func TestRunLinkAlreadyLinkedNoOp(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{
			IDReadable: "AX-804",
			Summary:    "Child",
			Links:      []youtrack.IssueLink{subtaskLink()},
		},
		linkTypes: mockLinkTypes,
	}
	run := setupTest(t, mock)

	out, err := run("link", "AX-804", "subtask-of", "AX-332")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.createdLinks) != 0 {
		t.Errorf("expected no create call, got %v", mock.createdLinks)
	}
	if !strings.Contains(out, "already linked") {
		t.Errorf("output missing 'already linked': %s", out)
	}
}

func TestRunLinkUnknownRelation(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "AX-1"},
		linkTypes: mockLinkTypes,
	}
	run := setupTest(t, mock)

	_, err := run("link", "AX-1", "frobnicate", "AX-2")
	if err == nil {
		t.Fatal("expected error for unknown relation")
	}
	if !strings.Contains(err.Error(), "unknown relation") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "subtask of") {
		t.Errorf("error should list valid relations: %v", err)
	}
}

func TestRunLinkJSON(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{
			IDReadable: "AX-804",
			Links:      []youtrack.IssueLink{subtaskLink()},
		},
		linkTypes: mockLinkTypes,
	}
	run := setupTest(t, mock)

	out, err := run("link", "AX-804", "subtask-of", "AX-332", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []youtrack.IssueLink
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(got) != 1 || got[0].Issues[0].IDReadable != "AX-332" {
		t.Errorf("got %+v", got)
	}
}

func TestRunUnlink(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{
			IDReadable: "AX-804",
			Links:      []youtrack.IssueLink{subtaskLink()},
		},
		linkTypes: mockLinkTypes,
	}
	run := setupTest(t, mock)

	out, err := run("unlink", "AX-804", "subtask-of", "AX-332")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.removedLinks) != 1 || mock.removedLinks[0] != "AX-804|105-3t|2-6261" {
		t.Errorf("removedLinks = %v, want [AX-804|105-3t|2-6261]", mock.removedLinks)
	}
	if !strings.Contains(out, "removed") {
		t.Errorf("output missing 'removed': %s", out)
	}
}

func TestRunUnlinkMissing(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "AX-804"},
		linkTypes: mockLinkTypes,
	}
	run := setupTest(t, mock)

	_, err := run("unlink", "AX-804", "subtask-of", "AX-999")
	if err == nil {
		t.Fatal("expected error for missing link")
	}
	if !strings.Contains(err.Error(), "no \"subtask of\" link") {
		t.Errorf("unexpected error: %v", err)
	}
	if len(mock.removedLinks) != 0 {
		t.Errorf("should not call RemoveLink, got %v", mock.removedLinks)
	}
}

func TestRunLinks(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{
			IDReadable: "AX-804",
			Links:      []youtrack.IssueLink{subtaskLink()},
		},
	}
	run := setupTest(t, mock)

	out, err := run("links", "AX-804")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "subtask of") || !strings.Contains(out, "AX-332") {
		t.Errorf("output missing link: %s", out)
	}
}

func TestRunLinksJSON(t *testing.T) {
	mock := &mockAPI{
		issue: &youtrack.Issue{
			IDReadable: "AX-804",
			Links:      []youtrack.IssueLink{subtaskLink()},
		},
	}
	run := setupTest(t, mock)

	out, err := run("links", "AX-804", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []youtrack.IssueLink
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(got) != 1 || got[0].LinkType.Name != "Subtask" {
		t.Errorf("got %+v", got)
	}
}

func TestRunLinkTypes(t *testing.T) {
	mock := &mockAPI{linkTypes: mockLinkTypes}
	run := setupTest(t, mock)

	out, err := run("link", "types")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Subtask") || !strings.Contains(out, "subtask of") {
		t.Errorf("output missing link type: %s", out)
	}
}

func TestRunLinkTypesJSON(t *testing.T) {
	mock := &mockAPI{linkTypes: mockLinkTypes}
	run := setupTest(t, mock)

	out, err := run("link", "types", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []youtrack.LinkType
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(got) != 4 {
		t.Errorf("got %d types, want 4", len(got))
	}
}
