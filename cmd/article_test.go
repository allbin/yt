package cmd

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestArticleCreateUnderParent(t *testing.T) {
	mock := &mockAPI{articles: map[string]*youtrack.Article{
		"SV-A-3": {ID: "109-3", IDReadable: "SV-A-3", Project: &youtrack.Project{ShortName: "SV"}},
	}}
	run := setupTest(t, mock)

	if _, err := run("article", "create", "--parent", "SV-A-3", "-s", "Planning"); err != nil {
		t.Fatal(err)
	}
	got := mock.createdArticle
	if got == nil || got.Project != "SV" || got.Summary != "Planning" || got.ParentID != "109-3" {
		t.Errorf("created %+v, want project SV under 109-3", got)
	}
}

func TestArticleUpdateAppend(t *testing.T) {
	mock := &mockAPI{articles: map[string]*youtrack.Article{
		"SV-A-3": {IDReadable: "SV-A-3", Content: "# Notes\n- one\n\n"},
	}}
	run := setupTest(t, mock)

	if _, err := run("article", "update", "SV-A-3", "--append", "- two\n"); err != nil {
		t.Fatal(err)
	}
	if mock.updatedArticle == nil || mock.updatedArticle.Content == nil {
		t.Fatalf("content not written: %+v", mock.updatedArticle)
	}
	if got, want := *mock.updatedArticle.Content, "# Notes\n- one\n\n- two"; got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestArticleDelete(t *testing.T) {
	tests := []struct {
		name    string
		article youtrack.Article
		args    []string
		wantErr string
		deleted bool
	}{
		{"needs --yes", youtrack.Article{IDReadable: "SV-A-3"}, nil, "--yes", false},
		{"refuses a parent", youtrack.Article{IDReadable: "SV-A-3", HasChildren: true}, []string{"--yes"}, "sub-articles", false},
		{"deletes", youtrack.Article{IDReadable: "SV-A-3"}, []string{"--yes"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAPI{articles: map[string]*youtrack.Article{"SV-A-3": &tt.article}}
			run := setupTest(t, mock)

			_, err := run(append([]string{"article", "delete", "SV-A-3"}, tt.args...)...)
			if tt.wantErr == "" && err != nil {
				t.Fatal(err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Errorf("err = %v, want it to mention %q", err, tt.wantErr)
			}
			if got := slices.Equal(mock.deletedArticles, []string{"SV-A-3"}); got != tt.deleted {
				t.Errorf("deleted = %v, want %v", mock.deletedArticles, tt.deleted)
			}
		})
	}
}

func TestArticleRefusesIssueID(t *testing.T) {
	run := setupTest(t, &mockAPI{})
	if _, err := run("article", "comment", "SV-12", "-m", "x"); err == nil || !strings.Contains(err.Error(), "not an article ID") {
		t.Errorf("err = %v", err)
	}
}

func TestIssueShowsMentioningArticles(t *testing.T) {
	mock := &mockAPI{
		issue:      &youtrack.Issue{IDReadable: "SV-12"},
		mentioning: []youtrack.ArticleRef{{IDReadable: "SV-A-3", Summary: "Planning"}},
	}
	run := setupTest(t, mock)

	out, err := run("issue", "sv-12", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got struct{ Articles []youtrack.ArticleRef }
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Articles) != 1 || got.Articles[0].IDReadable != "SV-A-3" {
		t.Errorf("articles = %+v", got.Articles)
	}
	if mock.mentioningID != "SV-12" {
		t.Errorf("searched mentions of %q, want the canonical SV-12", mock.mentioningID)
	}
}

func TestArticleUpdateMove(t *testing.T) {
	mock := &mockAPI{articles: map[string]*youtrack.Article{
		"SV-A-1": {ID: "109-1", IDReadable: "SV-A-1"},
		"SV-A-2": {ID: "109-2", IDReadable: "SV-A-2"},
	}}
	run := setupTest(t, mock)

	if _, err := run("article", "update", "SV-A-1", "--parent", "SV-A-2"); err != nil {
		t.Fatal(err)
	}
	if u := mock.updatedArticle; u == nil || !u.Move || u.ParentID != "109-2" {
		t.Errorf("update = %+v, want a move under 109-2", u)
	}
	if _, err := run("article", "update", "SV-A-1", "--parent", "none"); err != nil {
		t.Fatal(err)
	}
	if u := mock.updatedArticle; !u.Move || u.ParentID != "" {
		t.Errorf("update = %+v, want a move to the top level", u)
	}
}

func TestCommentDelete(t *testing.T) {
	mock := &mockAPI{comments: []youtrack.Comment{{ID: "4-1", Text: "typo here\nmore"}}}
	run := setupTest(t, mock)

	if _, err := run("issue", "comment", "delete", "SV-1", "4-1"); err == nil || !strings.Contains(err.Error(), `"typo here…"`) {
		t.Errorf("without --yes: err = %v, want a refusal quoting the comment", err)
	}
	if _, err := run("issue", "comment", "delete", "SV-1", "4-2", "--yes"); err == nil || !strings.Contains(err.Error(), "no comment 4-2") {
		t.Errorf("unknown id: err = %v", err)
	}
	if len(mock.deletedComments) != 0 {
		t.Fatalf("deleted %v before confirmation", mock.deletedComments)
	}
	if _, err := run("issue", "comment", "delete", "SV-1", "4-1", "--yes"); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(mock.deletedComments, []string{"4-1"}) {
		t.Errorf("deleted = %v", mock.deletedComments)
	}
}

func TestCommentEdit(t *testing.T) {
	mock := &mockAPI{}
	run := setupTest(t, mock)

	if _, err := run("article", "comment", "edit", "sv-a-1", "215-1", "-m", "fixed"); err != nil {
		t.Fatal(err)
	}
	if mock.editedComment != "215-1|fixed" {
		t.Errorf("edited = %q", mock.editedComment)
	}
	// The plain comment command still posts a new comment.
	if _, err := run("issue", "comment", "SV-1", "-m", "new"); err != nil {
		t.Fatal(err)
	}
	if mock.addedComment != "new" {
		t.Errorf("added = %q", mock.addedComment)
	}
}

func TestHistoryKindFilter(t *testing.T) {
	mock := &mockAPI{history: []youtrack.HistoryEntry{{Kind: "field"}, {Kind: "comment"}, {Kind: "tag"}}}
	run := setupTest(t, mock)

	out, err := run("issue", "history", "SV-1", "--kind", "Field,tag", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var got []youtrack.HistoryEntry
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Kind != "field" || got[1].Kind != "tag" {
		t.Errorf("entries = %+v", got)
	}
}
