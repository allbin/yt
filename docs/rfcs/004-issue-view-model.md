# RFC 004: Issue View Model

## Problem

Both `format/` and `tui/` import `youtrack/types.go` directly and call `issue.Field("State")`, `issue.Field("Priority")`, etc. extensively. Each `Field()` call does a linear scan through `CustomFields` and JSON-unmarshals the value (trying object, array, then string shapes). On a board with 50 cards rendered per frame, 6 fields per card = ~300 unmarshal operations per render tick, all producing the same strings.

Secondary issues:
- `commentAuthor` logic (nil-check Author, prefer FullName over Login) is duplicated in `format/comment.go` and `tui/issueviewer.go`
- Field names ("State", "Priority", "Assignee", "Type", "Subsystem") are magic strings scattered across 3 packages — a typo silently returns ""
- Test fixtures for `format/` and `tui/` require constructing `youtrack.Issue` with JSON-encoded `CustomFields` — verbose and fragile

Note: this is the lowest-severity candidate. `DisplayValue()` already hides JSON complexity. The main wins are performance and deduplication.

## Proposed Interface

Method on `*youtrack.Issue` in the existing `youtrack/` package — no new package needed.

```go
// youtrack/view.go

// IssueView is a pre-resolved, display-ready snapshot of an Issue.
type IssueView struct {
    ID          string
    Summary     string
    Description string
    State       string
    Priority    string
    Assignee    string
    Type        string
    Subsystem   string
    Tags        string // comma-joined
    IsResolved  bool
}

func (i *Issue) View() IssueView {
    return IssueView{
        ID:          i.IDReadable,
        Summary:     i.Summary,
        Description: i.Desc(),
        State:       i.Field("State"),
        Priority:    i.Field("Priority"),
        Assignee:    i.Field("Assignee"),
        Type:        i.Field("Type"),
        Subsystem:   i.Field("Subsystem"),
        Tags:        i.TagNames(),
        IsResolved:  i.Resolved != nil,
    }
}

// CommentView is a pre-resolved comment for display.
type CommentView struct {
    Author  string
    Created int64
    Text    string
}

func (c *Comment) View() CommentView {
    author := "Unknown"
    if c.Author != nil {
        if c.Author.FullName != "" {
            author = c.Author.FullName
        } else {
            author = c.Author.Login
        }
    }
    return CommentView{Author: author, Created: c.Created, Text: c.Text}
}
```

**Caller usage** — convert once, access directly:

```go
// tui/issuecard.go — before:
priority := issue.Field("Priority")
assignee := issue.Field("Assignee")

// after:
v := issue.View()
priority := v.Priority
assignee := v.Assignee

// tui/issueviewer.go — before:
state := m.issue.Field("State")
priority := m.issue.Field("Priority")
assignee := m.issue.Field("Assignee")

// after:
v := m.issueView  // computed once on load
state := v.State
```

**What it hides:**
- Repeated JSON unmarshaling in `DisplayValue()` — done once at conversion
- Linear scan through `CustomFields` — done once
- Nil-pointer handling for `Description` and `Comment.Author`
- Magic string field names centralized in one function
- `commentAuthor` deduplication

## Dependency Strategy

**In-process.** `IssueView` lives in `youtrack/` alongside `Issue` — no new package, no import cycle. `format/` and `tui/` can adopt `IssueView` incrementally; existing `*youtrack.Issue` parameters don't need to change immediately.

## Testing Strategy

- **New boundary tests**: Test `Issue.View()` with various CustomField shapes (object, array, string, null). Test `Comment.View()` with nil Author, FullName-only, Login-only. Test that `IssueView` fields match expected values.
- **Old tests unchanged**: Format and TUI tests continue to work — `IssueView` is additive. Tests that currently construct `youtrack.Issue` with encoded CustomFields can optionally switch to `IssueView{}` struct literals (much simpler).
- **Test environment**: None.

## Implementation Recommendations

- The module should own: one-time field resolution, nil-guarded access, comment author logic
- It should hide: CustomField JSON shapes, Field() linear scan, nil Description pointer
- It should expose: flat struct with pre-resolved string fields
- Migration path: add `View()` method and `IssueView` type. Adopt in callers incrementally — each file can switch from `issue.Field("X")` to `issue.View().X` independently. No big-bang migration needed.
- Consider: TUI models could store `IssueView` instead of `*Issue` after initial load, avoiding repeated `View()` calls per render. The board grid engine (RFC 001) could store `IssueView` per cell.
