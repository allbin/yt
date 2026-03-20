# RFC 003: TUI Modal Stack

## Problem

BoardViewer and IssueViewer both manage modals with identical switch-based routing:

**BoardViewer** has a `boardMode` enum with three values (normal, issueViewer, statePicker). `Update()` switches on mode, routing to `updateIssueViewer()`, `updateStatePicker()`, or `updateNormal()`. `View()` checks mode and delegates. Adding a new modal requires: new enum value, new field on struct, new case in Update, new case in View.

**IssueViewer** duplicates the same pattern with `viewerMode` (normal, statePicker) — identical StatePicker embedding and result-handling logic.

Specific friction:
- BoardViewer reads `m.issueViewer.mode == modeNormal` to decide whether `esc` should close the issue viewer or propagate to the state picker — cross-boundary internal state inspection
- StatePicker result polling (`!result.Cancelled && result.State == ""`) is duplicated in both viewers
- Adding a new modal (comment editor, filter picker) requires modifying the parent's mode enum and switch cases

## Proposed Interface

New file `internal/tui/modal/modal.go` — depends only on bubbletea.

```go
package modal

import tea "github.com/charmbracelet/bubbletea"

// Modal is a tea.Model that signals completion via Done().
type Modal interface {
    tea.Model
    Done() bool
}

// Stack manages a LIFO stack of modals.
type Stack struct { /* unexported layers []Modal */ }

// Push adds a modal and returns its Init cmd.
func (s *Stack) Push(m Modal) tea.Cmd

// Active returns true if any modal is on the stack.
func (s Stack) Active() bool

// Top returns the topmost modal, or nil.
func (s Stack) Top() Modal

// Update forwards msg to the top modal. If it becomes Done, pops it
// and returns the popped modal for result extraction via type-switch.
func (s Stack) Update(msg tea.Msg) (Stack, Modal, tea.Cmd)

// View returns the topmost modal's view, or "".
func (s Stack) View() string
```

**Modal changes needed** — existing models add `Done()`:

```go
// StatePicker — already has Result(); add Done()
func (m StatePicker) Done() bool {
    return m.result.State != "" || m.result.Cancelled
}

// IssueViewer — add closed flag + Done()
func (m IssueViewer) Done() bool { return m.closed }
```

**BoardViewer rewritten** — mode enum deleted entirely:

```go
type BoardViewer struct {
    modals modal.Stack // replaces: mode, issueViewer, statePicker
    // ...
}

func (m BoardViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.modals.Active() {
        var popped modal.Modal
        m.modals, popped, cmd = m.modals.Update(msg)
        if popped != nil {
            return m, tea.Batch(cmd, m.handleModalResult(popped))
        }
        return m, cmd
    }
    return m.updateNormal(msg)
}

func (m *BoardViewer) handleModalResult(mod modal.Modal) tea.Cmd {
    switch v := mod.(type) {
    case IssueViewer:
        m.loading = true
        return refreshBoardCmd(...)
    case StatePicker:
        r := v.Result()
        if !r.Cancelled { /* apply state change */ }
    }
    return nil
}

func (m BoardViewer) View() string {
    if m.modals.Active() { return m.modals.View() }
    // ... normal render
}
```

**What it hides:**
- Mode enum and const block — gone
- Per-modal update functions (`updateStatePicker`, `updateIssueViewer`) — collapsed into `handleModalResult`
- Type assertions (`updated.(IssueViewer)`) — inside Stack.Update
- "Is done?" polling logic — `Done()` contract on each modal
- Cross-boundary mode inspection (`issueViewer.mode == modeNormal`) — IssueViewer sets `closed=true`, board sees `popped != nil`

## Dependency Strategy

**In-process.** Package `internal/tui/modal` imports only `github.com/charmbracelet/bubbletea`. No domain types, no youtrack imports.

Typed results stay on concrete types (StatePicker.Result(), IssueViewer fields). `handleModalResult` does one type-switch per parent — the parent's package handles domain logic.

## Testing Strategy

- **New boundary tests**: Test `Stack.Push`/`Update`/`Active` with a mock Modal implementation. Test Done detection and auto-pop. Test nested stacks (IssueViewer has its own Stack for StatePicker).
- **Old tests to simplify**: Modal transition tests in `boardviewer_test.go` and `issueviewer_test.go` that test mode switching become simpler — they test `handleModalResult` directly instead of mode enum transitions.
- **Test environment**: Mock `Modal` implementation (trivial — just `Done()` returning true/false).

## Implementation Recommendations

- The module should own: modal LIFO ordering, done-detection, auto-pop, message forwarding
- It should hide: type assertions on `tea.Model.Update` return, stack management
- It should expose: `Push`, `Active`, `Update`, `View`, `Top`
- Modals remain usable standalone (StatePicker still works with `tea.NewProgram`) — `Done()` is additive
- Adding new modals: implement `Modal` interface (tea.Model + Done()), push onto stack, add one case to parent's `handleModalResult`
- The `handleModalResult` type-switch is manageable with 2-4 modal types. If it grows beyond that, consider a callback-at-push-site pattern instead.
