package modal

import tea "github.com/charmbracelet/bubbletea"

// Modal is a tea.Model that signals completion via Done().
type Modal interface {
	tea.Model
	Done() bool
}

// Stack manages a LIFO stack of modals.
type Stack struct {
	layers []Modal
}

// Push adds a modal and returns its Init cmd.
func (s *Stack) Push(m Modal) tea.Cmd {
	s.layers = append(s.layers, m)
	return m.Init()
}

// Active returns true if any modal is on the stack.
func (s Stack) Active() bool { return len(s.layers) > 0 }

// Top returns the topmost modal, or nil.
func (s Stack) Top() Modal {
	if len(s.layers) == 0 {
		return nil
	}
	return s.layers[len(s.layers)-1]
}

// Update forwards msg to the top modal. If it becomes Done, pops it.
// Returns the popped modal (or nil) for result extraction via type-switch.
func (s *Stack) Update(msg tea.Msg) (Modal, tea.Cmd) {
	if len(s.layers) == 0 {
		return nil, nil
	}
	top := s.layers[len(s.layers)-1]
	updated, cmd := top.Update(msg)
	top = updated.(Modal)
	s.layers[len(s.layers)-1] = top
	if top.Done() {
		s.layers = s.layers[:len(s.layers)-1]
		// Discard cmd (typically tea.Quit) — modals signal completion
		// via Done(), and the parent handles result extraction. Passing
		// tea.Quit through would kill the entire program.
		return top, nil
	}
	return nil, cmd
}

// View returns the topmost modal's view, or "".
func (s Stack) View() string {
	if m := s.Top(); m != nil {
		return m.View()
	}
	return ""
}
