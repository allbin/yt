package format

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	ColorAccent = lipgloss.Color("75")
	ColorDim    = lipgloss.Color("240")
	ColorGreen  = lipgloss.Color("76")
	ColorRed    = lipgloss.Color("196")
	ColorYellow = lipgloss.Color("214")
	ColorCyan   = lipgloss.Color("80")
	ColorBorder = lipgloss.Color("238")
)

var (
	StyleLabel = lipgloss.NewStyle().Foreground(ColorDim).Width(12)
	StyleBold  = lipgloss.NewStyle().Bold(true)
	StyleDim   = lipgloss.NewStyle().Foreground(ColorDim)
	StyleRule  = lipgloss.NewStyle().Foreground(ColorBorder)
	StyleID    = lipgloss.NewStyle().Bold(true).Foreground(ColorAccent)
)

func StateColor(state string) lipgloss.TerminalColor {
	lower := strings.ToLower(state)
	switch {
	// Terminal/rejected states must be checked before "complete"/"new" to
	// avoid substring collisions (e.g. "incomplete" contains "complete").
	case containsAny(lower, "obsolete", "duplicate", "won't fix", "can't reproduce", "incomplete"):
		return ColorDim
	case containsAny(lower, "done", "resolved", "fixed", "verified", "complete", "closed"):
		return ColorGreen
	case containsAny(lower, "review", "test", "testing"):
		return ColorAccent
	case containsAny(lower, "in dev", "in progress", "progress"):
		return ColorCyan
	case containsAny(lower, "planned", "scheduled"):
		return ColorDim
	case containsAny(lower, "open", "submitted", "new", "reopened"):
		return ColorYellow
	default:
		return lipgloss.NoColor{}
	}
}

func PriorityColor(priority string) lipgloss.TerminalColor {
	lower := strings.ToLower(priority)
	switch {
	case containsAny(lower, "critical", "show-stopper"):
		return ColorRed
	case containsAny(lower, "major"):
		return ColorYellow
	case containsAny(lower, "minor", "nice to have"):
		return ColorDim
	default:
		return lipgloss.NoColor{}
	}
}

func newTable(headers ...string) *table.Table {
	return table.New().
		Headers(headers...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorBorder))
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
