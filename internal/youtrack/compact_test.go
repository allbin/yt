package youtrack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestIssueCompact(t *testing.T) {
	updated := time.Date(2026, 10, 13, 9, 30, 0, 0, time.Local)
	issue := Issue{
		IDReadable: "AX-1050",
		Summary:    "Bind the phone",
		Updated:    updated.UnixMilli(),
		Tags:       []Tag{{Name: "To be discussed"}, {Name: "QA"}},
		CustomFields: []CustomField{
			{Name: "State", Value: json.RawMessage(`{"name":"Planned"}`)},
			{Name: "Type", Value: json.RawMessage(`{"name":"User Story"}`)},
			{Name: "Priority", Value: json.RawMessage(`{"name":"Critical"}`)},
			{Name: "Assignee", Value: json.RawMessage(`{"login":"jdoe","fullName":"Jane Doe"}`)},
			{Name: "Subsystem", Value: json.RawMessage(`[{"name":"Webticket"}]`)},
		},
	}

	got := issue.Compact()
	if got.ID != "AX-1050" || got.Summary != "Bind the phone" {
		t.Errorf("id/summary = %q/%q", got.ID, got.Summary)
	}
	if got.State != "Planned" || got.Type != "User Story" || got.Priority != "Critical" || got.Assignee != "Jane Doe" {
		t.Errorf("fields = %+v", got)
	}
	if want := []string{"To be discussed", "QA"}; !slices.Equal(got.Tags, want) {
		t.Errorf("tags = %q, want %q", got.Tags, want)
	}
	if ts, err := time.Parse(time.RFC3339, got.Updated); err != nil || !ts.Equal(updated) {
		t.Errorf("updated = %q, want %s", got.Updated, updated.Format(time.RFC3339))
	}
}

func TestIssueCompactEmpty(t *testing.T) {
	got := (&Issue{IDReadable: "AX-1"}).Compact()
	if got.Tags == nil {
		t.Error("tags should be an empty list, not null")
	}
	if got.Updated != "" {
		t.Errorf("updated = %q, want empty", got.Updated)
	}
}

func TestListIssuesCompactAsksForNarrowFields(t *testing.T) {
	var fields string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields = r.URL.Query().Get("fields")
		if err := json.NewEncoder(w).Encode([]Issue{{IDReadable: "AX-1", Summary: "One"}}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	got, err := NewClient(srv.URL, "token").ListIssuesCompact("project: AX", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "AX-1" {
		t.Errorf("got %+v", got)
	}
	if strings.Contains(fields, "description") || strings.Contains(fields, "attachments") {
		t.Errorf("compact list asked for %q", fields)
	}
}
