package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/allbin/yt/internal/board"
	"github.com/allbin/yt/internal/boarddata"
	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/state"
	"github.com/allbin/yt/internal/tui/modal"
	"github.com/allbin/yt/internal/youtrack"
)

type boardMode int

const (
	boardModeNormal boardMode = iota
	boardModeIssueViewer
)

// BoardViewer is a full-screen bubbletea model for viewing an agile board.
type BoardViewer struct {
	client     youtrack.API
	fetcher    *boarddata.Fetcher
	boardName  string
	sprintName string

	board   *youtrack.Agile
	grid    *board.Grid
	err     error
	loading bool

	scrollOffset int
	width        int
	height       int
	rendered     []string

	mode        boardMode
	issueViewer IssueViewer
	modals      modal.Stack

	appState *state.AppState
}

type boardLoadedMsg struct {
	board       *youtrack.Agile
	issues      []youtrack.Issue
	sprintBoard *youtrack.SprintBoard
	err         error
}

type boardRefreshedMsg struct {
	issues      []youtrack.Issue
	sprintBoard *youtrack.SprintBoard
	err         error
}

type boardStatesLoadedMsg struct {
	issueID string
	summary string
	current string
	states  []youtrack.StateBundleElement
	err     error
}

func NewBoardViewer(client youtrack.API, boardName, sprintName string, appState *state.AppState) BoardViewer {
	return BoardViewer{
		client:     client,
		fetcher:    boarddata.New(client),
		boardName:  boardName,
		sprintName: sprintName,
		loading:    true,
		appState:   appState,
	}
}

func (m BoardViewer) Init() tea.Cmd {
	return loadBoardCmd(m.fetcher, m.boardName, m.sprintName)
}

func (m BoardViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		if m.mode == boardModeIssueViewer {
			updated, cmd := m.issueViewer.Update(wsm)
			m.issueViewer = updated.(IssueViewer)
			return m, cmd
		}
		if m.modals.Active() {
			_, cmd := m.modals.Update(wsm)
			return m, cmd
		}
		if m.grid != nil {
			m.grid.SetWidth(m.width)
			m.rebuildContent()
		}
		return m, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.modals.Active() {
		popped, cmd := m.modals.Update(msg)
		if popped != nil {
			resultCmd := m.handleModalResult(popped)
			if resultCmd != nil {
				m.loading = true
			}
			return m, tea.Batch(cmd, resultCmd)
		}
		return m, cmd
	}

	switch m.mode {
	case boardModeIssueViewer:
		return m.updateIssueViewer(msg)
	default:
		return m.updateNormal(msg)
	}
}

func (m BoardViewer) updateIssueViewer(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if !m.issueViewer.modals.Active() {
			switch km.String() {
			case "esc":
				m.mode = boardModeNormal
				m.loading = true
				m.rebuildContent()
				return m, refreshBoardCmd(m.fetcher, m.board, m.sprintName)
			case "q":
				return m, tea.Quit
			}
		}
	}

	updated, cmd := m.issueViewer.Update(msg)
	m.issueViewer = updated.(IssueViewer)
	return m, cmd
}

func (m BoardViewer) handleModalResult(popped modal.Modal) tea.Cmd {
	if sp, ok := popped.(StatePicker); ok {
		r := sp.Result()
		if r.Cancelled {
			return nil
		}
		issue := m.grid.FocusedIssue()
		if issue != nil && r.State != issue.Field("State") {
			return refreshAfterStateChange(m.fetcher, issue.IDReadable, r.State, m.board, m.sprintName)
		}
	}
	return nil
}

func (m BoardViewer) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case boardLoadedMsg:
		m.loading = false
		if msg.board != nil {
			m.board = msg.board
		}
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		layout := loadLayout(m.appState, m.board.ID)
		if msg.sprintBoard != nil {
			m.grid = board.FromSprintBoard(m.board, msg.sprintBoard, layout)
		} else {
			m.grid = board.FromAgile(m.board, msg.issues, layout)
		}
		m.grid.SetWidth(m.width)
		m.rebuildContent()
		return m, nil

	case boardRefreshedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		savedID := ""
		if m.grid != nil {
			if issue := m.grid.FocusedIssue(); issue != nil {
				savedID = issue.IDReadable
			}
		}
		layout := loadLayout(m.appState, m.board.ID)
		if msg.sprintBoard != nil {
			m.grid = board.FromSprintBoard(m.board, msg.sprintBoard, layout)
		} else {
			m.grid = board.FromAgile(m.board, msg.issues, layout)
		}
		m.grid.RestoreCursor(savedID)
		m.grid.SetWidth(m.width)
		m.rebuildContent()
		m.ensureFocusedVisible()
		return m, nil

	case boardStatesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		if len(msg.states) == 0 {
			return m, nil
		}
		sp := NewStatePicker(msg.issueID, msg.summary, msg.current, msg.states)
		m.modals.Push(sp)
		return m, nil
	}

	return m, nil
}

