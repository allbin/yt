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
	v := issue.View()
	ew := &errWriter{w: w}

	ew.printf("%s  %s\n", StyleID.Render(v.ID), StyleBold.Render(v.Summary))

	hasMeta := v.State != "" || v.Assignee != "" || v.Priority != "" || v.Type != "" || v.Subsystem != "" || v.Tags != ""
	if hasMeta {
		ew.println()
	}
	if v.State != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("State"), lipgloss.NewStyle().Foreground(StateColor(v.State)).Render(v.State))
	}
	if v.Assignee != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Assignee"), v.Assignee)
	}
	if v.Priority != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Priority"), lipgloss.NewStyle().Foreground(PriorityColor(v.Priority)).Render(v.Priority))
	}
	if v.Type != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Type"), v.Type)
	}
	if v.Subsystem != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Subsystem"), v.Subsystem)
	}
	if v.Tags != "" {
		ew.printf("  %s %s\n", StyleLabel.Render("Tags"), StyleDim.Render(v.Tags))
	}

	if v.Description != "" {
		ew.println()
		ew.println(StyleRule.Render("────────────────────────────────────"))
		rendered := strings.Trim(RenderMarkdown(v.Description, 80), "\n")
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
	views := make([]youtrack.IssueView, len(issues))
	for i := range issues {
		views[i] = issues[i].View()
	}

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
				return s.Foreground(StateColor(views[row].State))
			case 2:
				return s.Foreground(PriorityColor(views[row].Priority))
			}
			return s
		})

	for _, v := range views {
		t.Row(v.ID, v.State, v.Priority, v.Assignee, v.Summary)
	}

	return t.Render()
}
