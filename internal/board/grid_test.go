package board

import (
	"encoding/json"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

var testAgile = &youtrack.Agile{
	ID:            "b1",
	Name:          "TestBoard",
	CurrentSprint: &youtrack.Sprint{ID: "s1", Name: "Sprint 1"},
	ColumnSettings: &youtrack.AgileColumnSettings{
		Field: &struct{ Name string `json:"name"` }{Name: "State"},
		Columns: []youtrack.AgileColumn{
			{Presentation: "Open", Ordinal: 0, FieldValues: []youtrack.AgileColumnValue{{Name: "Open"}}},
			{Presentation: "In Progress", Ordinal: 1, FieldValues: []youtrack.AgileColumnValue{{Name: "In Progress"}}},
			{Presentation: "Done", Ordinal: 2, FieldValues: []youtrack.AgileColumnValue{{Name: "Done", IsResolved: true}}},
		},
	},
}

func stateField(name string) youtrack.CustomField {
	return youtrack.CustomField{Name: "State", Value: json.RawMessage(`{"name":"` + name + `"}`)}
}

func testIssues() []youtrack.Issue {
	return []youtrack.Issue{
		{IDReadable: "T-1", Summary: "Open issue", CustomFields: []youtrack.CustomField{stateField("Open")}},
		{IDReadable: "T-2", Summary: "WIP issue", CustomFields: []youtrack.CustomField{stateField("In Progress")}},
		{IDReadable: "T-3", Summary: "Done issue", CustomFields: []youtrack.CustomField{stateField("Done")}},
		{IDReadable: "T-4", Summary: "Another open", CustomFields: []youtrack.CustomField{stateField("Open")}},
	}
}

func newGrid() *Grid {
	g := FromAgile(testAgile, testIssues(), Layout{})
	g.SetWidth(120)
	return g
}

func TestFromAgile(t *testing.T) {
	g := newGrid()

	cols := g.Columns()
	if len(cols) != 3 {
		t.Fatalf("columns = %d, want 3", len(cols))
	}
	if cols[0].Presentation != "Open" {
		t.Errorf("col[0] = %q, want Open", cols[0].Presentation)
	}
	if cols[2].Presentation != "Done" {
		t.Errorf("col[2] = %q, want Done", cols[2].Presentation)
	}
	if !cols[2].IsResolved {
		t.Error("col[2] should be resolved")
	}
}

func TestGridBuilding(t *testing.T) {
	g := newGrid()

	if len(g.CellIssues(0, 0)) != 2 {
		t.Errorf("Open = %d, want 2", len(g.CellIssues(0, 0)))
	}
	if len(g.CellIssues(1, 0)) != 1 {
		t.Errorf("InProgress = %d, want 1", len(g.CellIssues(1, 0)))
	}
	if len(g.CellIssues(2, 0)) != 1 {
		t.Errorf("Done = %d, want 1", len(g.CellIssues(2, 0)))
	}
	if len(g.AllIssues()) != 4 {
		t.Errorf("AllIssues = %d, want 4", len(g.AllIssues()))
	}
}

func TestFallbackColumns(t *testing.T) {
	board := &youtrack.Agile{ID: "b2", Name: "NoCols"}
	issues := testIssues()
	g := FromAgile(board, issues, Layout{})

	cols := g.Columns()
	if len(cols) == 0 {
		t.Fatal("expected fallback columns")
	}

	total := 0
	for ci := range cols {
		total += len(g.CellIssues(ci, 0))
	}
	if total != len(issues) {
		t.Errorf("total = %d, want %d", total, len(issues))
	}
}

func TestMoveCol(t *testing.T) {
	g := newGrid()

	col, _, _ := g.CursorPos()
	if col != 0 {
		t.Fatalf("initial col = %d", col)
	}

	g.MoveCol(1)
	col, _, _ = g.CursorPos()
	if col != 1 {
		t.Errorf("after MoveCol(1): col = %d", col)
	}

	g.MoveCol(1)
	col, _, _ = g.CursorPos()
	if col != 2 {
		t.Errorf("after MoveCol(1)x2: col = %d", col)
	}

	// Clamp right
	g.MoveCol(1)
	col, _, _ = g.CursorPos()
	if col != 2 {
		t.Errorf("clamp right: col = %d", col)
	}

	// Move left
	g.MoveCol(-1)
	col, _, _ = g.CursorPos()
	if col != 1 {
		t.Errorf("after MoveCol(-1): col = %d", col)
	}

	// Clamp left
	g.MoveCol(-1)
	g.MoveCol(-1)
	col, _, _ = g.CursorPos()
	if col != 0 {
		t.Errorf("clamp left: col = %d", col)
	}
}

func TestMoveRow(t *testing.T) {
	g := newGrid()

	_, _, row := g.CursorPos()
	if row != 0 {
		t.Fatalf("initial row = %d", row)
	}

	g.MoveRow(1)
	_, _, row = g.CursorPos()
	if row != 1 {
		t.Errorf("after MoveRow(1): row = %d", row)
	}

	// Clamp bottom
	g.MoveRow(1)
	_, _, row = g.CursorPos()
	if row != 1 {
		t.Errorf("clamp bottom: row = %d", row)
	}

	// Move up
	g.MoveRow(-1)
	_, _, row = g.CursorPos()
	if row != 0 {
		t.Errorf("after MoveRow(-1): row = %d", row)
	}

	// Clamp top
	g.MoveRow(-1)
	_, _, row = g.CursorPos()
	if row != 0 {
		t.Errorf("clamp top: row = %d", row)
	}
}

func TestFocusedIssue(t *testing.T) {
	g := newGrid()

	issue := g.FocusedIssue()
	if issue == nil || issue.IDReadable != "T-1" {
		t.Errorf("initial focused = %v, want T-1", issue)
	}

	g.MoveRow(1)
	issue = g.FocusedIssue()
	if issue == nil || issue.IDReadable != "T-4" {
		t.Errorf("after MoveRow(1): focused = %v, want T-4", issue)
	}

	g.MoveCol(1)
	issue = g.FocusedIssue()
	if issue == nil || issue.IDReadable != "T-2" {
		t.Errorf("after MoveCol(1): focused = %v, want T-2", issue)
	}
}

func TestRestoreCursor(t *testing.T) {
	g := newGrid()

	g.MoveCol(1)
	g.RestoreCursor("T-4")

	col, _, row := g.CursorPos()
	if col != 0 || row != 1 {
		t.Errorf("RestoreCursor: col=%d row=%d, want col=0 row=1", col, row)
	}
}

func TestToggleMinimize(t *testing.T) {
	g := newGrid()

	layout := g.ToggleMinimize()
	cols := g.Columns()
	if !cols[0].Minimized {
		t.Error("col 0 should be minimized")
	}
	if len(layout.MinimizedColumns) != 1 || layout.MinimizedColumns[0] != "Open" {
		t.Errorf("layout.MinimizedColumns = %v", layout.MinimizedColumns)
	}

	// Toggle back
	layout = g.ToggleMinimize()
	cols = g.Columns()
	if cols[0].Minimized {
		t.Error("col 0 should not be minimized after toggle")
	}
	if len(layout.MinimizedColumns) != 0 {
		t.Errorf("layout.MinimizedColumns = %v", layout.MinimizedColumns)
	}
}

func TestFocusedIssueMinimized(t *testing.T) {
	g := newGrid()
	g.ToggleMinimize()

	if issue := g.FocusedIssue(); issue != nil {
		t.Error("focused issue should be nil on minimized column")
	}
}

func TestColumnWidths(t *testing.T) {
	g := newGrid()

	widths := g.ColumnWidths()
	if len(widths) != 3 {
		t.Fatalf("widths = %d, want 3", len(widths))
	}
	for i, w := range widths {
		if w < minColWidth {
			t.Errorf("width[%d] = %d, want >= %d", i, w, minColWidth)
		}
	}
}

func TestColumnWidthsMinimized(t *testing.T) {
	g := newGrid()
	g.ToggleMinimize() // minimize col 0

	widths := g.ColumnWidths()
	if widths[0] != minimizedWidth {
		t.Errorf("minimized width = %d, want %d", widths[0], minimizedWidth)
	}
}

func TestVisibleColumns(t *testing.T) {
	g := newGrid()

	visible := g.VisibleColumns()
	if len(visible) == 0 {
		t.Fatal("expected visible columns")
	}
	if visible[0] != 0 {
		t.Errorf("first visible = %d, want 0", visible[0])
	}
}

func TestLayoutPersistence(t *testing.T) {
	layout := Layout{MinimizedColumns: []string{"Done"}, CollapsedLanes: []string{"lane1"}}
	g := FromAgile(testAgile, testIssues(), layout)
	g.SetWidth(120)

	cols := g.Columns()
	if !cols[2].Minimized {
		t.Error("Done should be minimized from layout")
	}

	got := g.Layout()
	if len(got.MinimizedColumns) != 1 || got.MinimizedColumns[0] != "Done" {
		t.Errorf("Layout().MinimizedColumns = %v", got.MinimizedColumns)
	}
}

func TestEmptyGrid(t *testing.T) {
	g := FromAgile(testAgile, nil, Layout{})
	g.SetWidth(120)

	if issue := g.FocusedIssue(); issue != nil {
		t.Error("no focused issue on empty grid")
	}

	cols := g.Columns()
	if len(cols) != 3 {
		t.Errorf("columns = %d, want 3", len(cols))
	}
}

func TestHasSwimlanes(t *testing.T) {
	g := newGrid()
	if g.HasSwimlanes() {
		t.Error("query-path grid should not have swimlanes")
	}
	if g.NumSwimlanes() != 1 {
		t.Errorf("NumSwimlanes = %d, want 1", g.NumSwimlanes())
	}
}

func TestFromSprintBoard(t *testing.T) {
	sb := &youtrack.SprintBoard{
		Columns: []youtrack.SprintBoardColumn{
			{
				AgileColumn: struct{ Presentation string `json:"presentation"` }{Presentation: "Open"},
				Cells: []youtrack.BoardCell{
					{
						Row:    youtrack.BoardRow{ID: "r1", Name: "Lane A", Type: "AttributeBasedSwimlane"},
						Issues: []youtrack.Issue{{IDReadable: "T-1", Summary: "A"}},
					},
					{
						Row:    youtrack.BoardRow{ID: "r2", Name: "Lane B", Type: "AttributeBasedSwimlane"},
						Issues: []youtrack.Issue{{IDReadable: "T-2", Summary: "B"}},
					},
				},
			},
			{
				AgileColumn: struct{ Presentation string `json:"presentation"` }{Presentation: "Done"},
				Cells: []youtrack.BoardCell{
					{Row: youtrack.BoardRow{ID: "r1", Name: "Lane A", Type: "AttributeBasedSwimlane"}},
					{
						Row:    youtrack.BoardRow{ID: "r2", Name: "Lane B", Type: "AttributeBasedSwimlane"},
						Issues: []youtrack.Issue{{IDReadable: "T-3", Summary: "C"}},
					},
				},
			},
		},
	}

	g := FromSprintBoard(testAgile, sb, Layout{})
	g.SetWidth(120)

	if !g.HasSwimlanes() {
		t.Fatal("expected swimlanes")
	}
	if g.NumSwimlanes() != 2 {
		t.Errorf("NumSwimlanes = %d, want 2", g.NumSwimlanes())
	}

	lanes := g.Swimlanes()
	if lanes[0].Name != "Lane A" {
		t.Errorf("lane[0] = %q", lanes[0].Name)
	}

	// Open col (0), Lane A (0): T-1
	if len(g.CellIssues(0, 0)) != 1 {
		t.Errorf("Open/LaneA = %d, want 1", len(g.CellIssues(0, 0)))
	}
	// Done col (2), Lane B (1): T-3
	if len(g.CellIssues(2, 1)) != 1 {
		t.Errorf("Done/LaneB = %d, want 1", len(g.CellIssues(2, 1)))
	}

	if len(g.AllIssues()) != 3 {
		t.Errorf("AllIssues = %d, want 3", len(g.AllIssues()))
	}
}

func TestToggleCollapse(t *testing.T) {
	sb := &youtrack.SprintBoard{
		Columns: []youtrack.SprintBoardColumn{
			{
				AgileColumn: struct{ Presentation string `json:"presentation"` }{Presentation: "Open"},
				Cells: []youtrack.BoardCell{
					{
						Row:    youtrack.BoardRow{ID: "r1", Name: "Lane A", Type: "AttributeBasedSwimlane"},
						Issues: []youtrack.Issue{{IDReadable: "T-1", Summary: "A"}},
					},
				},
			},
		},
	}

	g := FromSprintBoard(testAgile, sb, Layout{})
	g.SetWidth(120)

	layout := g.ToggleCollapse()
	lanes := g.Swimlanes()
	if !lanes[0].Collapsed {
		t.Error("lane 0 should be collapsed")
	}
	if len(layout.CollapsedLanes) != 1 {
		t.Errorf("CollapsedLanes = %v", layout.CollapsedLanes)
	}

	// Focused issue nil on collapsed lane
	if issue := g.FocusedIssue(); issue != nil {
		t.Error("no focused issue on collapsed lane")
	}
}
