package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/allbin/yt/internal/youtrack"
)

func activity(issue, field, from, to string, ts int64) youtrack.FieldActivity {
	val := func(name string) json.RawMessage {
		if name == "" {
			return json.RawMessage(`[]`)
		}
		return json.RawMessage(`[{"name":"` + name + `"}]`)
	}
	return youtrack.FieldActivity{
		ID:        issue + "-" + to,
		Timestamp: ts,
		Author:    &youtrack.User{Login: "alice", FullName: "Alice"},
		Target:    &youtrack.ActivityTarget{IDReadable: issue, Summary: "Summary " + issue},
		Field:     &youtrack.ActivityField{Name: field},
		Added:     val(to),
		Removed:   val(from),
	}
}

func testActivities() []youtrack.FieldActivity {
	return []youtrack.FieldActivity{
		activity("AX-1", "State", "Planned", "In Progress", 1759300000000),
		activity("AX-1", "Assignee", "", "Alice", 1759300000000),
		activity("AX-2", "State", "In Progress", "In Review", 1759400000000),
		activity("HK-7", "State", "Open", "Done", 1759500000000),
	}
}

func stageBoard() *youtrack.Agile {
	return &youtrack.Agile{
		ID:   "A1",
		Name: "AllTix",
		ColumnSettings: &youtrack.AgileColumnSettings{
			Field: &struct {
				Name string `json:"name"`
			}{Name: "State"},
			Columns: []youtrack.AgileColumn{
				{Presentation: "Todo", FieldValues: []youtrack.AgileColumnValue{{Name: "Planned"}}},
				{Presentation: "Doing", FieldValues: []youtrack.AgileColumnValue{{Name: "In Progress"}, {Name: "In Review"}}},
			},
		},
	}
}