func (m BoardViewer) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	}

	if m.loading || m.grid == nil {
		return m, nil
	}

	autoScroll := false
	vp := m.viewportHeight()
	switch msg.String() {
	case "h", "left":
		m.grid.MoveCol(-1)
		autoScroll = true
	case "l", "right":
		m.grid.MoveCol(1)
		autoScroll = true
	case "j", "down":
		m.grid.MoveRow(1)
		autoScroll = true
	case "k", "up":
		m.grid.MoveRow(-1)
		autoScroll = true
	case "J":
		m.grid.MoveSwimlane(1)
		autoScroll = true
	case "K":
		m.grid.MoveSwimlane(-1)
		autoScroll = true
	case "enter":
		if issue := m.grid.FocusedIssue(); issue != nil {
			m.issueViewer = NewIssueViewer(m.client, issue.IDReadable)
			m.mode = boardModeIssueViewer
			initCmd := m.issueViewer.Init()
			sizeCmd := func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
			return m, tea.Batch(initCmd, sizeCmd)
		}
	case "s":
		if issue := m.grid.FocusedIssue(); issue != nil {
			return m, loadStatesCmd(m.client, issue.IDReadable, issue.Summary, issue.Field("State"))
		}
	case "z":
		layout := m.grid.ToggleCollapse()
		saveLayout(m.appState, m.board.ID, layout)
	case "m":
		layout := m.grid.ToggleMinimize()
		saveLayout(m.appState, m.board.ID, layout)
	case "r":
		m.loading = true
		return m, refreshBoardCmd(m.fetcher, m.board, m.sprintName)
	case " ", "pgdown":
		m.scrollOffset = min(m.scrollOffset+vp, m.maxScroll())
	case "pgup":
		m.scrollOffset = max(m.scrollOffset-vp, 0)
	case "ctrl+d":
		m.scrollOffset = min(m.scrollOffset+vp/2, m.maxScroll())
	case "ctrl+u":
		m.scrollOffset = max(m.scrollOffset-vp/2, 0)
	case "g":
		m.scrollOffset = 0
	case "G":
		m.scrollOffset = m.maxScroll()
	}

	m.rebuildContent()
	if autoScroll {
		m.ensureFocusedVisible()
	}
	return m, nil
}

func (m BoardViewer) View() string {
	if m.modals.Active() {
		return m.modals.View()
	}
	if m.mode == boardModeIssueViewer {
		return m.issueViewer.View()
	}
	if m.err != nil {
		return fmt.Sprintf("error: %v\n\npress q to quit\n", m.err)
	}
	if m.loading || m.grid == nil {
		return "Loading\u2026\n"
	}

	var b strings.Builder
	b.WriteString(m.renderBoardHeader())

	vpHeight := m.viewportHeight()
	start := min(m.scrollOffset, len(m.rendered))
	end := min(start+vpHeight, len(m.rendered))

	vpLines := m.rendered[start:end]
	vpLines = m.applyColumnTooltip(vpLines)

	for _, line := range vpLines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	for range vpHeight - (end - start) {
		b.WriteString("\n")
	}

	b.WriteString(m.renderBoardFooter())
	return b.String()
}

// --- Rendering ---

