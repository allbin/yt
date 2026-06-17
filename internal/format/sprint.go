package format

import (
	"fmt"
	"io"
	"time"

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

	t := newTable("SPRINT", "DATES", "").
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(ColorAccent)
			}
			switch col {
			case 0:
				return s.Bold(true)
			case 1:
				return s.Foreground(ColorDim)
			default:
				return s.Foreground(ColorAccent)
			}
		})

	for _, sp := range sprints {
		marker := ""
		if sp.Name == current {
			marker = "● current"
		}
		t.Row(sp.Name, sprintDates(sp), marker)
	}

	_, err := fmt.Fprintln(w, t.Render())
	return err
}

// sprintDates renders a sprint's start/finish range, e.g. "2025-06-01 → 2025-06-14".
// Returns "" when no dates are set.
func sprintDates(sp youtrack.Sprint) string {
	switch {
	case sp.Start != nil && sp.Finish != nil:
		return millisDate(*sp.Start) + " → " + millisDate(*sp.Finish)
	case sp.Start != nil:
		return millisDate(*sp.Start) + " →"
	case sp.Finish != nil:
		return "→ " + millisDate(*sp.Finish)
	default:
		return ""
	}
}

func millisDate(ms int64) string {
	return time.Unix(ms/1000, 0).Format("2006-01-02")
}
