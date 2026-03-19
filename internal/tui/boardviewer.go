package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/state"
	"github.com/allbin/yt/internal/youtrack"
)

type boardMode int

const (
	boardModeNormal boardMode = iota
	boardModeIssueViewer
	boardModeStatePicker
)

type boardCursor struct {
	col      int // absolute column index (includes minimized)
	row      int // issue index within current column+swimlane
	swimlane int // swimlane index (0 if no swimlanes)
}

type columnDef struct {
	presentation string
	ordinal      int
	stateNames   []string
	isResolved   bool
	minimized    bool
}

// BoardViewer is a full-screen bubbletea model for viewing an agile board.
type BoardViewer struct {
	client     youtrack.API
	boardName  string
	sprintName string

	board       *youtrack.Agile
	columns     []columnDef
	swimlanes   []string // ordered swimlane names; empty slice = no swimlanes
	fieldName   string   // column field (usually "State")
	slFieldName string   // swimlane field name

	// issues[colIdx][swimlaneIdx] = []Issue
	issues    [][][]youtrack.Issue
	allIssues []youtrack.Issue

	err     error
	loading bool

	cursor       boardCursor
	width        int
	height       int
	scrollOffset int
	colOffset    int      // first column index rendered (horizontal scroll)
	rendered     []string // pre-rendered board lines

	mode        boardMode
	issueViewer IssueViewer
	statePicker StatePicker

	appState *state.AppState
}

type boardLoadedMsg struct {
	board  *youtrack.Agile
	issues []youtrack.Issue
	err    error
}

type boardRefreshedMsg struct {
	issues []youtrack.Issue
	err    error
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
		boardName:  boardName,
		sprintName: sprintName,
		loading:    true,
		appState:   appState,
	}
}

func (m BoardViewer) Init() tea.Cmd {
	return loadBoardCmd(m.client, m.boardName, m.sprintName)
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
		if m.board != nil {
			m.ensureColumnVisible()
			m.rebuildContent()
		}
		return m, nil
	}

	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.mode {
	case boardModeIssueViewer:
		return m.updateIssueViewer(msg)
	case boardModeStatePicker:
		return m.updateStatePicker(msg)
	default:
		return m.updateNormal(msg)
	}
}

func (m BoardViewer) updateIssueViewer(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if m.issueViewer.mode == modeNormal {
			switch km.String() {
			case "esc":
				m.mode = boardModeNormal
				m.loading = true
				m.rebuildContent()
				return m, refreshBoardCmd(m.client, m.board, m.sprintName)
			case "q":
				return m, tea.Quit
			}
		}
	}

	updated, cmd := m.issueViewer.Update(msg)
	m.issueViewer = updated.(IssueViewer)
	return m, cmd
}

func (m BoardViewer) updateStatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.statePicker.Update(msg)
	m.statePicker = updated.(StatePicker)

	result := m.statePicker.Result()
	if !result.Cancelled && result.State == "" {
		return m, cmd
	}

	m.mode = boardModeNormal
	if !result.Cancelled {
		issue := m.focusedIssue()
		if issue != nil && result.State != issue.Field("State") {
			m.loading = true
			return m, refreshAfterStateChange(m.client, issue.IDReadable, result.State, m.board, m.sprintName)
		}
	}
	return m, nil
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
		m.allIssues = msg.issues
		m.parseColumns()
		m.applyMinimizedState()
		m.parseSwimlanes()
		m.buildGrid()
		m.ensureColumnVisible()
		m.rebuildContent()
		return m, nil

	case boardRefreshedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		savedID := ""
		if issue := m.focusedIssue(); issue != nil {
			savedID = issue.IDReadable
		}
		m.allIssues = msg.issues
		m.buildGrid()
		m.restoreCursor(savedID)
		m.ensureColumnVisible()
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
		m.statePicker = NewStatePicker(msg.issueID, msg.summary, msg.current, msg.states)
		m.mode = boardModeStatePicker
		return m, nil
	}

	return m, nil
}