func TestRunTransitionsDefaults(t *testing.T) {
	api := &mockAPI{activities: testActivities()}
	run := setupTest(t, api)

	before := time.Now()
	out, err := run("transitions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := api.activityFilter
	if f.Author != "me" {
		t.Errorf("author = %q, want me", f.Author)
	}
	if age := before.Sub(f.Start); age < 7*24*time.Hour-time.Minute || age > 7*24*time.Hour+time.Minute {
		t.Errorf("start is %v ago, want 7d", age)
	}
	if f.End.Before(before) {
		t.Errorf("end = %v, want now", f.End)
	}
	if f.IssueQuery != "" {
		t.Errorf("issueQuery = %q, want empty", f.IssueQuery)
	}

	for _, want := range []string{"AX-1", "Planned", "In Progress", "AX-2", "In Review", "HK-7", "Alice"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRunTransitionsJSON(t *testing.T) {
	run := setupTest(t, &mockAPI{activities: testActivities()})

	out, err := run("transitions", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []youtrack.Transition
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(got) != 3 {
		t.Fatalf("got %d transitions, want 3 State changes: %+v", len(got), got)
	}
	tr := got[0]
	if tr.Issue != "AX-1" || tr.Summary != "Summary AX-1" || tr.From != "Planned" || tr.To != "In Progress" {
		t.Errorf("first = %+v", tr)
	}
	if tr.Author == nil || tr.Author.Login != "alice" || tr.Timestamp != 1759300000000 || tr.Time == "" {
		t.Errorf("first actor/time = %+v", tr)
	}
	if tr.FromColumn != "" || tr.ToColumn != "" {
		t.Errorf("columns set without --board: %+v", tr)
	}
}

func TestRunTransitionsEmptyJSON(t *testing.T) {
	run := setupTest(t, &mockAPI{})

	out, err := run("transitions", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("out = %q, want []", out)
	}
}

func TestRunTransitionsFilters(t *testing.T) {
	api := &mockAPI{activities: testActivities()}
	run := setupTest(t, api)

	out, err := run("transitions", "-u", "alice", "--since", "2026-09-28", "--until", "2026-10-04",
		"-p", "AX", "-q", "Type: Bug", "--field", "assignee", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := api.activityFilter
	if f.Author != "alice" {
		t.Errorf("author = %q, want alice", f.Author)
	}
	wantStart := time.Date(2026, 9, 28, 0, 0, 0, 0, time.Local)
	wantEnd := time.Date(2026, 10, 5, 0, 0, 0, 0, time.Local).Add(-time.Millisecond)
	if !f.Start.Equal(wantStart) || !f.End.Equal(wantEnd) {
		t.Errorf("range = %v .. %v, want %v .. %v", f.Start, f.End, wantStart, wantEnd)
	}
	if f.IssueQuery != "project: AX Type: Bug" {
		t.Errorf("issueQuery = %q", f.IssueQuery)
	}

	var got []youtrack.Transition
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Field != "Assignee" || got[0].To != "Alice" {
		t.Errorf("got %+v, want the one Assignee change", got)
	}
}

func TestRunTransitionsAllUsers(t *testing.T) {
	api := &mockAPI{}
	run := setupTest(t, api)

	if _, err := run("transitions", "-u", "all"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if api.activityFilter.Author != "" {
		t.Errorf("author = %q, want no author filter", api.activityFilter.Author)
	}
}

func TestRunTransitionsBoard(t *testing.T) {
	api := &mockAPI{
		activities:  testActivities(),
		board:       stageBoard(),
		boardIssues: []string{"AX-1", "ax-2"},
	}
	run := setupTest(t, api)

	out, err := run("transitions", "--board", "AllTix", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []youtrack.Transition
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d transitions, want the 2 on the board: %+v", len(got), got)
	}
	if got[0].FromColumn != "Todo" || got[0].ToColumn != "Doing" {
		t.Errorf("AX-1 columns = %q -> %q, want Todo -> Doing", got[0].FromColumn, got[0].ToColumn)
	}
	if got[1].Issue != "AX-2" || got[1].FromColumn != "Doing" || got[1].ToColumn != "Doing" {
		t.Errorf("AX-2 = %+v, want a move within Doing", got[1])
	}
}

func TestRunTransitionsBoardFieldMismatch(t *testing.T) {
	run := setupTest(t, &mockAPI{board: stageBoard()})

	_, err := run("transitions", "--board", "AllTix", "--field", "Priority")
	if err == nil || !strings.Contains(err.Error(), "bound to State") {
		t.Errorf("err = %v, want board field mismatch", err)
	}
}

func TestRunTransitionsBadRange(t *testing.T) {
	run := setupTest(t, &mockAPI{})

	if _, err := run("transitions", "--since", "2026-10-05", "--until", "2026-10-01"); err == nil {
		t.Error("want error when --since is after --until")
	}
	if _, err := run("transitions", "--since", "last week"); err == nil {
		t.Error("want error for an unparseable --since")
	}
}

func TestParseTimeFlag(t *testing.T) {
	now := time.Date(2026, 10, 7, 15, 30, 0, 0, time.Local)
	cases := []struct {
		in       string
		endOfDay bool
		want     time.Time
	}{
		{"2026-09-30", false, time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local)},
		{"2026-09-30", true, time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local).Add(-time.Millisecond)},
		{"2026-09-30T08:00:00Z", true, time.Date(2026, 9, 30, 8, 0, 0, 0, time.UTC)},
		{"7d", false, now.Add(-7 * 24 * time.Hour)},
		{"2w", false, now.Add(-14 * 24 * time.Hour)},
		{"36h", false, now.Add(-36 * time.Hour)},
	}
	for _, c := range cases {
		got, err := parseTimeFlag(c.in, now, c.endOfDay)
		if err != nil {
			t.Errorf("parseTimeFlag(%q): %v", c.in, err)
			continue
		}
		if !got.Equal(c.want) {
			t.Errorf("parseTimeFlag(%q, %v) = %v, want %v", c.in, c.endOfDay, got, c.want)
		}
	}

	for _, bad := range []string{"", "0d", "-3d", "xd", "yesterday", "2026-13-01"} {
		if _, err := parseTimeFlag(bad, now, false); err == nil {
			t.Errorf("parseTimeFlag(%q) = nil error, want error", bad)
		}
	}
}
