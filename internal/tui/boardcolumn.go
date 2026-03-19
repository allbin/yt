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
			slDef := m.swimlanes[sl]
			issues := m.issues[colIdx][sl]
			count := len(issues)
			cursorHere := isFocused && sl == m.cursor.swimlane

			if count == 0 && !cursorHere {
				continue
			}

			content.WriteString("\n")
			content.WriteString(m.renderSwimlaneDivider(slDef, count, innerWidth, isFocused, cursorHere))

			if slDef.collapsed {
				continue
			}
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

func (m *BoardViewer) renderSwimlaneDivider(sl swimlaneDef, count, width int, colFocused, slFocused bool) string {
	indicator := "\u25be" // ▾ expanded
	if sl.collapsed {
		indicator = "\u25b8" // ▸ collapsed
	}

	label := sl.name
	countStr := fmt.Sprintf(" (%d)", count)

	maxLabel := width - 2 - lipgloss.Width(countStr) - 3
	if maxLabel > 0 && lipgloss.Width(label) > maxLabel {
		label = truncateToWidth(label, maxLabel-1) + "\u2026"
	}

	text := indicator + " " + label + countStr + " "
	fill := max(width-lipgloss.Width(text), 0)
	line := text + strings.Repeat("\u2500", fill)

	if !colFocused {
		return format.StyleDim.Render(line)
	}
	if slFocused {
		return lipgloss.NewStyle().
			Foreground(format.ColorAccent).
			Bold(true).
			Render(line)
	}
	return line
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

// --- Swimlane row-major rendering ---

func (m *BoardViewer) renderColumnHeader(colIdx, width int) string {
	col := m.columns[colIdx]
	isFocused := colIdx == m.cursor.col

	if col.minimized {
		ch := string([]rune(col.presentation)[0])
		style := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
		if isFocused {
			style = style.Foreground(columnColor(col))
		} else {
			style = style.Foreground(format.ColorDim)
		}
		return style.Render(ch)
	}

	count := m.columnIssueCount(colIdx)
	prefix := ""
	if col.isResolved {
		prefix = "\uf00c "
	}
	header := fmt.Sprintf("%s%s (%d)", prefix, col.presentation, count)
	if lipgloss.Width(header) > width-2 {
		header = truncateToWidth(header, width-3) + "\u2026"
	}

	style := lipgloss.NewStyle().Width(width).Padding(0, 1)
	if isFocused {
		style = style.Foreground(columnColor(col)).Bold(true).Reverse(true)
	} else {
		style = style.Bold(true)
	}
	return style.Render(header)
}

func (m *BoardViewer) renderSwimlaneBanner(slIdx, totalWidth int) string {
	sl := m.swimlanes[slIdx]
	isFocused := m.cursor.swimlane == slIdx

	indicator := "\u25be" // ▾
	if sl.collapsed {
		indicator = "\u25b8" // ▸
	}

	total := 0
	for ci := range m.columns {
		total += len(m.issues[ci][slIdx])
	}

	label := sl.name
	countStr := fmt.Sprintf(" (%d)", total)
	maxLabel := totalWidth - 4 - lipgloss.Width(countStr)
	if maxLabel > 0 && lipgloss.Width(label) > maxLabel {
		label = truncateToWidth(label, maxLabel-1) + "\u2026"
	}

	text := indicator + " " + label + countStr + " "
	fill := max(totalWidth-lipgloss.Width(text), 0)
	line := text + strings.Repeat("\u2500", fill)

	if isFocused {
		return lipgloss.NewStyle().
			Foreground(format.ColorAccent).
			Bold(true).
			Render(line)
	}
	return format.StyleDim.Render(line)
}

func (m *BoardViewer) renderColumnCell(colIdx, slIdx, width int) string {
	col := m.columns[colIdx]
	if col.minimized {
		return lipgloss.NewStyle().Width(width).Render("")
	}

	colFocused := colIdx == m.cursor.col
	slFocused := slIdx == m.cursor.swimlane
	issues := m.issues[colIdx][slIdx]
	innerWidth := max(width-2, 10) // account for padding

	var b strings.Builder
	for i, issue := range issues {
		if i > 0 {
			b.WriteString("\n")
		}
		focused := colFocused && slFocused && i == m.cursor.row
		b.WriteString(renderCard(issue, innerWidth, focused, !colFocused))
	}

	return lipgloss.NewStyle().Width(width).Padding(0, 1).Render(b.String())
}

