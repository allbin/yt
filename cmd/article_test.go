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
	want := youtrack.NewArticle{Project: "SV", Summary: "Planning", ParentID: "109-3"}
	if mock.createdArticle == nil || *mock.createdArticle != want {
		t.Errorf("created %+v, want %+v", mock.createdArticle, want)
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
	if got, want := mock.updatedArticle["content"], "# Notes\n- one\n\n- two"; got != want {
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
