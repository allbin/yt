package format

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/allbin/yt/internal/youtrack"
)

// SprintList renders a board's sprints, marking the current one.
func SprintList(w io.Writer, board *youtrack.Agile) error {
	sprints := board.SprintList()
	if len(sprints) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No sprints."))
		return err
	}

	current := ""
	if board.CurrentSprint != nil {
		current = board.CurrentSprint.Name
	}

	t := newTable("SPRINT", "").
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(ColorAccent)
			}
			if col == 0 {
				return s.Bold(true)
			}
			return s.Foreground(ColorAccent)
		})

	for _, sp := range sprints {
		marker := ""
		if sp.Name == current {
			marker = "● current"
		}
		t.Row(sp.Name, marker)
	}

	_, err := fmt.Fprintln(w, t.Render())
	return err
}
