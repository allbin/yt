package youtrack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateIssueFields(t *testing.T) {
	var got map[string]string
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	err := client.UpdateIssueFields("TEST-1", map[string]string{
		"summary":     "New title",
		"description": "New body",
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/issues/TEST-1" {
		t.Errorf("path = %q, want /api/issues/TEST-1", gotPath)
	}
	if got["summary"] != "New title" {
		t.Errorf("summary = %q, want New title", got["summary"])
	}
	if got["description"] != "New body" {
		t.Errorf("description = %q, want New body", got["description"])
	}
}

func TestUpdateIssueFieldsSummaryOnly(t *testing.T) {
	var got map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	err := client.UpdateIssueFields("TEST-1", map[string]string{"summary": "Only summary"})
	if err != nil {
		t.Fatal(err)
	}
	if got["summary"] != "Only summary" {
		t.Errorf("summary = %q, want Only summary", got["summary"])
	}
	if _, ok := got["description"]; ok {
		t.Error("description should not be present")
	}
}

func TestUpdateIssueFieldsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"error":"Not Found"}`)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	err := client.UpdateIssueFields("NOPE-1", map[string]string{"summary": "x"})
	if err == nil {
		t.Fatal("expected error for 404")
	}
}
