package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
)

type viewerMode int

const (
	modeNormal viewerMode = iota
	modeStatePicker
)

type issueLoadedMsg struct {
	issue    *youtrack.Issue
	comments []youtrack.Comment
	states   []youtrack.StateBundleElement
	err      error
}

type stateChangedMsg struct{ err error }

// IssueViewer is a full-screen bubbletea model for viewing an issue.
type IssueViewer struct {
	client  youtrack.API
	issueID string

	issue    *youtrack.Issue
	comments []youtrack.Comment
	states   []youtrack.StateBundleElement
	err      error
	loading  bool

	lines        []string
	scrollOffset int
	headerHeight int
	width        int
	height       int

	mode        viewerMode
	statePicker StatePicker
}

func NewIssueViewer(client youtrack.API, issueID string) IssueViewer {
	return IssueViewer{
		client:  client,
		issueID: issueID,
		loading: true,
	}
}

func (m IssueViewer) Init() tea.Cmd {
	return loadIssueCmd(m.client, m.issueID)
}

func (m IssueViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		if m.issue != nil {
			m.rebuildContent()
		}
		return m, nil
	}

	if m.mode == modeStatePicker {
		return m.updateStatePicker(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case issueLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.issue = msg.issue
		m.comments = msg.comments
		m.states = msg.states
		m.err = nil
		m.rebuildContent()
		return m, nil

	case stateChangedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.loading = true
		return m, loadIssueCmd(m.client, m.issueID)
	}

	return m, nil
}

func (m *IssueViewer) rebuildContent() {
	header := m.renderHeader()
	m.headerHeight = strings.Count(header, "\n")
	m.lines = m.buildLines()
	if mx := m.maxScroll(); m.scrollOffset > mx {
		m.scrollOffset = mx
	}
}

func (m IssueViewer) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	}

	if m.loading {
		return m, nil
	}

	vp := m.viewportHeight()
	switch msg.String() {
	case "j", "down":
		m.scrollOffset = min(m.scrollOffset+1, m.maxScroll())
	case "k", "up":
		m.scrollOffset = max(m.scrollOffset-1, 0)
	case "ctrl+d":
		m.scrollOffset = min(m.scrollOffset+vp/2, m.maxScroll())
	case "ctrl+u":
		m.scrollOffset = max(m.scrollOffset-vp/2, 0)
	case " ", "pgdown", "ctrl+f":
		m.scrollOffset = min(m.scrollOffset+vp, m.maxScroll())
	case "pgup", "ctrl+b":
		m.scrollOffset = max(m.scrollOffset-vp, 0)
	case "g", "home":
		m.scrollOffset = 0
	case "G", "end":
		m.scrollOffset = m.maxScroll()
	case "s":
		if m.issue != nil && len(m.states) > 0 {
			m.statePicker = NewStatePicker(
				m.issue.IDReadable, m.issue.Summary,
				m.issue.Field("State"), m.states,
			)
			m.mode = modeStatePicker
		}
	case "r":
		m.loading = true
		return m, loadIssueCmd(m.client, m.issueID)
	}

	return m, nil
}

func (m IssueViewer) updateStatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.statePicker.Update(msg)
	m.statePicker = updated.(StatePicker)

	result := m.statePicker.Result()
	if !result.Cancelled && result.State == "" {
		return m, cmd
	}

	m.mode = modeNormal
	if !result.Cancelled && result.State != m.issue.Field("State") {
		m.loading = true
		return m, setStateCmd(m.client, m.issueID, result.State)
	}
	return m, nil
}

func (m IssueViewer) View() string {
	if m.mode == modeStatePicker {
		return m.statePicker.View()
	}
	if m.err != nil {
		return fmt.Sprintf("error: %v\n\npress q to quit\n", m.err)
	}
	if m.loading || m.issue == nil {
		return "Loading…\n"
	}

	var b strings.Builder
	b.WriteString(m.renderHeader())

	vpHeight := m.viewportHeight()
	start := min(m.scrollOffset, len(m.lines))
	end := min(start+vpHeight, len(m.lines))

	for i := start; i < end; i++ {
		b.WriteString(m.lines[i])
		b.WriteString("\n")
	}
	for range vpHeight - (end - start) {
		b.WriteString("\n")
	}

	b.WriteString(m.renderFooter())
	return b.String()
}

func (m IssueViewer) viewportHeight() int {
	return max(m.height-m.headerHeight-1, 1)
}

func (m IssueViewer) maxScroll() int {
	return max(len(m.lines)-m.viewportHeight(), 0)
}

