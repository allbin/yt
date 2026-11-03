package youtrack

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// An issue's history reads the issue's activity page with the issue
// categories, and each activity becomes one entry of its kind.
func TestIssueHistory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/api/issues/SV-1/activitiesPage" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.HasPrefix(q.Get("fields"), "afterCursor,hasAfter,activities(") {
			t.Errorf("fields = %q: the page wrapper is missing", q.Get("fields"))
		}
		if !strings.Contains(q.Get("categories"), "CustomFieldCategory") {
			t.Errorf("categories = %q", q.Get("categories"))
		}
		_, _ = io.WriteString(w, `{"hasAfter":false,"activities":[
			{"timestamp":1,"category":{"id":"CustomFieldCategory"},"field":{"name":"State"},
			 "removed":[{"name":"Open"}],"added":[{"name":"Done"}]},
			{"timestamp":2,"category":{"id":"CommentsCategory"},"field":{"name":"comments"},
			 "removed":[],"added":[{"id":"4-9","text":"hi"}]},
			{"timestamp":3,"category":{"id":"LinksCategory"},"field":{"name":"Relates to"},
			 "removed":[],"added":[{"idReadable":"SV-2","summary":"other"}]},
			{"timestamp":4,"category":{"id":"IssueResolvedCategory"},"field":{"name":"resolved date"},
			 "removed":123,"added":null}]}`)
	}))
	defer srv.Close()

	got, err := NewClient(srv.URL, "t").History("SV-1")
	if err != nil {
		t.Fatal(err)
	}
	type row struct{ Kind, Field, Removed, Added string }
	var rows []row
	for _, e := range got {
		rows = append(rows, row{e.Kind, e.Field, e.Removed, e.Added})
	}
	want := []row{
		{"field", "State", "Open", "Done"},
		{"comment", "4-9", "", "hi"},
		{"link", "Relates to", "", "SV-2"},
		{"resolved", "", "", "reopened"},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("entries = %+v\nwant      %+v", rows, want)
	}
}

func TestArticleHistoryUsesArticleCategories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/articles/SV-A-1/activitiesPage" || !strings.Contains(r.URL.Query().Get("categories"), "ArticleDescriptionCategory") {
			t.Errorf("request = %s %q", r.URL.Path, r.URL.Query().Get("categories"))
		}
		_, _ = io.WriteString(w, `{"hasAfter":false,"activities":[
			{"timestamp":1,"category":{"id":"ArticleDescriptionCategory"},"removed":"a","added":"b"}]}`)
	}))
	defer srv.Close()

	got, err := NewClient(srv.URL, "t").History("sv-a-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "content" || got[0].Added != "b" {
		t.Errorf("entries = %+v", got)
	}
}
