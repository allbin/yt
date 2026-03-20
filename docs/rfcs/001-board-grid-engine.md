# RFC 001: Extract Board Grid Engine

## Problem

`tui/boardviewer.go` is 1224 lines — 19% of the codebase — mixing three concerns:

1. **Grid data logic**: column parsing, swimlane parsing, grid building (two paths), state-to-column mapping, issue bucketing
2. **Cursor/navigation logic**: col/row/swimlane movement with clamping, cross-swimlane wrapping, collapsed-lane skipping, cursor restore after refresh
3. **Bubbletea lifecycle**: Init/Update/View, message routing, modal management, rendering

You can't test grid construction or cursor navigation without instantiating a full bubbletea model with a mock API. The grid-building and navigation logic is pure computation — no I/O, no rendering — but it's entangled with the tea.Model.

Specific pain points:
- `buildGrid()` and `buildGridFromBoard()` are two codepaths producing the same `[][][]Issue` structure — untestable in isolation
- `moveCursorRow()` has a 40-line loop scanning adjacent swimlanes — complex logic buried in a 1200-line file
- `columnWidths()` and `ensureColumnVisible()` are pure math that require a BoardViewer to test
- Adding new navigation modes (search, filter) means modifying the monolith

## Proposed Interface

New package `internal/board` — pure computation, no bubbletea or lipgloss imports.

```go
package board

import "github.com/allbin/yt/internal/youtrack"

// Grid is the board state engine. Pointer receiver, mutable.
type Grid struct { /* unexported fields */ }

// Column is a read-only view of a column.
type Column struct {
    Presentation string
    IsResolved   bool
    Minimized    bool
    IssueCount   int
}

// Swimlane is a read-only view of a swimlane.
type Swimlane struct {
    ID        string
    Name      string
    IssueID   string
    IsOrphan  bool
    Collapsed bool
}

// Layout carries persisted UI state for the persistence bridge.
type Layout struct {
    MinimizedColumns []string // column presentation names
    CollapsedLanes   []string // swimlane IDs
}

// --- Construction ---
func FromAgile(agile *youtrack.Agile, issues []youtrack.Issue, layout Layout) *Grid
func FromSprintBoard(agile *youtrack.Agile, sb *youtrack.SprintBoard, layout Layout) *Grid

// --- Navigation ---
func (g *Grid) MoveCol(delta int)
func (g *Grid) MoveRow(delta int)
func (g *Grid) MoveSwimlane(delta int)
func (g *Grid) RestoreCursor(issueID string)
func (g *Grid) EnsureColumnVisible(width int)

// --- Queries ---
func (g *Grid) FocusedIssue() *youtrack.Issue
func (g *Grid) CursorPos() (col, swimlane, row int)
func (g *Grid) CellIssues(col, swimlane int) []youtrack.Issue
func (g *Grid) AllIssues() []youtrack.Issue
func (g *Grid) HasSwimlanes() bool

// --- Layout ---
func (g *Grid) Columns() []Column
func (g *Grid) Swimlanes() []Swimlane
func (g *Grid) ColumnWidths(totalWidth int) []int
func (g *Grid) VisibleColumns(width int) []int
func (g *Grid) ToggleMinimize() Layout
func (g *Grid) ToggleCollapse() Layout
func (g *Grid) Layout() Layout
```

**Caller usage** (BoardViewer shrinks from ~1200 to ~400 lines):

```go
// On boardLoadedMsg:
layout := loadLayout(m.appState, msg.board.ID)
if msg.sprintBoard != nil {
    m.grid = board.FromSprintBoard(msg.board, msg.sprintBoard, layout)
} else {
    m.grid = board.FromAgile(msg.board, msg.issues, layout)
}
m.grid.EnsureColumnVisible(m.width)
m.rebuildContent()

// Key handling:
case "h": m.grid.MoveCol(-1)
case "j": m.grid.MoveRow(1)
case "m": saveLayout(m.appState, m.board.ID, m.grid.ToggleMinimize())

// Rendering:
for _, ci := range m.grid.VisibleColumns(m.width) {
    issues := m.grid.CellIssues(ci, sl)
    // render cards...
}
```

**What it hides internally:**
- `parseColumns()` — column derivation from ColumnSettings + ordinal sort + state-fallback
- `parseSwimlanes()` — BoardRow type dispatch (IssueBasedSwimlane, AttributeBasedSwimlane, OrphanRow)
- `buildGrid()` / `buildGridFromBoard()` — two codepaths, stateToCol mapping, allIssues accumulation
- `moveCursorRow` cross-swimlane wrapping with collapsed-lane skipping
- `clampRow()` — swimlane validity, scanning for first non-locked non-empty lane
- `ensureColumnVisible()` — column-offset loop with width accumulation
- `columnWidths()` — minimized vs full width arithmetic
- `numSwimlanes()` — the 1-vs-N invariant
- `restoreCursor` — 3-nested-loop search + clampRow fallback

## Dependency Strategy

**In-process** — pure computation. `board.Grid` imports only `internal/youtrack` for `Issue`, `Agile`, `SprintBoard` types. No bubbletea, no lipgloss, no state package.

The `Layout` type bridges persistence: BoardViewer converts between `state.AppState` and `board.Layout` via two small adapter functions (~10 lines in boardviewer.go).

Note: `focusedCardPosition` (used for auto-scroll) stays in BoardViewer — it depends on rendered card heights from lipgloss, which is a rendering concern.

## Testing Strategy

- **New boundary tests**: Test `FromAgile`/`FromSprintBoard` + navigation sequences directly. Assert `FocusedIssue()`, `CursorPos()`, `CellIssues()` after `MoveCol`/`MoveRow`/`MoveSwimlane`. Test `ColumnWidths` and `VisibleColumns` with various terminal widths. Test `RestoreCursor` after grid rebuild. Test `ToggleMinimize`/`ToggleCollapse` + `Layout()` round-trip.
- **Old tests to simplify**: `boardviewer_test.go` (402L) currently tests grid logic through tea.Model scaffolding — the grid-related tests move to `board/` as direct function calls. BoardViewer tests shrink to focus on message routing and rendering.
- **Test environment**: None — pure Go, no mocks needed.

## Implementation Recommendations

- The module should own: column/swimlane parsing, grid construction (both paths), cursor navigation, column width calculation, horizontal scroll, UI state (minimized/collapsed)
- It should hide: the two grid-build paths, column derivation fallback, swimlane type dispatch, cross-swimlane cursor wrapping, colOffset management
- It should expose: read-only column/swimlane metadata, cell contents, cursor position, layout state for persistence
- Migration: extract incrementally — move `buildGrid` first, then navigation, then column widths. Each step is independently testable.