func (m *BoardViewer) rebuildContent() {
	m.rendered = nil
	columns := m.grid.Columns()
	if len(columns) == 0 {
		m.rendered = []string{"No columns configured."}
		return
	}

	if m.grid.HasSwimlanes() {
		m.rebuildContentSwimlanes()
		return
	}

	widths := m.grid.ColumnWidths()
	renderCols := m.grid.VisibleColumns()

	if len(renderCols) == 0 {
		m.rendered = []string{"No columns visible."}
		return
	}

	var colStrings []string
	for _, ci := range renderCols {
		colStrings = append(colStrings, m.renderColumn(ci, widths[ci]))
	}

	joined := lipgloss.JoinHorizontal(lipgloss.Top, colStrings...)
	for _, line := range strings.Split(joined, "\n") {
		if lipgloss.Width(line) > m.width {
			line = ansi.Truncate(line, m.width, "")
		}
		m.rendered = append(m.rendered, line)
	}
}

func (m *BoardViewer) rebuildContentSwimlanes() {
	widths := m.grid.ColumnWidths()
	renderCols := m.grid.VisibleColumns()

	if len(renderCols) == 0 {
		m.rendered = []string{"No columns visible."}
		return
	}

	totalWidth := 0
	for _, ci := range renderCols {
		totalWidth += widths[ci]
	}
	totalWidth = min(totalWidth, m.width)

	appendLines := func(s string) {
		for _, line := range strings.Split(s, "\n") {
			if lipgloss.Width(line) > m.width {
				line = ansi.Truncate(line, m.width, "")
			}
			m.rendered = append(m.rendered, line)
		}
	}

	curCol, curSL, _ := m.grid.CursorPos()
	numSL := m.grid.NumSwimlanes()
	for sl := range numSL {
		hasIssues := false
		for _, ci := range renderCols {
			if len(m.grid.CellIssues(ci, sl)) > 0 {
				hasIssues = true
				break
			}
		}
		cursorHere := sl == curSL
		if !hasIssues && !cursorHere {
			continue
		}

		appendLines(m.renderSwimlaneBanner(sl, totalWidth))

		swimlanes := m.grid.Swimlanes()
		if swimlanes[sl].Collapsed {
			continue
		}

		var cellParts []string
		for _, ci := range renderCols {
			cellParts = append(cellParts, m.renderColumnCell(ci, sl, widths[ci]))
		}
		appendLines(lipgloss.JoinHorizontal(lipgloss.Top, cellParts...))
	}

	_ = curCol // used via CursorPos in render helpers
}

func (m *BoardViewer) renderBoardHeader() string {
	var b strings.Builder

	title := format.StyleBold.Render(m.board.Name)
	sprint := ""
	if m.board.CurrentSprint != nil {
		sprint = format.StyleDim.Render(" / " + m.board.CurrentSprint.Name)
	}
	if m.loading {
		sprint += format.StyleDim.Render("  refreshing\u2026")
	}

	curCol, _, _ := m.grid.CursorPos()
	columns := m.grid.Columns()
	colPos := format.StyleDim.Render(fmt.Sprintf("  [%d/%d]", curCol+1, len(columns)))

	laneInfo := ""
	swimlanes := m.grid.Swimlanes()
	_, curSL, _ := m.grid.CursorPos()
	if len(swimlanes) > 0 && curSL < len(swimlanes) {
		sl := swimlanes[curSL]
		if sl.IssueID != "" {
			laneInfo = "  " + format.StyleID.Render(sl.IssueID)
		} else {
			laneInfo = format.StyleDim.Render("  " + sl.Name)
		}
	}

	hintStr := "  h/l cols  j/k rows"
	if m.grid.HasSwimlanes() {
		hintStr += "  J/K lanes  z fold"
	}
	hintStr += "  enter view  s state  m min  r refresh  q quit"
	hints := format.StyleDim.Render(hintStr)
	b.WriteString(title + sprint + colPos + laneInfo + "\n")
	b.WriteString(hints + "\n")

	// Sticky column headers for swimlane mode
	if m.grid.HasSwimlanes() {
		widths := m.grid.ColumnWidths()
		renderCols := m.grid.VisibleColumns()
		var parts []string
		for _, ci := range renderCols {
			parts = append(parts, m.renderColumnHeader(ci, widths[ci]))
		}
		if len(parts) > 0 {
			row := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
			if lipgloss.Width(row) > m.width {
				row = ansi.Truncate(row, m.width, "")
			}
			b.WriteString(row + "\n")
		}
	}

	return b.String()
}

