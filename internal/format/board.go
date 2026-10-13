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

	t := newTable("NAME", "KIND", "PROJECTS", "CURRENT SPRINT").
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
		t.Row(b.Name, b.Kind(), formatProjects(b.Projects), sprint)
	}

	_, err := fmt.Fprintln(w, t.Render())
	return err
}

// BoardIssues renders a board's issues under a header naming the sprint
// shown, or, on a sprintless board, what puts an issue on it.
func BoardIssues(w io.Writer, board *youtrack.Agile, sprint *youtrack.Sprint, issues []youtrack.Issue) error {
	ew := &errWriter{w: w}

	header := lipgloss.NewStyle().Bold(true).Foreground(ColorAccent).Render(board.Name)
	scope := StyleBold.Render(board.Membership())
	if sprint != nil {
		scope = StyleBold.Render(sprint.Name)
	}
	count := StyleDim.Render(fmt.Sprintf("(%d issues)", len(issues)))
	ew.printf("%s — %s  %s\n\n", header, scope, count)

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
