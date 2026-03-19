package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/allbin/yt/internal/youtrack"
)

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) printf(format string, a ...any) {
	if ew.err == nil {
		_, ew.err = fmt.Fprintf(ew.w, format, a...)
	}
}

func (ew *errWriter) println(a ...any) {
	if ew.err == nil {
		_, ew.err = fmt.Fprintln(ew.w, a...)
	}
}

func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func Issue(w io.Writer, issue *youtrack.Issue) error {
	ew := &errWriter{w: w}
	ew.printf("# %s: %s\n\n", issue.IDReadable, issue.Summary)

	state := issue.Field("State")
	assignee := issue.Field("Assignee")
	priority := issue.Field("Priority")
	typ := issue.Field("Type")
	subsystem := issue.Field("Subsystem")
	tags := issue.TagNames()

	var meta []string
	if state != "" {
		meta = append(meta, "**State:** "+state)
	}
	if assignee != "" {
		meta = append(meta, "**Assignee:** "+assignee)
	}
	if priority != "" {
		meta = append(meta, "**Priority:** "+priority)
	}
	if typ != "" {
		meta = append(meta, "**Type:** "+typ)
	}
	if len(meta) > 0 {
		ew.println(strings.Join(meta, "  "))
		ew.println()
	}

	if subsystem != "" {
		ew.printf("**Subsystem:** %s\n", subsystem)
	}
	if tags != "" {
		ew.printf("**Tags:** %s\n", tags)
	}
	if subsystem != "" || tags != "" {
		ew.println()
	}

	desc := issue.Desc()
	if desc != "" {
		ew.println("## Description")
		ew.println()
		ew.println(desc)
	}

	return ew.err
}

func IssueList(w io.Writer, issues []youtrack.Issue) error {
	if len(issues) == 0 {
		_, err := fmt.Fprintln(w, "No issues found.")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	ew := &errWriter{w: tw}
	ew.println("ID\tSTATE\tPRIORITY\tASSIGNEE\tSUMMARY")
	for _, issue := range issues {
		ew.printf("%s\t%s\t%s\t%s\t%s\n",
			issue.IDReadable,
			issue.Field("State"),
			issue.Field("Priority"),
			issue.Field("Assignee"),
			issue.Summary,
		)
	}
	if ew.err != nil {
		return ew.err
	}
	return tw.Flush()
}
