package tui

import (
	"encoding/json"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allbin/yt/internal/youtrack"
)

var testBoard = &youtrack.Agile{
	ID:            "b1",
	Name:          "TestBoard",
	CurrentSprint: &youtrack.Sprint{ID: "s1", Name: "Sprint 1"},
	Sprints:       []youtrack.Sprint{{ID: "s1", Name: "Sprint 1"}, {ID: "s2", Name: "Sprint 2"}},
	ColumnSettings: &youtrack.AgileColumnSettings{
		Field: &struct{ Name string `json:"name"` }{Name: "State"},
		Columns: []youtrack.AgileColumn{
			{Presentation: "Open", Ordinal: 0, FieldValues: []youtrack.AgileColumnValue{{Name: "Open"}}},
			{Presentation: "In Progress", Ordinal: 1, FieldValues: []youtrack.AgileColumnValue{{Name: "In Progress"}}},
			{Presentation: "Done", Ordinal: 2, FieldValues: []youtrack.AgileColumnValue{{Name: "Done", IsResolved: true}}},
		},
	},
}

func testBoardIssues() []youtrack.Issue {
	return []youtrack.Issue{
		{
			IDReadable: "TEST-1",
			Summary:    "Open issue",
			CustomFields: []youtrack.CustomField{
				{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
			},
		},
		{
			IDReadable: "TEST-2",
			Summary:    "WIP issue",
			CustomFields: []youtrack.CustomField{
				{Name: "State", Value: json.RawMessage(`{"name":"In Progress"}`)},
			},
		},
		{
			IDReadable: "TEST-3",
			Summary:    "Done issue",
			CustomFields: []youtrack.CustomField{
				{Name: "State", Value: json.RawMessage(`{"name":"Done"}`)},
			},
		},
		{
			IDReadable: "TEST-4",
			Summary:    "Another open issue",
			CustomFields: []youtrack.CustomField{
				{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
			},
		},
	}
}

func newLoadedBoard(api *mockAPI) BoardViewer {
	v := NewBoardViewer(api, "TestBoard", "", nil)
	m, _ := v.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	v = m.(BoardViewer)
	m, _ = v.Update(boardLoadedMsg{board: api.board, issues: api.issues})
	return m.(BoardViewer)
}

func TestBoardViewerInit(t *testing.T) {
	v := NewBoardViewer(&mockAPI{}, "TestBoard", "", nil)
	if !v.loading {
		t.Error("expected loading=true on init")
	}
	if cmd := v.Init(); cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestBoardViewerLoaded(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	if v.loading {
		t.Error("expected loading=false after load")
	}
	if len(v.columns) == 0 {
		t.Error("expected columns to be parsed")
	}
	if len(v.issues) == 0 {
		t.Error("expected grid to be built")
	}
}

func TestBoardViewerLoadError(t *testing.T) {
	v := NewBoardViewer(&mockAPI{}, "TestBoard", "", nil)
	m, _ := v.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	v = m.(BoardViewer)
	m, _ = v.Update(boardLoadedMsg{err: fmt.Errorf("network error")})
	v = m.(BoardViewer)

	if v.err == nil {
		t.Error("expected error to be set")
	}
	view := v.View()
	if view == "" || !contains(view, "network error") {
		t.Error("expected error in view")
	}
}

func TestBoardViewerColumnParsing(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	if len(v.columns) != 3 {
		t.Fatalf("columns = %d, want 3", len(v.columns))
	}

	want := []struct {
		presentation string
		stateNames   []string
	}{
		{"Open", []string{"Open"}},
		{"In Progress", []string{"In Progress"}},
		{"Done", []string{"Done"}},
	}
	for i, w := range want {
		if v.columns[i].presentation != w.presentation {
			t.Errorf("col[%d].presentation = %q, want %q", i, v.columns[i].presentation, w.presentation)
		}
		if len(v.columns[i].stateNames) != len(w.stateNames) {
			t.Errorf("col[%d].stateNames len = %d, want %d", i, len(v.columns[i].stateNames), len(w.stateNames))
			continue
		}
		for j, sn := range w.stateNames {
			if v.columns[i].stateNames[j] != sn {
				t.Errorf("col[%d].stateNames[%d] = %q, want %q", i, j, v.columns[i].stateNames[j], sn)
			}
		}
	}
}

func TestBoardViewerGridBuilding(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	// col 0 (Open): TEST-1, TEST-4
	if len(v.issues[0][0]) != 2 {
		t.Errorf("Open column = %d issues, want 2", len(v.issues[0][0]))
	}
	// col 1 (In Progress): TEST-2
	if len(v.issues[1][0]) != 1 {
		t.Errorf("In Progress column = %d issues, want 1", len(v.issues[1][0]))
	}
	// col 2 (Done): TEST-3
	if len(v.issues[2][0]) != 1 {
		t.Errorf("Done column = %d issues, want 1", len(v.issues[2][0]))
	}
}

func TestBoardViewerNavigateColumns(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	if v.cursor.col != 0 {
		t.Fatalf("initial col = %d, want 0", v.cursor.col)
	}

	// Move right
	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	v = m.(BoardViewer)
	if v.cursor.col != 1 {
		t.Errorf("after l: col = %d, want 1", v.cursor.col)
	}

	// Move right again
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	v = m.(BoardViewer)
	if v.cursor.col != 2 {
		t.Errorf("after ll: col = %d, want 2", v.cursor.col)
	}

	// Clamped at right bound
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	v = m.(BoardViewer)
	if v.cursor.col != 2 {
		t.Errorf("at right bound: col = %d, want 2", v.cursor.col)
	}

	// Move left
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	v = m.(BoardViewer)
	if v.cursor.col != 1 {
		t.Errorf("after h: col = %d, want 1", v.cursor.col)
	}

	// Clamp at left bound
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	v = m.(BoardViewer)
	if v.cursor.col != 0 {
		t.Errorf("at left bound: col = %d, want 0", v.cursor.col)
	}
}

func TestBoardViewerNavigateRows(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	// Open column has 2 issues
	if v.cursor.row != 0 {
		t.Fatalf("initial row = %d, want 0", v.cursor.row)
	}

	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	v = m.(BoardViewer)
	if v.cursor.row != 1 {
		t.Errorf("after j: row = %d, want 1", v.cursor.row)
	}

	// Clamped at bottom
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	v = m.(BoardViewer)
	if v.cursor.row != 1 {
		t.Errorf("at bottom: row = %d, want 1", v.cursor.row)
	}

	// Move up
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	v = m.(BoardViewer)
	if v.cursor.row != 0 {
		t.Errorf("after k: row = %d, want 0", v.cursor.row)
	}

	// Clamped at top
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	v = m.(BoardViewer)
	if v.cursor.row != 0 {
		t.Errorf("at top: row = %d, want 0", v.cursor.row)
	}
}

func TestBoardViewerEnterIssueViewer(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	m, cmd := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v = m.(BoardViewer)

	if v.mode != boardModeIssueViewer {
		t.Errorf("mode = %d, want boardModeIssueViewer", v.mode)
	}
	if cmd == nil {
		t.Error("expected command on enter")
	}
}

func TestBoardViewerEscFromIssueViewer(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	// Enter issue viewer
	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	v = m.(BoardViewer)

	// Esc back
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyEsc})
	v = m.(BoardViewer)

	if v.mode != boardModeNormal {
		t.Errorf("mode = %d, want boardModeNormal", v.mode)
	}
}

func TestBoardViewerMinimizeColumn(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	// Minimize column 0 (cursor starts there)
	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	v = m.(BoardViewer)
	if !v.columns[0].minimized {
		t.Error("expected column 0 to be minimized")
	}
	// Cursor stays on col 0 (now minimized — can navigate over it)
	if v.cursor.col != 0 {
		t.Errorf("cursor.col = %d, want 0", v.cursor.col)
	}

	// Toggle back — m on minimized col restores it
	m, _ = v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	v = m.(BoardViewer)
	if v.columns[0].minimized {
		t.Error("expected column 0 to not be minimized after toggle")
	}
}

func TestBoardViewerRefresh(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	m, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	v = m.(BoardViewer)

	if !v.loading {
		t.Error("expected loading=true on refresh")
	}
	if cmd == nil {
		t.Error("expected command on refresh")
	}
}

func TestBoardViewerQuit(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command on 'q'")
	}
}

