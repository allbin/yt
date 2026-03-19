package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allbin/yt/internal/youtrack"
)

var (
	testDesc      = "This is a test description."
	testViewIssue = &youtrack.Issue{
		IDReadable:  "TEST-1",
		Summary:     "Test issue summary",
		Description: &testDesc,
		CustomFields: []youtrack.CustomField{
			{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
			{Name: "Priority", Value: json.RawMessage(`{"name":"Major"}`)},
			{Name: "Assignee", Value: json.RawMessage(`{"fullName":"John Doe"}`)},
			{Name: "Type", Value: json.RawMessage(`{"name":"Bug"}`)},
		},
	}
	testViewComments = []youtrack.Comment{
		{
			ID:      "c1",
			Text:    "First comment",
			Author:  &youtrack.User{Login: "alice", FullName: "Alice Smith"},
			Created: 1700000000000,
		},
	}
)

func newLoadedViewer(api *mockAPI) IssueViewer {
	v := NewIssueViewer(api, "TEST-1")
	m, _ := v.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	v = m.(IssueViewer)
	m, _ = v.Update(issueLoadedMsg{
		issue:    api.issue,
		comments: api.comments,
		states:   api.states,
	})
	return m.(IssueViewer)
}

func TestIssueViewerInit(t *testing.T) {
	v := NewIssueViewer(&mockAPI{}, "TEST-1")
	if !v.loading {
		t.Error("expected loading=true on init")
	}
	if cmd := v.Init(); cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestIssueViewerLoaded(t *testing.T) {
	api := &mockAPI{issue: testViewIssue, comments: testViewComments, states: testStates}
	v := newLoadedViewer(api)

	if v.loading {
		t.Error("expected loading=false after load")
	}
	if v.issue == nil {
		t.Error("expected issue to be set")
	}
	if len(v.lines) == 0 {
		t.Error("expected lines to be built")
	}
}

func TestIssueViewerLoadError(t *testing.T) {
	v := NewIssueViewer(&mockAPI{}, "TEST-1")
	m, _ := v.Update(issueLoadedMsg{err: fmt.Errorf("network error")})
	v = m.(IssueViewer)

	if v.err == nil {
		t.Error("expected error to be set")
	}
	if !strings.Contains(v.View(), "network error") {
		t.Error("expected error in view")
	}
}

func TestIssueViewerScrollDown(t *testing.T) {
	longDesc := strings.Repeat("This is a paragraph.\n\n", 30)
	issue := &youtrack.Issue{IDReadable: "TEST-1", Summary: "Test", Description: &longDesc}
	v := newLoadedViewer(&mockAPI{issue: issue, states: testStates})

	if v.scrollOffset != 0 {
		t.Fatalf("initial scroll = %d, want 0", v.scrollOffset)
	}

	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	v = m.(IssueViewer)
	if v.scrollOffset != 1 {
		t.Errorf("after j: scroll = %d, want 1", v.scrollOffset)
	}
}

func TestIssueViewerScrollUp(t *testing.T) {
	longDesc := strings.Repeat("This is a paragraph.\n\n", 30)
	issue := &youtrack.Issue{IDReadable: "TEST-1", Summary: "Test", Description: &longDesc}
	v := newLoadedViewer(&mockAPI{issue: issue, states: testStates})

	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	v = m.(IssueViewer)
	if v.scrollOffset != 1 {
		t.Errorf("after jjk: scroll = %d, want 1", v.scrollOffset)
	}
}

func TestIssueViewerScrollBounds(t *testing.T) {
	v := newLoadedViewer(&mockAPI{issue: testViewIssue, states: testStates})

	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	v = m.(IssueViewer)
	if v.scrollOffset != 0 {
		t.Errorf("at top k: scroll = %d, want 0", v.scrollOffset)
	}
}

func TestIssueViewerHalfPageScroll(t *testing.T) {
	longDesc := strings.Repeat("This is a paragraph.\n\n", 60)
	issue := &youtrack.Issue{IDReadable: "TEST-1", Summary: "Test", Description: &longDesc}
	v := newLoadedViewer(&mockAPI{issue: issue, states: testStates})

	halfPage := v.viewportHeight() / 2

	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	v = m.(IssueViewer)
	if v.scrollOffset != halfPage {
		t.Errorf("after ctrl+d: scroll = %d, want %d", v.scrollOffset, halfPage)
	}

	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	v = m.(IssueViewer)
	if v.scrollOffset != 0 {
		t.Errorf("after ctrl+u: scroll = %d, want 0", v.scrollOffset)
	}
}

func TestIssueViewerStatePicker(t *testing.T) {
	v := newLoadedViewer(&mockAPI{issue: testViewIssue, states: testStates})

	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	v = m.(IssueViewer)
	if v.mode != modeStatePicker {
		t.Error("expected modeStatePicker after 's'")
	}

	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	v = m.(IssueViewer)
	if v.mode != modeNormal {
		t.Error("expected modeNormal after cancel")
	}
}

func TestIssueViewerStatePickerNoStates(t *testing.T) {
	v := newLoadedViewer(&mockAPI{issue: testViewIssue})

	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if m.(IssueViewer).mode != modeNormal {
		t.Error("expected modeNormal when no states available")
	}
}

func TestIssueViewerQuit(t *testing.T) {
	v := newLoadedViewer(&mockAPI{issue: testViewIssue, states: testStates})

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command on 'q'")
	}
}

func TestIssueViewerQuitWhileLoading(t *testing.T) {
	v := NewIssueViewer(&mockAPI{}, "TEST-1")

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command on 'q' while loading")
	}
}

