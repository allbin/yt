package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/allbin/yt/internal/youtrack"
)

func BoardList(w io.Writer, boards []youtrack.Agile) error {
	if len(boards) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No boards found."))
		return err
	}

	t := newTable("NAME", "PROJECTS", "CURRENT SPRINT").
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(ColorAccent)
			}
			if col == 0 {
				return s.Bold(true)
			}
			return s
		})

	for _, b := range boards {
		sprint := ""
		if b.CurrentSprint != nil {
			sprint = b.CurrentSprint.Name
		}
		t.Row(b.Name, formatProjects(b.Projects), sprint)
	}

	_, err := fmt.Fprintln(w, t.Render())
	return err
}

func SprintIssues(w io.Writer, board, sprint string, issues []youtrack.Issue) error {
	ew := &errWriter{w: w}

	header := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(board)
	sprintName := StyleBold.Render(sprint)
	count := StyleDim.Render(fmt.Sprintf("(%d issues)", len(issues)))
	ew.printf("%s — %s  %s\n\n", header, sprintName, count)

	if len(issues) == 0 {
		ew.println(StyleDim.Render("No issues."))
		return ew.err
	}

	ew.println(issueTable(issues))
	return ew.err
}

func formatProjects(projects []youtrack.Project) string {
	const max = 3
	names := make([]string, 0, max+1)
	for i, p := range projects {
		if i >= max {
			names = append(names, fmt.Sprintf("+%d", len(projects)-max))
			break
		}
		names = append(names, p.ShortName)
	}
	return strings.Join(names, ", ")
}
