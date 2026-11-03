package format

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/allbin/yt/internal/youtrack"
)

// History prints one line per change, oldest first. Long text (description,
// content, comments) is cut to its first line unless full is set.
func History(w io.Writer, entries []youtrack.HistoryEntry, full bool) error {
	if len(entries) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No history."))
		return err
	}
	ew := &errWriter{w: w}
	for _, e := range entries {
		ts := time.UnixMilli(e.Timestamp).Format("2006-01-02 15:04")
		ew.printf("%s  %s  %s\n", StyleDim.Render(ts), StyleBold.Render(userName(e.Author)), historyChange(e, full))
	}
	return ew.err
}

func historyChange(e youtrack.HistoryEntry, full bool) string {
	text := func(s string) string {
		if full {
			return s
		}
		return FirstLine(s, 80)
	}
	switch e.Kind {
	case "created":
		return "created"
	case "resolved":
		return e.Added
	case "description", "content":
		if full {
			return fmt.Sprintf("%s changed\n--- before\n%s\n--- after\n%s", e.Kind, e.Removed, e.Added)
		}
		return fmt.Sprintf("%s changed %s", e.Kind, StyleDim.Render(fmt.Sprintf("(%d → %d chars)", len([]rune(e.Removed)), len([]rune(e.Added)))))
	case "comment":
		if e.Added != "" {
			return fmt.Sprintf("commented %s %s", StyleDim.Render(e.Field), text(e.Added))
		}
		return fmt.Sprintf("deleted comment %s %s", StyleDim.Render(e.Field), text(e.Removed))
	case "comment-edit":
		return fmt.Sprintf("edited comment %s %s", StyleDim.Render(e.Field), text(e.Added))
	}

	label := e.Kind
	if e.Field != "" {
		label += " " + e.Field
	}
	switch {
	case e.Removed != "" && e.Added != "":
		return fmt.Sprintf("%s: %s → %s", label, text(e.Removed), text(e.Added))
	case e.Added != "":
		return fmt.Sprintf("%s: + %s", label, text(e.Added))
	case e.Removed != "":
		return fmt.Sprintf("%s: − %s", label, text(e.Removed))
	}
	return label
}

// FirstLine cuts s to its first line and at most n runes, marking a cut with
// an ellipsis.
func FirstLine(s string, n int) string {
	s = strings.TrimSpace(s)
	line, _, cut := strings.Cut(s, "\n")
	if r := []rune(line); len(r) > n {
		line, cut = string(r[:n]), true
	}
	if cut {
		line += "…"
	}
	return line
}
