package youtrack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// Comments, attachments and tags of an article ID go to the article
// endpoints, those of an issue ID to the issue endpoints.
func TestArticleIDRoutesToArticles(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/comments") {
			_, _ = io.WriteString(w, `{}`)
			return
		}
		_, _ = io.WriteString(w, `[]`)
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "t")

	if _, err := c.ListComments("SV-A-3"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddComment("SV-3", "x"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UploadAttachments("SV-A-3", []UploadFile{{Name: "a.txt", Content: strings.NewReader("a")}}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.UpdateComment("SV-3", "4-1", "x"); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteComment("SV-A-3", "215-1"); err != nil {
		t.Fatal(err)
	}
	if err := c.AddTag("SV-A-3", "6-1"); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET /api/articles/SV-A-3/comments", "POST /api/issues/SV-3/comments", "POST /api/articles/SV-A-3/attachments",
		"POST /api/issues/SV-3/comments/4-1", "DELETE /api/articles/SV-A-3/comments/215-1", "POST /api/articles/SV-A-3/tags",
	}
	if strings.Join(paths, "\n") != strings.Join(want, "\n") {
		t.Errorf("paths = %q, want %q", paths, want)
	}
}

// Every page of search hits is read, and a hit whose mentioned issues do not
// include the exact ID is dropped.
func TestArticlesMentioningPagesAndFilters(t *testing.T) {
	orig := mentionPageSize
	mentionPageSize = 2
	t.Cleanup(func() { mentionPageSize = orig })

	pages := [][]Article{
		{{IDReadable: "SV-A-1", MentionedIssues: []LinkedIssue{{IDReadable: "SV-123"}}}, {IDReadable: "SV-A-2"}},
		{{IDReadable: "SV-A-3", MentionedIssues: []LinkedIssue{{IDReadable: "SV-12"}}}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query().Get("query"); q != "mentions: SV-12" {
			t.Errorf("query = %q", q)
		}
		skip, _ := strconv.Atoi(r.URL.Query().Get("$skip"))
		_ = json.NewEncoder(w).Encode(pages[skip/2])
	}))
	defer srv.Close()

	refs, err := NewClient(srv.URL, "t").ArticlesMentioning("SV-12")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].IDReadable != "SV-A-3" {
		t.Errorf("refs = %+v", refs)
	}
}

func TestCreateArticleBody(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		_, _ = io.WriteString(w, `{"idReadable":"SV-A-9"}`)
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL, "t").CreateArticle(NewArticle{Project: "SV", Summary: "S", ParentID: "109-3", TagIDs: []string{"6-1"}}); err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(body)
	if want := `{"parentArticle":{"id":"109-3"},"project":{"shortName":"SV"},"summary":"S","tags":[{"id":"6-1"}]}`; string(got) != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

// A move to the top level sends a null parent; a move under another article
// sends its id; no move leaves the parent out.
func TestUpdateArticleMove(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(b))
	}))
	defer srv.Close()
	c := NewClient(srv.URL, "t")

	title := "T"
	for _, u := range []ArticleUpdate{{Move: true}, {Move: true, ParentID: "109-3"}, {Summary: &title}} {
		if err := c.UpdateArticle("SV-A-1", u); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{`{"parentArticle":null}`, `{"parentArticle":{"id":"109-3"}}`, `{"summary":"T"}`}
	if strings.Join(bodies, "|") != strings.Join(want, "|") {
		t.Errorf("bodies = %q, want %q", bodies, want)
	}
}

func TestVisibilityFromAPI(t *testing.T) {
	var a Article
	err := json.Unmarshal([]byte(`{"visibility":{"$type":"LimitedVisibility",
		"permittedUsers":[{"login":"h","fullName":"Hampus"}],"permittedGroups":[{"name":"Devs"}]}}`), &a)
	if err != nil {
		t.Fatal(err)
	}
	if v := a.Visibility; v == nil || !v.Limited || len(v.Users) != 1 || v.Users[0] != "Hampus" || len(v.Groups) != 1 {
		t.Errorf("visibility = %+v", a.Visibility)
	}
}
