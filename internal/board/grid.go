package board

import (
	"sort"

	"github.com/allbin/yt/internal/youtrack"
)

const (
	minColWidth    = 28
	minimizedWidth = 5
)

// Column is a read-only view of a board column.
type Column struct {
	Presentation string
	StateNames   []string
	IsResolved   bool
	Minimized    bool
	IssueCount   int
}

// Swimlane is a read-only view of a board swimlane.
type Swimlane struct {
	ID        string
	Name      string
	IssueID   string
	Summary   string
	IsOrphan  bool
	Collapsed bool
}

// Layout carries persisted UI state for the persistence bridge.
type Layout struct {
	MinimizedColumns []string
	CollapsedLanes   []string
}

// Grid is the board data and navigation engine.
// Use FromAgile or FromSprintBoard to construct.
type Grid struct {
	columns   []columnDef
	swimlanes []swimlaneDef
	fieldName string

	issues    [][][]youtrack.Issue // [col][swimlane][]Issue
	allIssues []youtrack.Issue

	cursor    cursor
	colOffset int
	width     int
}

type columnDef struct {
	presentation string
	ordinal      int
	stateNames   []string
	isResolved   bool
	minimized    bool
}

type swimlaneDef struct {
	id        string
	name      string
	issueID   string
	summary   string
	isOrphan  bool
	collapsed bool
}

type cursor struct {
	col      int
	row      int
	swimlane int
}

// --- Construction ---

// FromAgile builds a Grid from a board config + flat issue list (query path).
func FromAgile(agile *youtrack.Agile, issues []youtrack.Issue, layout Layout) *Grid {
	g := &Grid{allIssues: issues}
	g.parseColumns(agile)
	g.applyLayout(layout)
	g.buildGrid()
	return g
}

// FromSprintBoard builds a Grid from a board config + sprint board response (swimlane path).
func FromSprintBoard(agile *youtrack.Agile, sb *youtrack.SprintBoard, layout Layout) *Grid {
	g := &Grid{}
	g.parseColumns(agile)
	g.parseSwimlanes(sb)
	g.applyLayout(layout)
	g.buildGridFromBoard(sb)
	return g
}

func (g *Grid) parseColumns(agile *youtrack.Agile) {
	g.columns = nil
	g.fieldName = "State"

	if agile.ColumnSettings != nil && len(agile.ColumnSettings.Columns) > 0 {
		if agile.ColumnSettings.Field != nil {
			g.fieldName = agile.ColumnSettings.Field.Name
		}
		for _, col := range agile.ColumnSettings.Columns {
			cd := columnDef{
				presentation: col.Presentation,
				ordinal:      col.Ordinal,
			}
			for _, fv := range col.FieldValues {
				cd.stateNames = append(cd.stateNames, fv.Name)
				if fv.IsResolved {
					cd.isResolved = true
				}
			}
			g.columns = append(g.columns, cd)
		}
		sort.Slice(g.columns, func(i, j int) bool {
			return g.columns[i].ordinal < g.columns[j].ordinal
		})
		return
	}

	// Fallback: derive columns from unique state values
	seen := map[string]bool{}
	var states []string
	for _, issue := range g.allIssues {
		s := issue.Field(g.fieldName)
		if s != "" && !seen[s] {
			seen[s] = true
			states = append(states, s)
		}
	}
	for i, s := range states {
		g.columns = append(g.columns, columnDef{
			presentation: s,
			ordinal:      i,
			stateNames:   []string{s},
		})
	}
}

func (g *Grid) parseSwimlanes(sb *youtrack.SprintBoard) {
	g.swimlanes = nil
	if sb == nil || len(sb.Columns) == 0 || len(sb.Columns[0].Cells) == 0 {
		return
	}
	for _, cell := range sb.Columns[0].Cells {
		row := cell.Row
		sd := swimlaneDef{id: row.ID}
		switch row.Type {
		case "IssueBasedSwimlane":
			if row.Issue != nil {
				sd.issueID = row.Issue.IDReadable
				sd.summary = row.Issue.Summary
				sd.name = row.Issue.IDReadable + " " + row.Issue.Summary
			}
		case "AttributeBasedSwimlane":
			sd.name = row.Name
		case "OrphanRow":
			sd.name = "Uncategorized"
			sd.isOrphan = true
		default:
			sd.name = row.Name
			if sd.name == "" {
				sd.name = "Unknown"
			}
		}
		g.swimlanes = append(g.swimlanes, sd)
	}
}

