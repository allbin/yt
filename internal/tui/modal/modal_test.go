package modal

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testModal struct {
	done  bool
	value string
}

func (m *testModal) Init() tea.Cmd       { return func() tea.Msg { return "init" } }
func (m *testModal) View() string        { return "modal:" + m.value }
func (m *testModal) Done() bool          { return m.done }
func (m *testModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "enter" {
		m.done = true
	}
	return m, nil
}

func TestStackEmpty(t *testing.T) {
	var s Stack
	if s.Active() {
		t.Error("empty stack should not be active")
	}
	if s.Top() != nil {
		t.Error("empty stack top should be nil")
	}
	if v := s.View(); v != "" {
		t.Errorf("empty stack view = %q", v)
	}
	popped, _ := s.Update(tea.KeyMsg{})
	if popped != nil {
		t.Error("update on empty stack should return nil")
	}
}

func TestStackPushAndView(t *testing.T) {
	var s Stack
	cmd := s.Push(&testModal{value: "hello"})
	if cmd == nil {
		t.Error("Push should return Init cmd")
	}
	if !s.Active() {
		t.Error("stack should be active after push")
	}
	if v := s.View(); v != "modal:hello" {
		t.Errorf("view = %q, want modal:hello", v)
	}
}

func TestStackUpdateAndPop(t *testing.T) {
	var s Stack
	s.Push(&testModal{value: "picker"})

	// Not done yet
	popped, _ := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if popped != nil {
		t.Error("should not pop on non-enter key")
	}
	if !s.Active() {
		t.Error("stack should still be active")
	}

	// Done
	popped, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if popped == nil {
		t.Fatal("should pop on enter")
	}
	if s.Active() {
		t.Error("stack should be empty after pop")
	}
}

func TestStackNested(t *testing.T) {
	var s Stack
	s.Push(&testModal{value: "first"})
	s.Push(&testModal{value: "second"})

	if v := s.View(); v != "modal:second" {
		t.Errorf("view = %q, want modal:second", v)
	}

	// Pop second
	popped, _ := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if popped == nil {
		t.Fatal("expected pop")
	}
	tm := popped.(*testModal)
	if tm.value != "second" {
		t.Errorf("popped value = %q, want second", tm.value)
	}

	// First is still there
	if !s.Active() {
		t.Error("stack should still have first modal")
	}
	if v := s.View(); v != "modal:first" {
		t.Errorf("view = %q, want modal:first", v)
	}
}
