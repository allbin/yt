package youtrack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// stateChange is a CustomFieldActivityItem as the activitiesPage API returns
// it for a State change.
func stateChange(id, issue string, ts int64, from, to string) map[string]any {
	removed := []map[string]string{}
	if from != "" {
		removed = append(removed, map[string]string{"name": from, "$type": "StateBundleElement"})
	}
	return map[string]any{
		"id":        id,
		"timestamp": ts,
		"author":    map[string]string{"login": "alice", "fullName": "Alice"},
		"target":    map[string]string{"idReadable": issue, "summary": "Summary of " + issue},
		"field":     map[string]string{"name": "State"},
		"added":     []map[string]string{{"name": to, "$type": "StateBundleElement"}},
		"removed":   removed,
		"$type":     "CustomFieldActivityItem",
	}
}

func TestListFieldActivitiesPaginates(t *testing.T) {
	pages := map[string]map[string]any{
		"": {
			"activities":  []any{stateChange("a1", "AX-1", 1000, "Planned", "In Progress")},
			"afterCursor": "c1",
			"hasAfter":    true,
		},
		"c1": {
			"activities":  []any{stateChange("a2", "AX-2", 2000, "In Progress", "In Review")},
			"afterCursor": "c2",
			"hasAfter":    true,
		},
		"c2": {
			"activities":  []any{stateChange("a3", "AX-3", 3000, "In Review", "Done")},
			"afterCursor": "c3",
			"hasAfter":    false,
		},
	}

	var cursors []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/activitiesPage" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		want := map[string]string{
			"categories": "CustomFieldCategory",
			"author":     "me",
			"start":      "1700000000000",
			"end":        "1700600000000",
			"issueQuery": "project: AX",
			"$top":       "100",
		}
		for k, v := range want {
			if got := q.Get(k); got != v {
				t.Errorf("%s = %q, want %q", k, got, v)
			}
		}
		if !strings.Contains(q.Get("fields"), "afterCursor") || !strings.Contains(q.Get("fields"), "hasAfter") {
			t.Errorf("fields = %q, want cursor fields", q.Get("fields"))
		}

		cursor := q.Get("cursor")
		cursors = append(cursors, cursor)
		page, ok := pages[cursor]
		if !ok {
			t.Errorf("unexpected cursor %q", cursor)
			http.NotFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode(page); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	acts, err := client.ListFieldActivities(FieldActivityFilter{
		Author:     "me",
		Start:      time.UnixMilli(1700000000000),
		End:        time.UnixMilli(1700600000000),
		IssueQuery: "project: AX",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(cursors, ",") != ",c1,c2" {
		t.Errorf("cursors = %q, want first page then c1, c2", cursors)
	}
	if len(acts) != 3 || acts[0].ID != "a1" || acts[2].ID != "a3" {
		t.Fatalf("activities = %+v, want a1..a3 in order", acts)
	}
}

func TestListFieldActivitiesOmitsUnsetFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, k := range []string{"author", "start", "end", "issueQuery", "cursor"} {
			if r.URL.Query().Has(k) {
				t.Errorf("%s sent for an unset filter", k)
			}
		}
		if _, err := w.Write([]byte(`{"activities":[],"hasAfter":false}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	acts, err := NewClient(srv.URL, "token").ListFieldActivities(FieldActivityFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(acts) != 0 {
		t.Errorf("got %d activities, want 0", len(acts))
	}
}

func TestListFieldActivitiesStuckCursor(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls > 3 {
			t.Error("kept paging on a cursor that does not advance")
			http.Error(w, "stop", http.StatusInternalServerError)
			return
		}
		if _, err := w.Write([]byte(`{"activities":[],"afterCursor":"same","hasAfter":true}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, "token").ListFieldActivities(FieldActivityFilter{})
	if err == nil || !strings.Contains(err.Error(), "did not advance") {
		t.Errorf("err = %v, want cursor did not advance", err)
	}
}

func decodeActivities(t *testing.T, items ...map[string]any) []FieldActivity {
	t.Helper()
	data, err := json.Marshal(items)
	if err != nil {
		t.Fatal(err)
	}
	var acts []FieldActivity
	if err := json.Unmarshal(data, &acts); err != nil {
		t.Fatal(err)
	}
	return acts
}

func TestTransitions(t *testing.T) {
	assignee := map[string]any{
		"id":        "x1",
		"timestamp": 1500,
		"author":    map[string]string{"login": "alice"},
		"target":    map[string]string{"idReadable": "AX-1"},
		"field":     map[string]string{"name": "Assignee"},
		"added":     []map[string]string{{"login": "alice", "name": "Alice"}},
		"removed":   []any{},
	}
	// Scalar added/removed, as simple (integer) fields report them, must not
	// break decoding of the rest of the page.
	points := map[string]any{
		"id":        "x2",
		"timestamp": 1600,
		"target":    map[string]string{"idReadable": "AX-1"},
		"field":     map[string]string{"name": "Story points"},
		"added":     3,
		"removed":   nil,
	}
	noTarget := stateChange("x3", "AX-9", 1700, "Open", "Done")
	delete(noTarget, "target")

	acts := decodeActivities(t,
		stateChange("a1", "AX-1", 1000, "", "Submitted"),
		assignee,
		points,
		stateChange("a2", "AX-1", 2000, "Submitted", "In Progress"),
		noTarget,
	)

	got := Transitions(acts, "state")
	if len(got) != 2 {
		t.Fatalf("got %d transitions, want 2: %+v", len(got), got)
	}

	first := got[0]
	if first.Issue != "AX-1" || first.Summary != "Summary of AX-1" || first.Field != "State" {
		t.Errorf("first = %+v", first)
	}
	if first.From != "" || first.To != "Submitted" {
		t.Errorf("first from/to = %q -> %q, want \"\" -> Submitted", first.From, first.To)
	}
	if first.Author == nil || first.Author.Login != "alice" {
		t.Errorf("first author = %+v, want alice", first.Author)
	}
	if first.Timestamp != 1000 {
		t.Errorf("first timestamp = %d, want 1000", first.Timestamp)
	}
	if ts, err := time.Parse(time.RFC3339, first.Time); err != nil || ts.Unix() != 1 {
		t.Errorf("first time = %q, want RFC 3339 of 1000ms (err %v)", first.Time, err)
	}
	if got[1].From != "Submitted" || got[1].To != "In Progress" {
		t.Errorf("second from/to = %q -> %q", got[1].From, got[1].To)
	}

	if users := Transitions(acts, "Assignee"); len(users) != 1 || users[0].To != "Alice" {
		t.Errorf("assignee transitions = %+v, want one to Alice", users)
	}
}

func TestTransitionsEmptyIsNotNil(t *testing.T) {
	// JSON output must be [] rather than null when nothing matched.
	if got := Transitions(nil, "State"); got == nil {
		t.Error("Transitions(nil) = nil, want empty slice")
	}
}

func TestAgileColumnOf(t *testing.T) {
	board := &Agile{ColumnSettings: &AgileColumnSettings{
		Field: &struct {
			Name string `json:"name"`
		}{Name: "Stage"},
		Columns: []AgileColumn{
			{Presentation: "To do", FieldValues: []AgileColumnValue{{Name: "Open"}, {Name: "Submitted"}}},
			{Presentation: "Doing", FieldValues: []AgileColumnValue{{Name: "In Progress"}}},
		},
	}}

	if f := board.ColumnField(); f != "Stage" {
		t.Errorf("ColumnField = %q, want Stage", f)
	}
	cases := map[string]string{
		"Open":        "To do",
		"submitted":   "To do",
		"In Progress": "Doing",
		"Done":        "",
		"":            "",
	}
	for value, want := range cases {
		if got := board.ColumnOf(value); got != want {
			t.Errorf("ColumnOf(%q) = %q, want %q", value, got, want)
		}
	}

	var bare Agile
	if bare.ColumnField() != "" || bare.ColumnOf("Open") != "" {
		t.Error("board without column settings should map nothing")
	}
}
