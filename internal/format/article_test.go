package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

// Children sit under a listed parent in sidebar order; an article whose
// parent is not listed is a root.
func TestArticleListTree(t *testing.T) {
	parent := &youtrack.ArticleRef{IDReadable: "SV-A-1"}
	var buf bytes.Buffer
	err := ArticleList(&buf, []youtrack.Article{
		{IDReadable: "SV-A-3", Parent: parent, Ordinal: 1},
		{IDReadable: "SV-A-1"},
		{IDReadable: "SV-A-2", Parent: parent, Ordinal: 0},
		{IDReadable: "SV-A-9", Parent: &youtrack.ArticleRef{IDReadable: "SV-A-8"}, Ordinal: 5},
	})
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, l := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
		lines = append(lines, strings.TrimRight(l, " "))
	}
	want := []string{"SV-A-1", "  SV-A-2", "  SV-A-3", "SV-A-9"}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Errorf("tree = %q, want %q", lines, want)
	}
}
