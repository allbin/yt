package youtrack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" || r.URL.Query().Get("query") != "discussed" {
			t.Errorf("unexpected request %s", r.URL)
		}
		_, _ = io.WriteString(w, `[{"id":"6-63","name":"To be discussed"}]`)
	}))
	defer srv.Close()

	tags, err := NewClient(srv.URL, "token").ListTags("discussed")
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0].ID != "6-63" || tags[0].Name != "To be discussed" {
		t.Errorf("tags = %+v", tags)
	}
}

func TestCreateTag(t *testing.T) {
	var body map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/tags" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		_, _ = io.WriteString(w, `{"id":"6-116","name":"`+body["name"]+`"}`)
	}))
	defer srv.Close()

	tag, err := NewClient(srv.URL, "token").CreateTag("needs review")
	if err != nil {
		t.Fatal(err)
	}
	if body["name"] != "needs review" {
		t.Errorf("sent name %q", body["name"])
	}
	if tag.ID != "6-116" || tag.Name != "needs review" {
		t.Errorf("tag = %+v", tag)
	}
}

func TestTagIssue(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
	}))
	defer srv.Close()

	if err := NewClient(srv.URL, "token").TagIssue("AX-1", "6-63"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "POST /api/issues/AX-1/tags" {
		t.Errorf("request = %q", gotPath)
	}
	if !strings.Contains(gotBody, `"id":"6-63"`) {
		t.Errorf("body = %q", gotBody)
	}
}

func TestUntagIssue(t *testing.T) {
	var gotPath string
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.Method + " " + r.URL.Path
		w.WriteHeader(status)
		_, _ = io.WriteString(w, `{"error":"Not Found"}`)
	}))
	defer srv.Close()
	client := NewClient(srv.URL, "token")

	if err := client.UntagIssue("AX-1", "6-63"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "DELETE /api/issues/AX-1/tags/6-63" {
		t.Errorf("request = %q", gotPath)
	}

	// A tag the issue does not carry answers 404 and is a no-op.
	status = http.StatusNotFound
	if err := client.UntagIssue("AX-1", "6-63"); err != nil {
		t.Errorf("404 should be a no-op, got %v", err)
	}

	status = http.StatusForbidden
	if err := client.UntagIssue("AX-1", "6-63"); err == nil {
		t.Error("403 should be an error")
	}
}

func TestCreateIssueSendsTagIDs(t *testing.T) {
	var body map[string]json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		_, _ = io.WriteString(w, `{"idReadable":"AX-9"}`)
	}))
	defer srv.Close()

	_, err := NewClient(srv.URL, "token").CreateIssue(NewIssue{Project: "AX", Summary: "S", TagIDs: []string{"6-63", "6-1"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(body["tags"]); got != `[{"id":"6-63"},{"id":"6-1"}]` {
		t.Errorf("tags = %s", got)
	}
	if _, ok := body["customFields"]; ok {
		t.Error("empty customFields should be omitted")
	}
}
