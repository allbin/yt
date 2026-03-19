package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/format"
)

const (
	minColWidth    = 28
	minimizedWidth = 5 // 1 char text + 2 padding + 2 border
)

func (m *BoardViewer) columnWidths() []int {
	numCols := len(m.columns)
	widths := make([]int, numCols)

	minimizedSpace := 0
	visibleCount := 0
	for _, col := range m.columns {
		if col.minimized {
			minimizedSpace += minimizedWidth
		} else {
			visibleCount++
		}
	}
	if visibleCount == 0 {
		for i := range widths {
			widths[i] = minimizedWidth
		}
		return widths
	}

	available := max(m.width-minimizedSpace, minColWidth)
	maxFit := max(available/minColWidth, 1)
	colWidth := minColWidth
	if visibleCount <= maxFit {
		colWidth = available / visibleCount
	}

	for i, col := range m.columns {
		if col.minimized {
			widths[i] = minimizedWidth
		} else {
			widths[i] = colWidth
		}
	}
	return widths
}

func (m *BoardViewer) renderColumn(colIdx, width int) string {
	col := m.columns[colIdx]
	isFocused := colIdx == m.cursor.col

	if col.minimized {
		return m.renderMinimizedColumn(col, isFocused)
	}

	innerWidth := max(width-4, 10)

	count := m.columnIssueCount(colIdx)
	prefix := ""
	if col.isResolved {
		prefix = "\uf00c "
	}
	header := fmt.Sprintf("%s%s (%d)", prefix, col.presentation, count)
	textArea := max(innerWidth-2, 1)
	if lipgloss.Width(header) > textArea {
		header = truncateToWidth(header, textArea)
	}

	var content strings.Builder
	if isFocused {
		headerStyle := lipgloss.NewStyle().
			Foreground(columnColor(col)).
			Bold(true).
			Reverse(true)
		gap := max(textArea-lipgloss.Width(header), 0)
		content.WriteString(headerStyle.Render(header + strings.Repeat(" ", gap)))
	} else {
		content.WriteString(format.StyleBold.Render(header))
	}

	numSL := m.numSwimlanes()
	for sl := range numSL {
		if len(m.swimlanes) > 0 {
			content.WriteString("\n")
			name := m.swimlanes[sl]
			divider := fmt.Sprintf("\u2500\u2500 %s ", name)
			if pad := innerWidth - lipgloss.Width(divider); pad > 0 {
				divider += strings.Repeat("\u2500", pad)
			}
			content.WriteString(format.StyleDim.Render(divider))
		}

		issues := m.issues[colIdx][sl]
		for ri, issue := range issues {
			content.WriteString("\n")
			focused := isFocused && sl == m.cursor.swimlane && ri == m.cursor.row
			content.WriteString(renderCard(issue, innerWidth, focused, !isFocused))
		}
	}

	borderColor := lipgloss.TerminalColor(format.ColorDim)
	if isFocused {
		borderColor = columnColor(col)
	}
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerWidth).
		Padding(0, 1)

	return style.Render(content.String())
}

func (m *BoardViewer) renderMinimizedColumn(col columnDef, focused bool) string {
	borderColor := lipgloss.TerminalColor(format.ColorDim)
	if focused {
		borderColor = columnColor(col)
	}

	var lines []string
	for _, r := range col.presentation {
		lines = append(lines, string(r))
	}
	if len(lines) == 0 {
		lines = []string{" "}
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(3).
		Padding(0, 1).
		Align(lipgloss.Center)

	return style.Render(strings.Join(lines, "\n"))
}

func columnColor(col columnDef) lipgloss.TerminalColor {
	if len(col.stateNames) > 0 {
		return format.StateColor(col.stateNames[0])
	}
	return format.ColorBorder
}

func (m *BoardViewer) columnIssueCount(colIdx int) int {
	count := 0
	for sl := range m.numSwimlanes() {
		count += len(m.issues[colIdx][sl])
	}
	return count
}