func (m BoardViewer) renderBoardFooter() string {
	issue := m.grid.FocusedIssue()
	left := ""
	curCol, curSL, _ := m.grid.CursorPos()
	columns := m.grid.Columns()
	swimlanes := m.grid.Swimlanes()

	if issue != nil {
		state := issue.Field("State")
		stateStyled := lipgloss.NewStyle().Foreground(format.StateColor(state)).Render(state)
		left = format.StyleID.Render(issue.IDReadable) + " " + stateStyled
		if len(swimlanes) > 0 && curSL < len(swimlanes) {
			sl := swimlanes[curSL]
			lname := sl.Name
			if sl.IssueID != "" {
				lname = sl.IssueID
			}
			left += format.StyleDim.Render("  [" + lname + "]")
		}
	} else if curCol < len(columns) && columns[curCol].Minimized {
		left = format.StyleDim.Render(columns[curCol].Presentation + " (minimized)")
	} else if len(swimlanes) > 0 && curSL < len(swimlanes) {
		sl := swimlanes[curSL]
		if sl.Collapsed {
			left = format.StyleDim.Render(sl.Name + " (collapsed)")
		}
	}

	pos := ""
	if len(m.rendered) > 0 {
		pos = fmt.Sprintf("%d/%d", m.scrollOffset+1, len(m.rendered))
	}
	gap := max(m.width-lipgloss.Width(left)-lipgloss.Width(pos), 2)
	return left + strings.Repeat(" ", gap) + format.StyleDim.Render(pos)
}

// --- Tooltip overlay ---

func (m *BoardViewer) applyColumnTooltip(vpLines []string) []string {
	if m.grid == nil {
		return vpLines
	}
	curCol, _, _ := m.grid.CursorPos()
	columns := m.grid.Columns()
	if curCol >= len(columns) || !columns[curCol].Minimized {
		return vpLines
	}

	widths := m.grid.ColumnWidths()
	renderCols := m.grid.VisibleColumns()
	x := 0
	found := false
	for _, ci := range renderCols {
		if ci == curCol {
			found = true
			break
		}
		x += widths[ci]
	}
	if !found {
		return vpLines
	}

	tooltip := m.renderColumnTooltip(columns[curCol])
	tooltipLines := strings.Split(tooltip, "\n")

	result := make([]string, len(vpLines))
	copy(result, vpLines)
	for i, tl := range tooltipLines {
		if i >= len(result) {
			break
		}
		result[i] = overlayOnLine(result[i], tl, x, m.width)
	}
	return result
}

func overlayOnLine(base, insert string, x, maxWidth int) string {
	insertWidth := lipgloss.Width(insert)
	prefix := ansi.Truncate(base, x, "")
	suffix := ansi.TruncateLeft(base, x+insertWidth, "")

	if pw := lipgloss.Width(prefix); pw < x {
		prefix += strings.Repeat(" ", x-pw)
	}

	result := prefix + insert + suffix
	if maxWidth > 0 && lipgloss.Width(result) > maxWidth {
		result = ansi.Truncate(result, maxWidth, "")
	}
	return result
}

// --- Scrolling ---

func (m BoardViewer) viewportHeight() int {
	h := 3
	if m.grid != nil && m.grid.HasSwimlanes() {
		h++
	}
	return max(m.height-h, 1)
}

func (m BoardViewer) maxScroll() int {
	return max(len(m.rendered)-m.viewportHeight(), 0)
}

func (m *BoardViewer) ensureFocusedVisible() {
	start, h := m.focusedCardPosition()
	if h == 0 {
		return
	}
	vp := m.viewportHeight()
	if start < m.scrollOffset {
		m.scrollOffset = start
	}
	if start+h > m.scrollOffset+vp {
		m.scrollOffset = start + h - vp
	}
	m.scrollOffset = max(min(m.scrollOffset, m.maxScroll()), 0)
}

func (m *BoardViewer) focusedCardPosition() (start, height int) {
	if m.grid.HasSwimlanes() {
		return m.focusedCardPositionSwimlanes()
	}
	return m.focusedCardPositionColumns()
}

func (m *BoardViewer) focusedCardPositionColumns() (start, height int) {
	col, _, row := m.grid.CursorPos()
	columns := m.grid.Columns()
	if col >= len(columns) || columns[col].Minimized {
		return 0, 0
	}

	widths := m.grid.ColumnWidths()
	innerWidth := max(widths[col]-4, 10)

	line := 2
	issues := m.grid.CellIssues(col, 0)
	for ri, issue := range issues {
		card := renderCard(issue, innerWidth, false, false)
		cardH := strings.Count(card, "\n") + 1
		if ri == row {
			return line, cardH
		}
		line += cardH
	}
	return 0, 0
}

