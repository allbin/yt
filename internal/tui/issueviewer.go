package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/tui/modal"
	"github.com/allbin/yt/internal/youtrack"
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

	modals modal.Stack
}

// Done reports whether the viewer has been closed (always false — standalone viewer).
func (m IssueViewer) Done() bool { return false }

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
		if m.modals.Active() {
			_, cmd := m.modals.Update(wsm)
			return m, cmd
		}
		if m.issue != nil {
			m.rebuildContent()
		}
		return m, nil
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

func (m IssueViewer) handleModalResult(popped modal.Modal) tea.Cmd {
	if sp, ok := popped.(StatePicker); ok {
		r := sp.Result()
		if !r.Cancelled && r.State != m.issue.View().State {
			return setStateCmd(m.client, m.issueID, r.State)
		}
	}
	return nil
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
			v := m.issue.View()
			sp := NewStatePicker(v.ID, v.Summary, v.State, m.states)
			m.modals.Push(sp)
		}
	case "r":
		m.loading = true
		return m, loadIssueCmd(m.client, m.issueID)
	}

	return m, nil
}

func (m IssueViewer) View() string {
	if m.modals.Active() {
		return m.modals.View()
	}
	if m.err != nil {
		return fmt.Sprintf("error: %v\n\npress q to quit\n", m.err)
	}
	if m.loading || m.issue == nil {
		return "Loading\u2026\n"
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

func (m IssueViewer) renderHeader() string {
	v := m.issue.View()
	var b strings.Builder

	fmt.Fprintf(&b, "%s  %s\n", format.StyleID.Render(v.ID), format.StyleBold.Render(v.Summary))

	state := v.State
	priority := v.Priority
	assignee := v.Assignee
	typ := v.Type
	subsystem := v.Subsystem
	tags := v.Tags

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
	b.WriteString(format.StyleRule.Render(strings.Repeat("\u2500", ruleWidth)))
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

func (m IssueViewer) buildLines() []string {
	if m.issue == nil {
		return nil
	}

	w := max(m.width, 20)
	border := format.StyleDim.Render("\u2502")

	var lines []string

	desc := m.issue.View().Description
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
	header := fmt.Sprintf("\u2500\u2500 Comments (%d) ", len(m.comments))
	if pad := w - 4 - lipgloss.Width(header); pad > 0 {
		header += strings.Repeat("\u2500", pad)
	}
	lines = append(lines, "  "+format.StyleDim.Render(header))

	for _, c := range m.comments {
		lines = append(lines, "")
		cv := c.View()
		date := time.UnixMilli(cv.Created).Format("2006-01-02 15:04")
		lines = append(lines, fmt.Sprintf("  %s %s  %s",
			border,
			format.StyleBold.Render(cv.Author),
			format.StyleDim.Render(date)))

		for _, line := range format.SplitRendered(format.RenderMarkdown(c.Text, w-5)) {
			lines = append(lines, "  "+border+" "+line)
		}
	}

	return lines
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
