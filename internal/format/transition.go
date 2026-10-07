package format

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/allbin/yt/internal/youtrack"
)

// TransitionList renders field changes oldest first, one row per change.
func TransitionList(w io.Writer, transitions []youtrack.Transition) error {
	if len(transitions) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No changes found."))
		return err
	}

	t := newTable("TIME", "ISSUE", "FROM", "TO", "BY", "SUMMARY").
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Bold(true).Foreground(ColorAccent)
			}
			switch col {
			case 0:
				return s.Foreground(ColorDim)
			case 1:
				return s.Bold(true)
			case 2:
				return s.Foreground(StateColor(transitions[row].From))
			case 3:
				return s.Foreground(StateColor(transitions[row].To))
			}
			return s
		})

	for _, tr := range transitions {
		t.Row(
			time.UnixMilli(tr.Timestamp).Format("2006-01-02 15:04"),
			tr.Issue,
			valueOrNone(tr.From),
			valueOrNone(tr.To),
			userName(tr.Author),
			tr.Summary,
		)
	}

	_, err := fmt.Fprintln(w, t.Render())
	return err
}

func valueOrNone(v string) string {
	if v == "" {
		return "—"
	}
	return v
}

func userName(u *youtrack.User) string {
	switch {
	case u == nil:
		return "Unknown"
	case u.FullName != "":
		return u.FullName
	default:
		return u.Login
	}
}