func (m *BoardViewer) focusedCardPositionSwimlanes() (start, height int) {
	col, curSL, row := m.grid.CursorPos()
	columns := m.grid.Columns()
	swimlanes := m.grid.Swimlanes()
	if col >= len(columns) || columns[col].Minimized {
		return 0, 0
	}
	if curSL >= len(swimlanes) {
		return 0, 0
	}

	widths := m.grid.ColumnWidths()
	renderCols := m.grid.VisibleColumns()
	innerWidth := max(widths[col]-2, 10)

	line := 0
	numSL := m.grid.NumSwimlanes()
	for s := range numSL {
		hasIssues := false
		for _, ci := range renderCols {
			if len(m.grid.CellIssues(ci, s)) > 0 {
				hasIssues = true
				break
			}
		}
		cursorHere := s == curSL
		if !hasIssues && !cursorHere {
			continue
		}

		line++ // banner

		if swimlanes[s].Collapsed {
			if cursorHere {
				return line - 1, 1
			}
			continue
		}

		if s == curSL {
			cardLine := line
			for ri, issue := range m.grid.CellIssues(col, curSL) {
				card := renderCard(issue, innerWidth, false, false)
				cardH := strings.Count(card, "\n") + 1
				if ri == row {
					return cardLine, cardH
				}
				cardLine += cardH
			}
			return 0, 0
		}

		maxH := 0
		for _, ci := range renderCols {
			if columns[ci].Minimized {
				continue
			}
			cellH := 0
			for _, issue := range m.grid.CellIssues(ci, s) {
				card := renderCard(issue, max(widths[ci]-2, 10), false, false)
				cellH += strings.Count(card, "\n") + 1
			}
			maxH = max(maxH, cellH)
		}
		if maxH == 0 {
			maxH = 1
		}
		line += maxH
	}

	return 0, 0
}

// --- Helpers ---

func truncateToWidth(s string, width int) string {
	runes := []rune(s)
	for i := range runes {
		if lipgloss.Width(string(runes[:i+1])) > width {
			return string(runes[:i])
		}
	}
	return s
}

// --- State persistence bridge ---

func loadLayout(s *state.AppState, boardID string) board.Layout {
	if s == nil {
		return board.Layout{}
	}
	return board.Layout{
		MinimizedColumns: s.BoardMinimized(boardID),
		CollapsedLanes:   s.BoardCollapsed(boardID),
	}
}

func saveLayout(s *state.AppState, boardID string, l board.Layout) {
	if s == nil {
		return
	}
	s.SetBoardMinimized(boardID, l.MinimizedColumns)
	s.SetBoardCollapsed(boardID, l.CollapsedLanes)
	_ = s.Save()
}

// --- Commands ---

func loadBoardCmd(f *boarddata.Fetcher, boardName, sprintName string) tea.Cmd {
	return func() tea.Msg {
		r, err := f.Load(boardName, sprintName)
		return boardLoadedMsg{board: r.Board, issues: r.Issues, sprintBoard: r.SprintBoard, err: err}
	}
}

func refreshBoardCmd(f *boarddata.Fetcher, board *youtrack.Agile, sprintName string) tea.Cmd {
	return func() tea.Msg {
		r, err := f.Refresh(board, sprintName)
		return boardRefreshedMsg{issues: r.Issues, sprintBoard: r.SprintBoard, err: err}
	}
}

func refreshAfterStateChange(f *boarddata.Fetcher, issueID, state string, board *youtrack.Agile, sprintName string) tea.Cmd {
	return func() tea.Msg {
		r, err := f.SetStateAndRefresh(issueID, state, board, sprintName)
		return boardRefreshedMsg{issues: r.Issues, sprintBoard: r.SprintBoard, err: err}
	}
}

func loadStatesCmd(client youtrack.API, issueID, summary, current string) tea.Cmd {
	return func() tea.Msg {
		states, err := client.GetIssueStates(issueID)
		return boardStatesLoadedMsg{
			issueID: issueID,
			summary: summary,
			current: current,
			states:  states,
			err:     err,
		}
	}
}
