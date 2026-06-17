package youtrack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListSprintIssues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/agiles/A1/sprints/S1/issues" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("$top") != "500" {
			t.Errorf("$top = %q, want 500", r.URL.Query().Get("$top"))
		}
		if err := json.NewEncoder(w).Encode([]map[string]string{
			{"idReadable": "AX-1"},
			{"idReadable": "AX-2"},
		}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	ids, err := client.ListSprintIssues("A1", "S1")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != "AX-1" || ids[1] != "AX-2" {
		t.Errorf("ids = %v, want [AX-1 AX-2]", ids)
	}
}

func TestAddIssueToSprint(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	if err := client.AddIssueToSprint("A1", "S1", "AX-812"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/agiles/A1/sprints/S1/issues" {
		t.Errorf("path = %q", gotPath)
	}
	if !strings.Contains(gotBody, `"idReadable":"AX-812"`) {
		t.Errorf("body = %q, want idReadable AX-812", gotBody)
	}
}

func TestRemoveIssueFromSprint(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	if err := client.RemoveIssueFromSprint("A1", "S1", "AX-812"); err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
	if gotPath != "/api/agiles/A1/sprints/S1/issues/AX-812" {
		t.Errorf("path = %q", gotPath)
	}
}

func TestIssueBoards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/agiles":
			// Two boards: AllTix (project AX) and Other (project OT).
			if err := json.NewEncoder(w).Encode([]Agile{
				{
					ID:            "A1",
					Name:          "AllTix",
					Projects:      []Project{{ShortName: "AX"}},
					CurrentSprint: &Sprint{ID: "S2", Name: "2025-06"},
					Sprints:       []Sprint{{ID: "S1", Name: "2025-05"}, {ID: "S2", Name: "2025-06"}},
				},
				{
					ID:       "A2",
					Name:     "Other",
					Projects: []Project{{ShortName: "OT"}},
					Sprints:  []Sprint{{ID: "S9", Name: "x"}},
				},
			}); err != nil {
				t.Error(err)
			}
		case strings.HasPrefix(r.URL.Path, "/api/agiles/A1/sprints/S1/"):
			writeIDs(t, w, "AX-1")
		case strings.HasPrefix(r.URL.Path, "/api/agiles/A1/sprints/S2/"):
			writeIDs(t, w, "AX-812", "AX-3")
		default:
			t.Errorf("unexpected path %q (Other board should be filtered out by project)", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	boards, err := client.IssueBoards("AX-812")
	if err != nil {
		t.Fatal(err)
	}
	if len(boards) != 1 {
		t.Fatalf("memberships = %v, want 1", boards)
	}
	if boards[0].Board != "AllTix" || boards[0].Sprint != "2025-06" {
		t.Errorf("membership = %+v, want AllTix/2025-06", boards[0])
	}
}

func writeIDs(t *testing.T, w http.ResponseWriter, ids ...string) {
	t.Helper()
	refs := make([]map[string]string, len(ids))
	for i, id := range ids {
		refs[i] = map[string]string{"idReadable": id}
	}
	if err := json.NewEncoder(w).Encode(refs); err != nil {
		t.Error(err)
	}
}

func TestProjectPrefix(t *testing.T) {
	tests := map[string]string{
		"AX-812":   "AX",
		"PROJ-1":   "PROJ",
		"FOO-BAR-9": "FOO-BAR",
		"nodelim":  "",
		"-5":       "",
	}
	for in, want := range tests {
		if got := projectPrefix(in); got != want {
			t.Errorf("projectPrefix(%q) = %q, want %q", in, got, want)
		}
	}
}