func (m BoardViewer) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	}

	if m.loading || m.board == nil {
		return m, nil
	}

	autoScroll := false
	vp := m.viewportHeight()
	switch msg.String() {
	case "h", "left":
		m.moveCursorCol(-1)
		autoScroll = true
	case "l", "right":
		m.moveCursorCol(1)
		autoScroll = true
	case "j", "down":
		m.moveCursorRow(1)
		autoScroll = true
	case "k", "up":
		m.moveCursorRow(-1)
		autoScroll = true
	case "J":
		m.moveCursorSwimlane(1)
		autoScroll = true
	case "K":
		m.moveCursorSwimlane(-1)
		autoScroll = true
	case "enter":
		if issue := m.focusedIssue(); issue != nil {
			m.issueViewer = NewIssueViewer(m.client, issue.IDReadable)
			m.mode = boardModeIssueViewer
			initCmd := m.issueViewer.Init()
			sizeCmd := func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
			return m, tea.Batch(initCmd, sizeCmd)
		}
	case "s":
		if issue := m.focusedIssue(); issue != nil {
			return m, loadStatesCmd(m.client, issue.IDReadable, issue.Summary, issue.Field("State"))
		}
	case "m":
		m.toggleMinimize()
	case "r":
		m.loading = true
		return m, refreshBoardCmd(m.client, m.board, m.sprintName)
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

	m.ensureColumnVisible()
	m.rebuildContent()
	if autoScroll {
		m.ensureFocusedVisible()
	}
	return m, nil
}

func (m BoardViewer) View() string {
	if m.mode == boardModeStatePicker {
		return m.statePicker.View()
	}
	if m.mode == boardModeIssueViewer {
		return m.issueViewer.View()
	}
	if m.err != nil {
		return fmt.Sprintf("error: %v\n\npress q to quit\n", m.err)
	}
	if m.loading || m.board == nil {
		return "Loading\u2026\n"
	}

	var b strings.Builder
	b.WriteString(m.renderBoardHeader())

	vpHeight := m.viewportHeight()
	start := min(m.scrollOffset, len(m.rendered))
	end := min(start+vpHeight, len(m.rendered))

	for i := start; i < end; i++ {
		b.WriteString(m.rendered[i])
		b.WriteString("\n")
	}
	for range vpHeight - (end - start) {
		b.WriteString("\n")
	}

	b.WriteString(m.renderBoardFooter())
	return b.String()
}

// --- Column Parsing ---

func (m *BoardViewer) parseColumns() {
	m.columns = nil
	m.fieldName = "State"

	if m.board.ColumnSettings != nil && len(m.board.ColumnSettings.Columns) > 0 {
		if m.board.ColumnSettings.Field != nil {
			m.fieldName = m.board.ColumnSettings.Field.Name
		}
		for _, col := range m.board.ColumnSettings.Columns {
			cd := columnDef{
				presentation: col.Presentation,
				ordinal:      col.Ordinal,
			}
			for _, fv := range col.FieldValues {
				cd.stateNames = append(cd.stateNames, fv.Name)
				if fv.IsResolved {
					cd.isResolved = true
				}
			}
			m.columns = append(m.columns, cd)
		}
		sort.Slice(m.columns, func(i, j int) bool {
			return m.columns[i].ordinal < m.columns[j].ordinal
		})
		return
	}

	// Fallback: derive columns from unique state values
	seen := map[string]bool{}
	var states []string
	for _, issue := range m.allIssues {
		s := issue.Field(m.fieldName)
		if s != "" && !seen[s] {
			seen[s] = true
			states = append(states, s)
		}
	}
	for i, s := range states {
		m.columns = append(m.columns, columnDef{
			presentation: s,
			ordinal:      i,
			stateNames:   []string{s},
		})
	}
}

func (m *BoardViewer) parseSwimlanes() {
	// TODO: implement proper swimlane support (per-issue swimlanes for Epic/User Story)
	m.swimlanes = nil
	m.slFieldName = ""
}

