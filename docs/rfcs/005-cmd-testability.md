# RFC 005: cmd/ Integration Testability

## Problem

The `cmd/` package has 18 command files (~1100 lines) but only 1 test (`issue_update_test.go`, 33 lines). Commands are thin Cobra wrappers but contain real logic:

- `issueIDFromArgs` — resolves issue ID from args or git branch name
- `resolveAssignee` — user name → login via API
- Issue list query building from flags (`--project`, `--state`, `--assignee`, `--query`, `--limit`)
- JSON vs text output routing
- Error formatting (`formatError` handles `APIError` vs generic errors)

The root cause of untestability: `newClient()` in `root.go` reads env vars via viper and creates a real HTTP client. Every command calls it at the start of `RunE`. No way to inject a mock.

```go
func newClient() (youtrack.API, error) {
    u := viper.GetString("URL")    // requires YOUTRACK_URL
    token := viper.GetString("TOKEN") // requires YOUTRACK_TOKEN
    return youtrack.NewClient(u, token), nil
}
```

## Proposed Interface

Minimal change: replace `newClient()` calls with an injectable factory variable.

```go
// cmd/root.go — add one variable
var apiFactory func() (youtrack.API, error)

func init() {
    apiFactory = newClient // default production behavior
}
```

Each command changes one call:

```go
// before:
client, err := newClient()

// after:
client, err := apiFactory()
```

**Test helper** (in `cmd/testing_test.go`):

```go
func setupTest(t *testing.T, api youtrack.API) func(args ...string) (string, error) {
    t.Helper()
    apiFactory = func() (youtrack.API, error) { return api, nil }
    t.Cleanup(func() { apiFactory = newClient })

    return func(args ...string) (string, error) {
        buf := new(bytes.Buffer)
        rootCmd.SetOut(buf)
        rootCmd.SetErr(buf)
        rootCmd.SetArgs(args)
        t.Cleanup(func() { rootCmd.SetOut(nil); rootCmd.SetErr(nil) })
        err := rootCmd.Execute()
        return buf.String(), err
    }
}
```

**Test examples** (2-3 lines each):

```go
func TestRunIssue(t *testing.T) {
    run := setupTest(t, &mockAPI{
        issue: &youtrack.Issue{IDReadable: "TEST-1", Summary: "Fix login"},
    })
    out, err := run("issue", "TEST-1")
    assert.NoError(t, err)
    assert.Contains(t, out, "TEST-1")
}

func TestRunIssueList(t *testing.T) {
    run := setupTest(t, &mockAPI{
        issues: []youtrack.Issue{{IDReadable: "A-1"}, {IDReadable: "A-2"}},
    })
    out, err := run("issue", "list")
    assert.NoError(t, err)
    assert.Contains(t, out, "A-1")
}

func TestRunIssueJSON(t *testing.T) {
    run := setupTest(t, &mockAPI{
        issue: &youtrack.Issue{IDReadable: "TEST-1"},
    })
    out, err := run("issue", "TEST-1", "--json")
    assert.NoError(t, err)
    assert.Contains(t, out, `"idReadable"`)
}
```

**What it hides:** Nothing new — this is about exposure, not encapsulation. `newClient()` still reads viper; it just isn't called in tests.

## Dependency Strategy

**Remote but owned (Ports & Adapters).** The `youtrack.API` interface already exists as the port. The `apiFactory` variable is the injection point. Production path: `apiFactory = newClient` → viper → HTTP client. Test path: `apiFactory = func() { return mock, nil }`.

The mock already exists in `tui/mock_test.go` — copy or move to `cmd/mock_test.go` (or share via `internal/testutil/mock.go`).

## Testing Strategy

- **New tests to write**: Test every command's happy path (text output), JSON output, error handling. Test `issueIDFromArgs` with explicit args and git branch fallback. Test query building from flags. Test `formatError` with `APIError` vs generic error.
- **Old tests**: `issue_update_test.go` (33L) stays — it already tests via the API interface.
- **Test environment**: Mock `youtrack.API`. Note: tests must not use `t.Parallel()` due to shared `apiFactory` variable and cobra's global `rootCmd` state.

## Implementation Recommendations

- The module should own: nothing new — `cmd/` stays thin
- Migration: add `apiFactory` variable to `root.go`, replace `newClient()` calls in all 18 files (mechanical, one line each), add `setupTest` helper, write tests incrementally
- Caveats:
  - Cobra's `rootCmd` is a package-level var — flag state accumulates between test runs. If this causes issues, refactor `rootCmd` into a `newRootCmd()` factory and rebuild per test.
  - `issueIDFromArgs` calls `git.CurrentBranch()` which shells out. For deterministic testing, make the git function injectable too (separate concern, can be done later).
  - Tests run sequentially (no `t.Parallel`) due to shared mutable state. Acceptable for a test suite this size.