func (g *Grid) applyLayout(layout Layout) {
	minimized := make(map[string]bool, len(layout.MinimizedColumns))
	for _, name := range layout.MinimizedColumns {
		minimized[name] = true
	}
	for i := range g.columns {
		g.columns[i].minimized = minimized[g.columns[i].presentation]
	}

	collapsed := make(map[string]bool, len(layout.CollapsedLanes))
	for _, id := range layout.CollapsedLanes {
		collapsed[id] = true
	}
	for i := range g.swimlanes {
		g.swimlanes[i].collapsed = collapsed[g.swimlanes[i].id]
	}
}

func (g *Grid) buildGrid() {
	numCols := len(g.columns)
	numSL := g.numSwimlanes()

	g.issues = make([][][]youtrack.Issue, numCols)
	for c := range numCols {
		g.issues[c] = make([][]youtrack.Issue, numSL)
	}

	stateToCol := map[string]int{}
	for ci, col := range g.columns {
		for _, sn := range col.stateNames {
			stateToCol[sn] = ci
		}
	}

	for _, issue := range g.allIssues {
		colIdx, ok := stateToCol[issue.Field(g.fieldName)]
		if !ok {
			continue
		}
		g.issues[colIdx][0] = append(g.issues[colIdx][0], issue)
	}
}

func (g *Grid) buildGridFromBoard(sb *youtrack.SprintBoard) {
	numCols := len(g.columns)
	numSL := g.numSwimlanes()

	g.issues = make([][][]youtrack.Issue, numCols)
	for c := range numCols {
		g.issues[c] = make([][]youtrack.Issue, numSL)
	}

	colMap := map[string]int{}
	for ci, col := range g.columns {
		colMap[col.presentation] = ci
	}

	g.allIssues = nil
	for _, sbCol := range sb.Columns {
		ci, ok := colMap[sbCol.AgileColumn.Presentation]
		if !ok {
			continue
		}
		for si, cell := range sbCol.Cells {
			if si >= numSL {
				break
			}
			g.issues[ci][si] = cell.Issues
			g.allIssues = append(g.allIssues, cell.Issues...)
		}
	}
}

func (g *Grid) numSwimlanes() int {
	if len(g.swimlanes) == 0 {
		return 1
	}
	return len(g.swimlanes)
}

// --- Navigation ---

// MoveCol moves the cursor left (delta<0) or right (delta>0).
func (g *Grid) MoveCol(delta int) {
	if len(g.columns) == 0 {
		return
	}
	g.cursor.col = max(min(g.cursor.col+delta, len(g.columns)-1), 0)
	if !g.columns[g.cursor.col].minimized {
		g.clampRow()
	}
	g.ensureColumnVisible()
}

// MoveRow moves the cursor up (delta<0) or down (delta>0) within the column,
// crossing swimlane boundaries when at the edge.
func (g *Grid) MoveRow(delta int) {
	if g.cursor.col >= len(g.columns) || g.columns[g.cursor.col].minimized {
		return
	}
	if g.isSwimlaneLocked(g.cursor.swimlane) {
		g.MoveSwimlane(delta)
		return
	}
	sl := g.cursor.swimlane
	issues := g.issues[g.cursor.col][sl]

	next := g.cursor.row + delta
	if next >= 0 && next < len(issues) {
		g.cursor.row = next
		return
	}

	numSL := g.numSwimlanes()
	if delta > 0 && next >= len(issues) {
		for s := sl + 1; s < numSL; s++ {
			if g.isSwimlaneLocked(s) {
				continue
			}
			if len(g.issues[g.cursor.col][s]) > 0 {
				g.cursor.swimlane = s
				g.cursor.row = 0
				return
			}
		}
		g.cursor.row = max(len(issues)-1, 0)
	} else if delta < 0 && next < 0 {
		for s := sl - 1; s >= 0; s-- {
			if g.isSwimlaneLocked(s) {
				continue
			}
			if len(g.issues[g.cursor.col][s]) > 0 {
				g.cursor.swimlane = s
				g.cursor.row = len(g.issues[g.cursor.col][s]) - 1
				return
			}
		}
		g.cursor.row = 0
	}
}

