# RFC 002: Board Data Normalization

## Problem

The board viewer has two completely different code paths that produce the same output:

1. **Query path** (no swimlanes): `ListIssues("Board Name: {Sprint}")` → flat issue list
2. **SprintBoard path** (swimlanes): `GetSprintBoard(boardID, sprintID)` → pre-bucketed cells

This branching logic is duplicated **three times** in nearly identical functions:
- `loadBoardCmd` — initial load
- `refreshBoardCmd` — manual refresh
- `refreshAfterStateChange` — state change + refresh

Each function contains the same `swimlanesEnabled` check, sprint resolution, and path selection. Sprint resolution itself has two flavors: `resolveSprintID` (for SprintBoard API) and `resolveSprint` (for query string). A bug fix in path selection must be applied in three places.

## Proposed Interface

New package `internal/boarddata` (or `internal/tui/boarddata`).

```go
package boarddata

import "github.com/allbin/yt/internal/youtrack"

// Result is the normalized output regardless of fetch path.
type Result struct {
    Board       *youtrack.Agile
    Issues      []youtrack.Issue      // always populated (flattened from SprintBoard if needed)
    SprintBoard *youtrack.SprintBoard // nil = no swimlanes
}

type Fetcher struct { /* holds youtrack.API */ }

func New(client youtrack.API) *Fetcher

// Load fetches board metadata + issues. Used on initial load.
func (f *Fetcher) Load(boardName, sprintName string) (Result, error)

// Refresh re-fetches issues for an already-loaded board (skips metadata fetch).
func (f *Fetcher) Refresh(board *youtrack.Agile, sprintName string) (Result, error)

// SetStateAndRefresh applies a state change then refreshes.
func (f *Fetcher) SetStateAndRefresh(issueID, state string, board *youtrack.Agile, sprintName string) (Result, error)
```

**Caller usage** — three tea.Cmd functions collapse to two-liners:

```go
func loadBoardCmd(f *boarddata.Fetcher, boardName, sprintName string) tea.Cmd {
    return func() tea.Msg {
        r, err := f.Load(boardName, sprintName)
        return boardLoadedMsg{board: r.Board, issues: r.Issues, sprintBoard: r.SprintBoard, err: err}
    }
}

func refreshBoardCmd(f *boarddata.Fetcher, board *youtrack.Agile, sprintName string) tea.Cmd {
    return func() tea.Msg {
        r, err := f.Refresh(board, sprintName)
        return boardRefreshedMsg{issues: r.Issues, sprintBoard: r.SprintBoard, err: err}
    }
}

func refreshAfterStateChange(f *boarddata.Fetcher, issueID, state string, board *youtrack.Agile, sprintName string) tea.Cmd {
    return func() tea.Msg {
        r, err := f.SetStateAndRefresh(issueID, state, board, sprintName)
        return boardRefreshedMsg{issues: r.Issues, sprintBoard: r.SprintBoard, err: err}
    }
}
```

**What it hides:**
- `swimlanesEnabled` check (`SwimlaneSettings != nil && Enabled`)
- `resolveSprintID` — sprint name → ID matching or current sprint fallback
- `resolveSprint` — sprint name for query string
- Query string construction (`fmt.Sprintf("Board %s: {%s}", ...)`)
- The branch: `GetSprintBoard` vs `ListIssues`
- `flattenSprintBoard` — ensures `Result.Issues` is always populated regardless of path
- Partial-error semantics (board returned even when issue fetch fails)

## Dependency Strategy

**Remote but owned (Ports & Adapters).** `Fetcher` takes `youtrack.API` by interface at construction — constructor injection. The module sits in `internal/` and imports `youtrack.API` directly.

Testing: pass a mock `youtrack.API` to `New()`. Test that `Load` calls `GetSprintBoard` when swimlanes enabled, `ListIssues` otherwise. Test sprint resolution edge cases (no current sprint, named sprint not found).

## Testing Strategy

- **New boundary tests**: Test `Load`/`Refresh`/`SetStateAndRefresh` with mock API. Assert correct API calls made (path selection). Test sprint resolution (current, named, missing). Test `Result.Issues` always populated.
- **Old tests to delete**: The `boardLoadedMsg`/`boardRefreshedMsg` handling tests in `boardviewer_test.go` that verify branching logic become redundant — the branching is now in `boarddata/` with direct tests.
- **Test environment**: Mock `youtrack.API` (already exists).

## Implementation Recommendations

- The module should own: fetch path selection, sprint resolution, query construction, SprintBoard flattening
- It should hide: the two fetch paths, sprint name/ID resolution, query string format
- It should expose: normalized `Result` with board, issues, optional SprintBoard
- The `boardLoadedMsg`/`boardRefreshedMsg` structs stay unchanged — the module changes what produces them, not how they're consumed
- `BoardViewer` holds `*boarddata.Fetcher` instead of raw `youtrack.API` for board operations; raw client stays for non-board ops (IssueViewer, loadStatesCmd)
