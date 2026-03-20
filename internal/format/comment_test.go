package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestCommentList(t *testing.T) {
	comments := []youtrack.Comment{
		{
			ID:      "4-1",
			Text:    "First comment",
			Author:  &youtrack.User{Login: "alice", FullName: "Alice"},
			Created: 1700000000000,
		},
		{
			ID:      "4-2",
			Text:    "Second comment",
			Author:  &youtrack.User{Login: "bob", FullName: "Bob"},
			Created: 1700000060000,
		},
	}

	var buf bytes.Buffer
	if err := CommentList(&buf, comments); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	checks := []string{"Alice", "Bob", "First comment", "Second comment", "2023-11-14"}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestCommentListEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := CommentList(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No comments") {
		t.Error("expected 'No comments' for empty list")
	}
}

func TestCommentListNilAuthor(t *testing.T) {
	comments := []youtrack.Comment{
		{
			ID:      "4-1",
			Text:    "Orphan comment",
			Author:  nil,
			Created: 1700000000000,
		},
	}

	var buf bytes.Buffer
	if err := CommentList(&buf, comments); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Unknown") {
		t.Error("expected 'Unknown' for nil author")
	}
}
