package youtrack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIssue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/issues/TEST-1" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode(Issue{
			IDReadable: "TEST-1",
			Summary:    "Test issue",
			Tags:       []Tag{{Name: "backend"}},
			CustomFields: []CustomField{
				{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
			},
		}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-token")

	issue, err := client.GetIssue("TEST-1")
	if err != nil {
		t.Fatal(err)
	}
	if issue.IDReadable != "TEST-1" {
		t.Errorf("IDReadable = %q, want TEST-1", issue.IDReadable)
	}
	if issue.Summary != "Test issue" {
		t.Errorf("Summary = %q, want Test issue", issue.Summary)
	}
	if issue.Field("State") != "Open" {
		t.Errorf("State = %q, want Open", issue.Field("State"))
	}
}

func TestGetIssueNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"error":"Not Found"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	_, err := client.GetIssue("NOPE-1")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestListIssues(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("query") != "project: PROJ" {
			t.Errorf("query = %q, want 'project: PROJ'", q.Get("query"))
		}
		if q.Get("$top") != "5" {
			t.Errorf("$top = %q, want 5", q.Get("$top"))
		}
		if err := json.NewEncoder(w).Encode([]Issue{
			{IDReadable: "PROJ-1", Summary: "First"},
			{IDReadable: "PROJ-2", Summary: "Second"},
		}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	issues, err := client.ListIssues("project: PROJ", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 2 {
		t.Fatalf("got %d issues, want 2", len(issues))
	}
	if issues[0].IDReadable != "PROJ-1" {
		t.Errorf("issues[0].IDReadable = %q, want PROJ-1", issues[0].IDReadable)
	}
}

func TestListIssuesNoQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "" {
			t.Error("expected no query parameter")
		}
		if err := json.NewEncoder(w).Encode([]Issue{}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	issues, err := client.ListIssues("", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Errorf("got %d issues, want 0", len(issues))
	}
}