// --- Grid Building ---

func (m *BoardViewer) buildGrid() {
	numCols := len(m.columns)
	numSL := m.numSwimlanes()

	m.issues = make([][][]youtrack.Issue, numCols)
	for c := range numCols {
		m.issues[c] = make([][]youtrack.Issue, numSL)
	}

	stateToCol := map[string]int{}
	for ci, col := range m.columns {
		for _, sn := range col.stateNames {
			stateToCol[sn] = ci
		}
	}

	slToIdx := map[string]int{}
	for si, name := range m.swimlanes {
		slToIdx[name] = si
	}

	for _, issue := range m.allIssues {
		colIdx, ok := stateToCol[issue.Field(m.fieldName)]
		if !ok {
			continue
		}

		slIdx := 0
		if len(m.swimlanes) > 0 {
			sv := issue.Field(m.slFieldName)
			si, found := slToIdx[sv]
			if found {
				slIdx = si
			} else {
				slIdx = numSL - 1
			}
		}

		m.issues[colIdx][slIdx] = append(m.issues[colIdx][slIdx], issue)
	}
}

func (m *BoardViewer) numSwimlanes() int {
	if len(m.swimlanes) == 0 {
		return 1
	}
	return len(m.swimlanes)
}

// --- Horizontal Scrolling ---

func (m *BoardViewer) ensureColumnVisible() {
	if len(m.columns) == 0 {
		return
	}
	m.cursor.col = max(min(m.cursor.col, len(m.columns)-1), 0)

	widths := m.columnWidths()

	// Shift colOffset left if cursor is before it
	if m.cursor.col < m.colOffset {
		m.colOffset = m.cursor.col
	}

	// Shift colOffset right until cursor column fits in viewport
	for {
		used := 0
		cursorFits := false
		for ci := m.colOffset; ci < len(m.columns); ci++ {
			used += widths[ci]
			if used > m.width && ci > m.colOffset {
				break
			}
			if ci == m.cursor.col {
				cursorFits = true
				break
			}
		}
		if cursorFits {
			break
		}
		m.colOffset++
		if m.colOffset >= len(m.columns) {
			m.colOffset = m.cursor.col
			break
		}
	}

	// Clamp vertical scroll
	if mx := m.maxScroll(); m.scrollOffset > mx {
		m.scrollOffset = mx
	}
}

// --- Rendering ---

func (m *BoardViewer) rebuildContent() {
	m.rendered = nil
	if len(m.columns) == 0 {
		m.rendered = []string{"No columns configured."}
		return
	}

	widths := m.columnWidths()

	// Render columns whose left edge starts within the viewport (partial column at right edge)
	var renderCols []int
	used := 0
	for ci := m.colOffset; ci < len(m.columns); ci++ {
		if used >= m.width && len(renderCols) > 0 {
			break
		}
		renderCols = append(renderCols, ci)
		used += widths[ci]
	}

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

	// Column position indicator
	colPos := format.StyleDim.Render(fmt.Sprintf("  [%d/%d]", m.cursor.col+1, len(m.columns)))

	hints := format.StyleDim.Render("  h/l cols  j/k rows  enter view  s state  m min  r refresh  q quit")
	b.WriteString(title + sprint + colPos + "\n")
	b.WriteString(hints + "\n")

	return b.String()
}

func (m BoardViewer) renderBoardFooter() string {
	issue := m.focusedIssue()
	left := ""
	if issue != nil {
		state := issue.Field("State")
		stateStyled := lipgloss.NewStyle().Foreground(format.StateColor(state)).Render(state)
		left = format.StyleID.Render(issue.IDReadable) + " " + stateStyled
	} else if m.cursor.col < len(m.columns) && m.columns[m.cursor.col].minimized {
		left = format.StyleDim.Render(m.columns[m.cursor.col].presentation + " (minimized)")
	}

	pos := ""
	if len(m.rendered) > 0 {
		pos = fmt.Sprintf("%d/%d", m.scrollOffset+1, len(m.rendered))
	}
	gap := max(m.width-lipgloss.Width(left)-lipgloss.Width(pos), 2)
	return left + strings.Repeat(" ", gap) + format.StyleDim.Render(pos)
}

