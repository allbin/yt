package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
)

// StateResult holds the outcome of the state picker.
type StateResult struct {
	State     string
	Cancelled bool
}

// StatePicker is a bubbletea model for interactively selecting an issue state.
type StatePicker struct {
	states  []youtrack.StateBundleElement
	cursor  int
	current int
	issueID string
	summary string
	result  StateResult
}

// NewStatePicker creates a picker with cursor on the current state.
func NewStatePicker(issueID, summary, currentState string, states []youtrack.StateBundleElement) StatePicker {
	current := 0
	for i, s := range states {
		if s.Name == currentState {
			current = i
			break
		}
	}
	return StatePicker{
		states:  states,
		cursor:  current,
		current: current,
		issueID: issueID,
		summary: summary,
	}
}

// Result returns the picker outcome after Run completes.
func (m StatePicker) Result() StateResult { return m.result }

func (m StatePicker) Init() tea.Cmd { return nil }

func (m StatePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.result = StateResult{Cancelled: true}
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.states)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.states) > 0 {
				m.result = StateResult{State: m.states[m.cursor].Name}
			} else {
				m.result = StateResult{Cancelled: true}
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m StatePicker) View() string {
	if m.result.State != "" || m.result.Cancelled {
		return ""
	}

	var b strings.Builder

	fmt.Fprintf(&b, "%s  %s\n\n", format.StyleID.Render(m.issueID), format.StyleBold.Render(m.summary))

	for i, s := range m.states {
		pointer := "  "
		if i == m.cursor {
			pointer = "▸ "
		}

		marker := "○"
		if i == m.current {
			marker = "●"
		}

		name := lipgloss.NewStyle().Foreground(format.StateColor(s.Name)).Render(s.Name)
		fmt.Fprintf(&b, "%s%s %s", pointer, marker, name)

		if i == m.current {
			fmt.Fprintf(&b, "  %s", format.StyleDim.Render("current"))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(format.StyleDim.Render("↑/k up  ↓/j down  enter select  esc cancel"))
	b.WriteString("\n")

	return b.String()
}
