package youtrack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Comments and attachments of an article ID go to the article endpoints.
func TestArticleIDRoutesToArticles(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.Method+" "+r.URL.Path)
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/comments") {
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
	want := []string{"GET /api/articles/SV-A-3/comments", "POST /api/issues/SV-3/comments", "POST /api/articles/SV-A-3/attachments"}
	if strings.Join(paths, "\n") != strings.Join(want, "\n") {
		t.Errorf("paths = %q, want %q", paths, want)
	}
}

// A search hit whose mentioned issues do not include the exact ID is dropped.
func TestArticlesMentioningFiltersExact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if q := r.URL.Query().Get("query"); q != "mentions: SV-12" {
			t.Errorf("query = %q", q)
		}
		_ = json.NewEncoder(w).Encode([]Article{
			{IDReadable: "SV-A-1", MentionedIssues: []LinkedIssue{{IDReadable: "SV-12"}}},
			{IDReadable: "SV-A-2", MentionedIssues: []LinkedIssue{{IDReadable: "SV-123"}}},
		})
	}))
	defer srv.Close()

	refs, err := NewClient(srv.URL, "t").ArticlesMentioning("SV-12")
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) != 1 || refs[0].IDReadable != "SV-A-1" {
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

	if _, err := NewClient(srv.URL, "t").CreateArticle(NewArticle{Project: "SV", Summary: "S", ParentID: "109-3"}); err != nil {
		t.Fatal(err)
	}
	got, _ := json.Marshal(body)
	if want := `{"parentArticle":{"id":"109-3"},"project":{"shortName":"SV"},"summary":"S"}`; string(got) != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}