func TestBoardViewerStatePicker(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues(), states: testStates}
	v := newLoadedBoard(api)

	_, cmd := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Error("expected command on 's' to load states")
	}
}

func TestBoardViewerCursorPreservation(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: testBoardIssues()}
	v := newLoadedBoard(api)

	// Move to second issue in Open column
	m, _ := v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	v = m.(BoardViewer)

	focused := v.focusedIssue()
	if focused == nil || focused.IDReadable != "TEST-4" {
		t.Fatalf("expected focused issue TEST-4, got %v", focused)
	}

	// Simulate refresh
	m, _ = v.Update(boardRefreshedMsg{issues: api.issues})
	v = m.(BoardViewer)

	restored := v.focusedIssue()
	if restored == nil || restored.IDReadable != "TEST-4" {
		t.Errorf("cursor not restored: got %v, want TEST-4", restored)
	}
}

func TestBoardViewerNoColumnSettings(t *testing.T) {
	board := &youtrack.Agile{
		ID:            "b2",
		Name:          "NoCols",
		CurrentSprint: &youtrack.Sprint{ID: "s1", Name: "Sprint 1"},
	}
	issues := testBoardIssues()
	api := &mockAPI{board: board, issues: issues}
	v := newLoadedBoard(api)

	// Should derive columns from issue states
	if len(v.columns) == 0 {
		t.Error("expected fallback columns derived from issue states")
	}

	// Verify issues are bucketed
	total := 0
	for ci := range v.columns {
		total += len(v.issues[ci][0])
	}
	if total != len(issues) {
		t.Errorf("total bucketed = %d, want %d", total, len(issues))
	}
}

func TestBoardViewerEmptyBoard(t *testing.T) {
	api := &mockAPI{board: testBoard, issues: nil}
	v := newLoadedBoard(api)

	// Should not panic
	view := v.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
