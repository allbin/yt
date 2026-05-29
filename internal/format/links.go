package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
)

// Links formats an issue's links grouped by relation phrase.
func Links(w io.Writer, links []youtrack.IssueLink) error {
	if len(links) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No links."))
		return err
	}

	ew := &errWriter{w: w}
	for i, l := range links {
		if i > 0 {
			ew.println()
		}
		ew.printf("%s\n", StyleBold.Render(l.Phrase()))
		for _, iss := range l.Issues {
			ew.printf("  %s  %s\n", StyleID.Render(iss.IDReadable), StyleDim.Render(iss.Summary))
		}
	}
	return ew.err
}

// LinkTypes formats the instance's available link types and their phrases.
func LinkTypes(w io.Writer, types []youtrack.LinkType) error {
	if len(types) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No link types."))
		return err
	}

	ew := &errWriter{w: w}
	for i, t := range types {
		if i > 0 {
			ew.println()
		}
		ew.printf("%s\n", StyleBold.Render(t.Name))
		ew.printf("  %s %s\n", StyleLabel.Render("outward"), t.SourceToTarget)
		if t.Directed && t.TargetToSource != "" {
			ew.printf("  %s %s\n", StyleLabel.Render("inward"), t.TargetToSource)
		} else {
			ew.printf("  %s\n", StyleDim.Render("(symmetric)"))
		}
	}
	return ew.err
}

// linkSummary renders an issue's links as compact one-line-per-relation entries
// for inclusion in the issue detail view. Returns "" if there are no links.
func linkSummary(links []youtrack.IssueLink) string {
	if len(links) == 0 {
		return ""
	}
	const pad = "               " // 2 + label width 12 + 1, aligns continuation lines
	var b strings.Builder
	for i, l := range links {
		ids := make([]string, len(l.Issues))
		for j, iss := range l.Issues {
			ids[j] = iss.IDReadable
		}
		label := pad
		if i == 0 {
			label = "  " + StyleLabel.Render("Links") + " "
		}
		fmt.Fprintf(&b, "%s%s %s %s\n", label, l.Phrase(), StyleDim.Render("→"), strings.Join(ids, ", "))
	}
	return b.String()
}