// MoveSwimlane moves the cursor to the next (delta>0) or previous (delta<0) swimlane.
func (g *Grid) MoveSwimlane(delta int) {
	if len(g.swimlanes) == 0 || g.columns[g.cursor.col].minimized {
		return
	}
	numSL := g.numSwimlanes()
	next := g.cursor.swimlane + delta

	if delta > 0 {
		for s := next; s < numSL; s++ {
			if len(g.issues[g.cursor.col][s]) > 0 {
				g.cursor.swimlane = s
				g.cursor.row = 0
				return
			}
		}
	} else {
		for s := next; s >= 0; s-- {
			if len(g.issues[g.cursor.col][s]) > 0 {
				g.cursor.swimlane = s
				g.cursor.row = 0
				return
			}
		}
	}
}

// RestoreCursor moves the cursor to the issue with the given ID.
// Falls back to clamping if not found.
func (g *Grid) RestoreCursor(issueID string) {
	if issueID == "" {
		g.clampRow()
		return
	}
	for ci := range g.columns {
		if g.columns[ci].minimized {
			continue
		}
		for sl := range g.numSwimlanes() {
			for ri, issue := range g.issues[ci][sl] {
				if issue.IDReadable == issueID {
					g.cursor.col = ci
					g.cursor.swimlane = sl
					g.cursor.row = ri
					return
				}
			}
		}
	}
	g.clampRow()
}

func (g *Grid) isSwimlaneLocked(sl int) bool {
	return len(g.swimlanes) > 0 && sl < len(g.swimlanes) && g.swimlanes[sl].collapsed
}

func (g *Grid) clampRow() {
	col := g.cursor.col
	if col >= len(g.columns) || g.columns[col].minimized {
		return
	}
	numSL := g.numSwimlanes()
	sl := g.cursor.swimlane
	if sl >= numSL {
		sl = 0
		g.cursor.swimlane = 0
	}

	if len(g.issues[col][sl]) == 0 || g.isSwimlaneLocked(sl) {
		for s := range numSL {
			if !g.isSwimlaneLocked(s) && len(g.issues[col][s]) > 0 {
				g.cursor.swimlane = s
				g.cursor.row = 0
				return
			}
		}
		g.cursor.row = 0
		return
	}

	issues := g.issues[col][sl]
	if g.cursor.row >= len(issues) {
		g.cursor.row = max(len(issues)-1, 0)
	}
}

// --- Queries ---

// FocusedIssue returns the issue under the cursor, or nil.
func (g *Grid) FocusedIssue() *youtrack.Issue {
	col := g.cursor.col
	if col >= len(g.columns) || g.columns[col].minimized {
		return nil
	}
	sl := g.cursor.swimlane
	if sl >= g.numSwimlanes() {
		return nil
	}
	if g.isSwimlaneLocked(sl) {
		return nil
	}
	issues := g.issues[col][sl]
	if g.cursor.row >= len(issues) {
		return nil
	}
	return &issues[g.cursor.row]
}

// CursorPos returns the current cursor position.
func (g *Grid) CursorPos() (col, swimlane, row int) {
	return g.cursor.col, g.cursor.swimlane, g.cursor.row
}

// CellIssues returns the issues at the given column and swimlane.
func (g *Grid) CellIssues(col, swimlane int) []youtrack.Issue {
	if col < 0 || col >= len(g.columns) {
		return nil
	}
	numSL := g.numSwimlanes()
	if swimlane < 0 || swimlane >= numSL {
		return nil
	}
	return g.issues[col][swimlane]
}

// AllIssues returns a flat slice of all issues in the grid.
func (g *Grid) AllIssues() []youtrack.Issue {
	return g.allIssues
}

// HasSwimlanes reports whether this grid has swimlane rows.
func (g *Grid) HasSwimlanes() bool {
	return len(g.swimlanes) > 0
}

// NumSwimlanes returns the number of logical swimlanes (minimum 1).
func (g *Grid) NumSwimlanes() int {
	return g.numSwimlanes()
}

// FieldName returns the column-grouping field name (usually "State").
func (g *Grid) FieldName() string {
	return g.fieldName
}

// --- Layout ---

// SetWidth stores the terminal width and adjusts horizontal scroll.
func (g *Grid) SetWidth(width int) {
	g.width = width
	g.ensureColumnVisible()
}