// --- Scrolling ---

func (m BoardViewer) viewportHeight() int {
	return max(m.height-3, 1) // 2 header lines + 1 footer
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

// focusedCardPosition returns the rendered line where the focused card starts
// and its height in lines, calculated analytically from column structure.
func (m *BoardViewer) focusedCardPosition() (start, height int) {
	col := m.cursor.col
	if col >= len(m.columns) || m.columns[col].minimized {
		return 0, 0
	}

	widths := m.columnWidths()
	innerWidth := max(widths[col]-4, 10)

	// Line 0: column top border, Line 1: header
	line := 2

	numSL := m.numSwimlanes()
	for s := range numSL {
		if len(m.swimlanes) > 0 {
			line++ // swimlane divider
		}

		issues := m.issues[col][s]
		for ri, issue := range issues {
			card := renderCard(issue, innerWidth, false, false)
			cardH := strings.Count(card, "\n") + 1

			if s == m.cursor.swimlane && ri == m.cursor.row {
				return line, cardH
			}
			line += cardH
		}
	}

	return 0, 0
}

// --- Cursor Navigation ---

func (m *BoardViewer) moveCursorCol(delta int) {
	if len(m.columns) == 0 {
		return
	}
	m.cursor.col = max(min(m.cursor.col+delta, len(m.columns)-1), 0)
	if !m.columns[m.cursor.col].minimized {
		m.clampRow()
	}
}

func (m *BoardViewer) moveCursorRow(delta int) {
	if m.columns[m.cursor.col].minimized {
		return
	}
	sl := m.cursor.swimlane
	issues := m.issues[m.cursor.col][sl]

	next := m.cursor.row + delta
	if next >= 0 && next < len(issues) {
		m.cursor.row = next
		return
	}

	numSL := m.numSwimlanes()
	if delta > 0 && next >= len(issues) {
		for s := sl + 1; s < numSL; s++ {
			if len(m.issues[m.cursor.col][s]) > 0 {
				m.cursor.swimlane = s
				m.cursor.row = 0
				return
			}
		}
		m.cursor.row = max(len(issues)-1, 0)
	} else if delta < 0 && next < 0 {
		for s := sl - 1; s >= 0; s-- {
			if len(m.issues[m.cursor.col][s]) > 0 {
				m.cursor.swimlane = s
				m.cursor.row = len(m.issues[m.cursor.col][s]) - 1
				return
			}
		}
		m.cursor.row = 0
	}
}

func (m *BoardViewer) moveCursorSwimlane(delta int) {
	if len(m.swimlanes) == 0 || m.columns[m.cursor.col].minimized {
		return
	}
	numSL := m.numSwimlanes()
	next := m.cursor.swimlane + delta

	if delta > 0 {
		for s := next; s < numSL; s++ {
			if len(m.issues[m.cursor.col][s]) > 0 {
				m.cursor.swimlane = s
				m.cursor.row = 0
				return
			}
		}
	} else {
		for s := next; s >= 0; s-- {
			if len(m.issues[m.cursor.col][s]) > 0 {
				m.cursor.swimlane = s
				m.cursor.row = 0
				return
			}
		}
	}
}

func (m *BoardViewer) clampRow() {
	col := m.cursor.col
	if col >= len(m.columns) || m.columns[col].minimized {
		return
	}
	numSL := m.numSwimlanes()
	sl := m.cursor.swimlane
	if sl >= numSL {
		sl = 0
		m.cursor.swimlane = 0
	}

	if len(m.issues[col][sl]) == 0 {
		for s := range numSL {
			if len(m.issues[col][s]) > 0 {
				m.cursor.swimlane = s
				m.cursor.row = 0
				return
			}
		}
		m.cursor.row = 0
		return
	}

	issues := m.issues[col][sl]
	if m.cursor.row >= len(issues) {
		m.cursor.row = max(len(issues)-1, 0)
	}
}

func (m *BoardViewer) toggleMinimize() {
	m.columns[m.cursor.col].minimized = !m.columns[m.cursor.col].minimized
	if !m.columns[m.cursor.col].minimized {
		m.clampRow()
	}
	m.saveMinimizedState()
}

func (m *BoardViewer) applyMinimizedState() {
	if m.appState == nil || m.board == nil {
		return
	}
	minimized := m.appState.BoardMinimized(m.board.ID)
	lookup := make(map[string]bool, len(minimized))
	for _, name := range minimized {
		lookup[name] = true
	}
	for i := range m.columns {
		m.columns[i].minimized = lookup[m.columns[i].presentation]
	}
}

func (m *BoardViewer) saveMinimizedState() {
	if m.appState == nil || m.board == nil {
		return
	}
	var names []string
	for _, col := range m.columns {
		if col.minimized {
			names = append(names, col.presentation)
		}
	}
	m.appState.SetBoardMinimized(m.board.ID, names)
	_ = m.appState.Save()
}

func (m *BoardViewer) focusedIssue() *youtrack.Issue {
	col := m.cursor.col
	if col >= len(m.columns) || m.columns[col].minimized {
		return nil
	}
	sl := m.cursor.swimlane
	if sl >= m.numSwimlanes() {
		return nil
	}
	issues := m.issues[col][sl]
	if m.cursor.row >= len(issues) {
		return nil
	}
	return &issues[m.cursor.row]
}

func (m *BoardViewer) restoreCursor(issueID string) {
	if issueID == "" {
		m.clampRow()
		return
	}
	for ci := range m.columns {
		if m.columns[ci].minimized {
			continue
		}
		for sl := range m.numSwimlanes() {
			for ri, issue := range m.issues[ci][sl] {
				if issue.IDReadable == issueID {
					m.cursor.col = ci
					m.cursor.swimlane = sl
					m.cursor.row = ri
					return
				}
			}
		}
	}
	m.clampRow()
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

func resolveSprint(board *youtrack.Agile, sprintName string) string {
	if sprintName != "" {
		return sprintName
	}
	if board.CurrentSprint != nil {
		return board.CurrentSprint.Name
	}
	return ""
}

// --- Commands ---

func loadBoardCmd(client youtrack.API, boardName, sprintName string) tea.Cmd {
	return func() tea.Msg {
		board, err := client.GetBoardForView(boardName)
		if err != nil {
			return boardLoadedMsg{err: err}
		}

		sprint := resolveSprint(board, sprintName)
		query := fmt.Sprintf("Board %s: {%s}", board.Name, sprint)
		issues, err := client.ListIssues(query, 0)
		if err != nil {
			return boardLoadedMsg{board: board, err: err}
		}

		return boardLoadedMsg{board: board, issues: issues}
	}
}

func refreshBoardCmd(client youtrack.API, board *youtrack.Agile, sprintName string) tea.Cmd {
	return func() tea.Msg {
		sprint := resolveSprint(board, sprintName)
		query := fmt.Sprintf("Board %s: {%s}", board.Name, sprint)
		issues, err := client.ListIssues(query, 0)
		return boardRefreshedMsg{issues: issues, err: err}
	}
}

func refreshAfterStateChange(client youtrack.API, issueID, state string, board *youtrack.Agile, sprintName string) tea.Cmd {
	return func() tea.Msg {
		if err := client.SetIssueState(issueID, state); err != nil {
			return boardRefreshedMsg{err: err}
		}
		sprint := resolveSprint(board, sprintName)
		query := fmt.Sprintf("Board %s: {%s}", board.Name, sprint)
		issues, err := client.ListIssues(query, 0)
		return boardRefreshedMsg{issues: issues, err: err}
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
