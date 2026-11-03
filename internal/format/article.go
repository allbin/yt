package format

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/allbin/yt/internal/youtrack"
)

func Article(w io.Writer, a *youtrack.Article) error {
	ew := &errWriter{w: w}

	ew.printf("%s  %s\n", StyleID.Render(a.IDReadable), StyleBold.Render(a.Summary))
	ew.println()
	if a.Project != nil {
		ew.printf("  %s %s\n", StyleLabel.Render("Project"), a.Project.ShortName)
	}
	if a.Parent != nil {
		ew.printf("  %s %s\n", StyleLabel.Render("Parent"), articleRef(*a.Parent))
	}
	refList(ew, "Children", a.Children)
	if len(a.Tags) > 0 {
		names := make([]string, len(a.Tags))
		for i, t := range a.Tags {
			names[i] = t.Name
		}
		ew.printf("  %s %s\n", StyleLabel.Render("Tags"), StyleDim.Render(strings.Join(names, ", ")))
	}
	if v := a.Visibility; v != nil && v.Limited {
		ew.printf("  %s %s\n", StyleLabel.Render("Visible to"), strings.Join(append(slices.Clone(v.Users), v.Groups...), ", "))
	}
	if a.Updated > 0 {
		updated := time.UnixMilli(a.Updated).Format("2006-01-02 15:04")
		if a.UpdatedBy != nil {
			updated += StyleDim.Render(" by " + userName(a.UpdatedBy))
		}
		ew.printf("  %s %s\n", StyleLabel.Render("Updated"), updated)
	}
	for i, m := range a.MentionedIssues {
		label := StyleLabel.Render("Mentions")
		if i > 0 {
			label = StyleLabel.Render("")
		}
		ew.printf("  %s %s %s\n", label, StyleBold.Render(m.IDReadable), m.Summary)
	}

	if strings.TrimSpace(a.Content) != "" {
		ew.println()
		ew.println(StyleRule.Render("────────────────────────────────────"))
		ew.println(strings.Trim(RenderMarkdown(a.Content, 80), "\n"))
	}

	if len(a.Attachments) > 0 {
		ew.println()
		parts := make([]string, len(a.Attachments))
		for i, at := range a.Attachments {
			parts[i] = fmt.Sprintf("%s (%s)", at.Name, FormatSize(at.Size))
		}
		ew.printf("  %s %s\n", StyleLabel.Render("Attachments"), strings.Join(parts, ", "))
	}
	return ew.err
}

// ArticleList prints articles as a tree: each under its parent when the
// parent is in the list, at the top level otherwise, siblings in the order
// the knowledge base shows them.
func ArticleList(w io.Writer, articles []youtrack.Article) error {
	if len(articles) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No articles found."))
		return err
	}

	present := make(map[string]bool, len(articles))
	for _, a := range articles {
		present[a.IDReadable] = true
	}
	children := map[string][]youtrack.Article{}
	var roots []youtrack.Article
	for _, a := range articles {
		if a.Parent != nil && present[a.Parent.IDReadable] {
			children[a.Parent.IDReadable] = append(children[a.Parent.IDReadable], a)
			continue
		}
		roots = append(roots, a)
	}

	ew := &errWriter{w: w}
	var walk func(level []youtrack.Article, depth int)
	walk = func(level []youtrack.Article, depth int) {
		slices.SortStableFunc(level, compareArticles)
		for _, a := range level {
			ew.printf("%s%s  %s\n", strings.Repeat("  ", depth), StyleID.Render(a.IDReadable), a.Summary)
			walk(children[a.IDReadable], depth+1)
		}
	}
	walk(roots, 0)
	return ew.err
}

// compareArticles orders by project, then the article's place among its
// siblings.
func compareArticles(a, b youtrack.Article) int {
	var pa, pb string
	if a.Project != nil {
		pa = a.Project.ShortName
	}
	if b.Project != nil {
		pb = b.Project.ShortName
	}
	return cmp.Or(cmp.Compare(pa, pb), cmp.Compare(a.Ordinal, b.Ordinal))
}

func articleRef(r youtrack.ArticleRef) string {
	return StyleBold.Render(r.IDReadable) + " " + r.Summary
}

func refList(ew *errWriter, name string, refs []youtrack.ArticleRef) {
	for i, r := range refs {
		label := StyleLabel.Render(name)
		if i > 0 {
			label = StyleLabel.Render("")
		}
		ew.printf("  %s %s\n", label, articleRef(r))
	}
}
