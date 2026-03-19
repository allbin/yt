package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

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

	ew.printf("%s  %s\n", StyleID.Render(issue.IDReadable), StyleBold.Render(issue.Summary))

	state := issue.Field("State")
	assignee := issue.Field("Assignee")
	priority := issue.Field("Priority")
	typ := issue.Field("Type")
	subsystem := issue.Field("Subsystem")
	tags := issue.TagNames()

	hasMeta := state != "" || assignee != "" || priority != "" || typ != "" || subsystem != "" || tags != ""
	if hasMeta {
		ew.println()
	}
	if state != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("State"), lipgloss.NewStyle().Foreground(StateColor(state)).Render(state))
	}
	if assignee != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Assignee"), assignee)
	}
	if priority != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Priority"), lipgloss.NewStyle().Foreground(PriorityColor(priority)).Render(priority))
	}
	if typ != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Type"), typ)
	}
	if subsystem != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Subsystem"), subsystem)
	}
	if tags != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Tags"), StyleDim.Render(tags))
	}

	desc := issue.Desc()
	if desc != "" {
		ew.println()
		ew.println(StyleRule.Render("────────────────────────────────────"))
		rendered := strings.Trim(RenderMarkdown(desc, 80), "\n")
		ew.println(rendered)
	}

	return ew.err
}

func IssueList(w io.Writer, issues []youtrack.Issue) error {
	if len(issues) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No issues found."))
		return err
	}
	_, err := fmt.Fprintln(w, issueTable(issues))
	return err
}

func issueTable(issues []youtrack.Issue) string {
	t := newTable("ID", "STATE", "PRIORITY", "ASSIGNEE", "SUMMARY").
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(ColorAccent)
			}
			switch col {
			case 0:
				return s.Bold(true)
			case 1:
				return s.Foreground(StateColor(issues[row].Field("State")))
			case 2:
				return s.Foreground(PriorityColor(issues[row].Field("Priority")))
			}
			return s
		})

	for _, issue := range issues {
		t.Row(
			issue.IDReadable,
			issue.Field("State"),
			issue.Field("Priority"),
			issue.Field("Assignee"),
			issue.Summary,
		)
	}

	return t.Render()
}