// renderHeader builds the fixed header: ID, summary, metadata grid, rule.
func (m IssueViewer) renderHeader() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s  %s\n", format.StyleID.Render(m.issue.IDReadable), format.StyleBold.Render(m.issue.Summary))

	state := m.issue.Field("State")
	priority := m.issue.Field("Priority")
	assignee := m.issue.Field("Assignee")
	typ := m.issue.Field("Type")
	subsystem := m.issue.Field("Subsystem")
	tags := m.issue.TagNames()

	if state+priority+assignee+typ+subsystem+tags != "" {
		b.WriteString("\n")
		colWidth := max(m.width/2, 30)

		stateVal := lipgloss.NewStyle().Foreground(format.StateColor(state)).Render(state)
		prioVal := lipgloss.NewStyle().Foreground(format.PriorityColor(priority)).Render(priority)
		tagsVal := format.StyleDim.Render(tags)

		writeMetaRow(&b, colWidth, "State", stateVal, "Priority", prioVal)
		writeMetaRow(&b, colWidth, "Assignee", assignee, "Type", typ)
		writeMetaRow(&b, colWidth, "Subsystem", subsystem, "Tags", tagsVal)
	}

	b.WriteString("\n")
	ruleWidth := max(m.width, 40)
	b.WriteString(format.StyleRule.Render(strings.Repeat("─", ruleWidth)))
	b.WriteString("\n")

	return b.String()
}

func writeMetaRow(b *strings.Builder, colWidth int, label1, val1, label2, val2 string) {
	if val1 == "" && val2 == "" {
		return
	}
	left := fmt.Sprintf("  %s %s", format.StyleLabel.Render(label1), val1)
	if val2 != "" {
		right := fmt.Sprintf("  %s %s", format.StyleLabel.Render(label2), val2)
		fmt.Fprintf(b, "%s%s\n", lipgloss.NewStyle().Width(colWidth).Render(left), right)
		return
	}
	fmt.Fprintf(b, "%s\n", left)
}

// buildLines renders description + comments into scrollable lines.
func (m IssueViewer) buildLines() []string {
	if m.issue == nil {
		return nil
	}

	w := max(m.width, 20)
	border := format.StyleDim.Render("│")

	var lines []string

	desc := m.issue.Desc()
	if desc != "" {
		for _, line := range format.SplitRendered(format.RenderMarkdown(desc, w-2)) {
			lines = append(lines, "  "+line)
		}
	} else {
		lines = append(lines, "  "+format.StyleDim.Render("No description."))
	}

	if len(m.comments) == 0 {
		return lines
	}

	lines = append(lines, "")
	header := fmt.Sprintf("── Comments (%d) ", len(m.comments))
	if pad := w - 4 - lipgloss.Width(header); pad > 0 {
		header += strings.Repeat("─", pad)
	}
	lines = append(lines, "  "+format.StyleDim.Render(header))

	for _, c := range m.comments {
		lines = append(lines, "")
		author := commentAuthor(c)
		date := time.UnixMilli(c.Created).Format("2006-01-02 15:04")
		lines = append(lines, fmt.Sprintf("  %s %s  %s",
			border,
			format.StyleBold.Render(author),
			format.StyleDim.Render(date)))

		for _, line := range format.SplitRendered(format.RenderMarkdown(c.Text, w-5)) {
			lines = append(lines, "  "+border+" "+line)
		}
	}

	return lines
}

func commentAuthor(c youtrack.Comment) string {
	if c.Author == nil {
		return "Unknown"
	}
	if c.Author.FullName != "" {
		return c.Author.FullName
	}
	return c.Author.Login
}

func (m IssueViewer) renderFooter() string {
	hints := "j/k scroll  space/pgdn page  g/G top/end  s state  r refresh  q quit"
	pos := ""
	if len(m.lines) > 0 {
		pos = fmt.Sprintf("%d/%d", m.scrollOffset+1, len(m.lines))
	}
	gap := max(m.width-lipgloss.Width(hints)-lipgloss.Width(pos), 2)
	return format.StyleDim.Render(hints + strings.Repeat(" ", gap) + pos)
}

func loadIssueCmd(client youtrack.API, issueID string) tea.Cmd {
	return func() tea.Msg {
		issue, err := client.GetIssue(issueID)
		if err != nil {
			return issueLoadedMsg{err: err}
		}

		var (
			wg       sync.WaitGroup
			comments []youtrack.Comment
			states   []youtrack.StateBundleElement
		)
		wg.Add(2)
		go func() { defer wg.Done(); comments, _ = client.ListComments(issueID) }()
		go func() { defer wg.Done(); states, _ = client.GetIssueStates(issueID) }()
		wg.Wait()

		return issueLoadedMsg{issue: issue, comments: comments, states: states}
	}
}

func setStateCmd(client youtrack.API, issueID, state string) tea.Cmd {
	return func() tea.Msg {
		return stateChangedMsg{err: client.SetIssueState(issueID, state)}
	}
}