func TestIssueViewerRefresh(t *testing.T) {
	v := newLoadedViewer(&mockAPI{issue: testViewIssue, states: testStates})

	m, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	v = m.(IssueViewer)
	if !v.loading {
		t.Error("expected loading=true on refresh")
	}
	if cmd == nil {
		t.Error("expected command on refresh")
	}
}

func TestIssueViewerViewContainsData(t *testing.T) {
	api := &mockAPI{issue: testViewIssue, comments: testViewComments, states: testStates}
	v := newLoadedViewer(api)

	view := v.View()
	for _, want := range []string{"TEST-1", "Test issue summary", "test description", "Comments (1)", "Alice Smith", "First comment"} {
		if !strings.Contains(view, want) {
			t.Errorf("view missing %q", want)
		}
	}
}

func TestIssueViewerNoDescription(t *testing.T) {
	v := newLoadedViewer(&mockAPI{issue: &youtrack.Issue{IDReadable: "TEST-2", Summary: "No desc"}, states: testStates})

	if !strings.Contains(v.View(), "No description") {
		t.Error("expected 'No description' placeholder")
	}
}

func TestIssueViewerWindowResize(t *testing.T) {
	v := newLoadedViewer(&mockAPI{issue: testViewIssue, states: testStates})

	m, _ := v.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	v = m.(IssueViewer)

	if v.width != 40 || v.height != 12 {
		t.Errorf("dimensions = %dx%d, want 40x12", v.width, v.height)
	}
	if len(v.lines) == 0 {
		t.Error("expected lines after resize")
	}
}

func TestCommentAuthor(t *testing.T) {
	tests := []struct {
		name    string
		comment youtrack.Comment
		want    string
	}{
		{"full name", youtrack.Comment{Author: &youtrack.User{FullName: "Alice", Login: "alice"}}, "Alice"},
		{"login only", youtrack.Comment{Author: &youtrack.User{Login: "bob"}}, "bob"},
		{"nil author", youtrack.Comment{}, "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commentAuthor(tt.comment); got != tt.want {
				t.Errorf("commentAuthor() = %q, want %q", got, tt.want)
			}
		})
	}
}
