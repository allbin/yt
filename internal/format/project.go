package format

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/allbin/yt/internal/youtrack"
)

func ProjectList(w io.Writer, projects []youtrack.Project) error {
	if len(projects) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No projects found."))
		return err
	}

	t := newTable("SHORT", "NAME").
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

	for _, p := range projects {
		t.Row(p.ShortName, p.Name)
	}

	_, err := fmt.Fprintln(w, t.Render())
	return err
}