// Columns returns read-only column metadata.
func (g *Grid) Columns() []Column {
	cols := make([]Column, len(g.columns))
	for i, c := range g.columns {
		count := 0
		for sl := range g.numSwimlanes() {
			count += len(g.issues[i][sl])
		}
		cols[i] = Column{
			Presentation: c.presentation,
			StateNames:   c.stateNames,
			IsResolved:   c.isResolved,
			Minimized:    c.minimized,
			IssueCount:   count,
		}
	}
	return cols
}

// Swimlanes returns read-only swimlane metadata.
func (g *Grid) Swimlanes() []Swimlane {
	lanes := make([]Swimlane, len(g.swimlanes))
	for i, s := range g.swimlanes {
		lanes[i] = Swimlane{
			ID:        s.id,
			Name:      s.name,
			IssueID:   s.issueID,
			Summary:   s.summary,
			IsOrphan:  s.isOrphan,
			Collapsed: s.collapsed,
		}
	}
	return lanes
}

// ColumnWidths returns computed character widths for all columns.
func (g *Grid) ColumnWidths() []int {
	numCols := len(g.columns)
	widths := make([]int, numCols)

	minimizedSpace := 0
	visibleCount := 0
	for _, col := range g.columns {
		if col.minimized {
			minimizedSpace += minimizedWidth
		} else {
			visibleCount++
		}
	}
	if visibleCount == 0 {
		for i := range widths {
			widths[i] = minimizedWidth
		}
		return widths
	}

	available := max(g.width-minimizedSpace, minColWidth)
	maxFit := max(available/minColWidth, 1)
	colWidth := minColWidth
	if visibleCount <= maxFit {
		colWidth = available / visibleCount
	}

	for i, col := range g.columns {
		if col.minimized {
			widths[i] = minimizedWidth
		} else {
			widths[i] = colWidth
		}
	}
	return widths
}

// VisibleColumns returns column indices that fit within the terminal width,
// starting from the internal scroll offset.
func (g *Grid) VisibleColumns() []int {
	widths := g.ColumnWidths()
	var cols []int
	used := 0
	for ci := g.colOffset; ci < len(g.columns); ci++ {
		if used >= g.width && len(cols) > 0 {
			break
		}
		cols = append(cols, ci)
		used += widths[ci]
	}
	return cols
}

func (g *Grid) ensureColumnVisible() {
	if len(g.columns) == 0 || g.width == 0 {
		return
	}
	g.cursor.col = max(min(g.cursor.col, len(g.columns)-1), 0)

	widths := g.ColumnWidths()

	if g.cursor.col < g.colOffset {
		g.colOffset = g.cursor.col
	}

	for {
		used := 0
		cursorFits := false
		for ci := g.colOffset; ci < len(g.columns); ci++ {
			used += widths[ci]
			if used > g.width && ci > g.colOffset {
				break
			}
			if ci == g.cursor.col {
				cursorFits = true
				break
			}
		}
		if cursorFits {
			break
		}
		g.colOffset++
		if g.colOffset >= len(g.columns) {
			g.colOffset = g.cursor.col
			break
		}
	}
}

// ToggleMinimize toggles the minimized state of the cursor column.
// Returns the updated Layout for persistence.
func (g *Grid) ToggleMinimize() Layout {
	if g.cursor.col >= len(g.columns) {
		return g.Layout()
	}
	g.columns[g.cursor.col].minimized = !g.columns[g.cursor.col].minimized
	if !g.columns[g.cursor.col].minimized {
		g.clampRow()
	}
	return g.Layout()
}

// ToggleCollapse toggles the collapsed state of the cursor swimlane.
// Returns the updated Layout for persistence.
func (g *Grid) ToggleCollapse() Layout {
	if len(g.swimlanes) == 0 {
		return g.Layout()
	}
	sl := g.cursor.swimlane
	if sl >= len(g.swimlanes) {
		return g.Layout()
	}
	g.swimlanes[sl].collapsed = !g.swimlanes[sl].collapsed
	return g.Layout()
}

// Layout returns the current persisted UI state.
func (g *Grid) Layout() Layout {
	var minimized []string
	for _, col := range g.columns {
		if col.minimized {
			minimized = append(minimized, col.presentation)
		}
	}
	var collapsed []string
	for _, sl := range g.swimlanes {
		if sl.collapsed {
			collapsed = append(collapsed, sl.id)
		}
	}
	return Layout{
		MinimizedColumns: minimized,
		CollapsedLanes:   collapsed,
	}
}
