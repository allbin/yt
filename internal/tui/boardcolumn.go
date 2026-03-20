package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/board"
	"github.com/allbin/yt/internal/format"
)

func (m *BoardViewer) renderColumn(colIdx, width int) string {
	columns := m.grid.Columns()
	col := columns[colIdx]
	curCol, curSL, curRow := m.grid.CursorPos()
	isFocused := colIdx == curCol

	if col.Minimized {
		return m.renderMinimizedColumn(col, isFocused)
	}

	innerWidth := max(width-4, 10)

	prefix := ""
	if col.IsResolved {
		prefix = "\uf00c "
	}
	header := fmt.Sprintf("%s%s (%d)", prefix, col.Presentation, col.IssueCount)
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

	swimlanes := m.grid.Swimlanes()
	numSL := m.grid.NumSwimlanes()
	for sl := range numSL {
		if len(swimlanes) > 0 {
			slDef := swimlanes[sl]
			issues := m.grid.CellIssues(colIdx, sl)
			count := len(issues)
			cursorHere := isFocused && sl == curSL

			if count == 0 && !cursorHere {
				continue
			}

			content.WriteString("\n")
			content.WriteString(renderSwimlaneDivider(slDef, count, innerWidth, isFocused, cursorHere))

			if slDef.Collapsed {
				continue
			}
		}

		issues := m.grid.CellIssues(colIdx, sl)
		for ri, issue := range issues {
			content.WriteString("\n")
			focused := isFocused && sl == curSL && ri == curRow
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

func (m *BoardViewer) renderMinimizedColumn(col board.Column, focused bool) string {
	borderColor := lipgloss.TerminalColor(format.ColorDim)
	if focused {
		borderColor = columnColor(col)
	}

	var lines []string
	for _, r := range col.Presentation {
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

func renderSwimlaneDivider(sl board.Swimlane, count, width int, colFocused, slFocused bool) string {
	indicator := "\u25be" // ▾ expanded
	if sl.Collapsed {
		indicator = "\u25b8" // ▸ collapsed
	}

	label := sl.Name
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

func columnColor(col board.Column) lipgloss.TerminalColor {
	if len(col.StateNames) > 0 {
		return format.StateColor(col.StateNames[0])
	}
	return format.ColorBorder
}

// --- Swimlane row-major rendering ---

func (m *BoardViewer) renderColumnHeader(colIdx, width int) string {
	columns := m.grid.Columns()
	col := columns[colIdx]
	curCol, _, _ := m.grid.CursorPos()
	isFocused := colIdx == curCol

	if col.Minimized {
		ch := string([]rune(col.Presentation)[0])
		style := lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
		if isFocused {
			style = style.Foreground(columnColor(col))
		} else {
			style = style.Foreground(format.ColorDim)
		}
		return style.Render(ch)
	}

	header := fmt.Sprintf("%s (%d)", col.Presentation, col.IssueCount)
	if col.IsResolved {
		header = "\uf00c " + header
	}
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
	swimlanes := m.grid.Swimlanes()
	sl := swimlanes[slIdx]
	_, curSL, _ := m.grid.CursorPos()
	isFocused := curSL == slIdx

	indicator := "\u25be" // ▾
	if sl.Collapsed {
		indicator = "\u25b8" // ▸
	}

	total := 0
	columns := m.grid.Columns()
	for ci := range columns {
		total += len(m.grid.CellIssues(ci, slIdx))
	}

	label := sl.Name
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
	columns := m.grid.Columns()
	col := columns[colIdx]
	if col.Minimized {
		return lipgloss.NewStyle().Width(width).Render("")
	}

	curCol, curSL, curRow := m.grid.CursorPos()
	colFocused := colIdx == curCol
	slFocused := slIdx == curSL
	issues := m.grid.CellIssues(colIdx, slIdx)
	innerWidth := max(width-2, 10)

	var b strings.Builder
	for i, issue := range issues {
		if i > 0 {
			b.WriteString("\n")
		}
		focused := colFocused && slFocused && i == curRow
		b.WriteString(renderCard(issue, innerWidth, focused, !colFocused))
	}

	return lipgloss.NewStyle().Width(width).Padding(0, 1).Render(b.String())
}
