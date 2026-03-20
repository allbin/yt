package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
)

const minCardInnerWidth = 16

func renderCard(issue youtrack.Issue, width int, focused, dimmed bool) string {
	v := issue.View()
	border := lipgloss.RoundedBorder()
	borderColor := format.StateColor(v.State)
	if focused {
		border = lipgloss.ThickBorder()
		borderColor = format.ColorAccent
	} else if dimmed {
		borderColor = format.ColorBorder
	}

	// 2 for left+right border chars, 2 for padding
	innerWidth := max(width-4, minCardInnerWidth)

	style := lipgloss.NewStyle().
		Border(border).
		BorderForeground(borderColor).
		Width(innerWidth).
		Padding(0, 1)
	if focused {
		style = style.Bold(true)
	}

	var lines []string

	// Line 1: ID + priority
	icon := priorityIcon(v.Priority)
	dimStyle := lipgloss.NewStyle().Foreground(format.ColorDim)

	id := format.StyleID.Render(v.ID)
	if dimmed {
		id = dimStyle.Render(v.ID)
	}
	if icon != "" {
		pColor := format.PriorityColor(v.Priority)
		pStyle := lipgloss.NewStyle().Foreground(pColor)
		if dimmed {
			pStyle = dimStyle
		}
		right := pStyle.Render(icon)
		gap := max(innerWidth-2-lipgloss.Width(id)-lipgloss.Width(right), 1)
		lines = append(lines, id+strings.Repeat(" ", gap)+right)
	} else {
		lines = append(lines, id)
	}

	// Summary lines
	wrapped := wrapText(v.Summary, innerWidth-2)
	if dimmed {
		for i, line := range wrapped {
			wrapped[i] = dimStyle.Render(line)
		}
	}
	lines = append(lines, wrapped...)

	// Assignee
	if v.Assignee != "" {
		lines = append(lines, format.StyleDim.Render("\uf007 "+v.Assignee))
	}

	content := strings.Join(lines, "\n")
	return style.Render(content)
}

func priorityIcon(p string) string {
	lower := strings.ToLower(p)
	switch {
	case containsAny(lower, "critical", "show-stopper"):
		return " \uf139"
	case containsAny(lower, "major"):
		return " \uf139"
	case containsAny(lower, "minor", "nice to have"):
		return " \uf13a"
	default:
		return ""
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var result []string
	for paragraph := range strings.SplitSeq(text, "\n") {
		if paragraph == "" {
			result = append(result, "")
			continue
		}
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if len(line)+1+len(w) > width {
				result = append(result, line)
				line = w
			} else {
				line += " " + w
			}
		}
		result = append(result, line)
	}
	return result
}
