package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allbin/yt/internal/youtrack"
)

var testStates = []youtrack.StateBundleElement{
	{Name: "Open", Ordinal: 0},
	{Name: "In Progress", Ordinal: 1},
	{Name: "Done", Ordinal: 2},
}

func TestNewStatePickerCursorOnCurrent(t *testing.T) {
	p := NewStatePicker("T-1", "Test", "In Progress", testStates)
	if p.cursor != 1 {
		t.Errorf("cursor = %d, want 1", p.cursor)
	}
	if p.current != 1 {
		t.Errorf("current = %d, want 1", p.current)
	}
}

func TestNewStatePickerUnknownState(t *testing.T) {
	p := NewStatePicker("T-1", "Test", "Nonexistent", testStates)
	if p.cursor != 0 {
		t.Errorf("cursor = %d, want 0 for unknown state", p.cursor)
	}
}

func TestStatePickerEmptyStates(t *testing.T) {
	p := NewStatePicker("T-1", "Test", "Open", nil)
	m, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	result := m.(StatePicker).Result()
	if !result.Cancelled {
		t.Error("expected cancelled for empty states")
	}
}

func TestStatePickerNavigation(t *testing.T) {
	p := NewStatePicker("T-1", "Test", "Open", testStates)

	// Move down
	m, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	p = m.(StatePicker)
	if p.cursor != 1 {
		t.Errorf("after j: cursor = %d, want 1", p.cursor)
	}

	// Move down again
	m, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	p = m.(StatePicker)
	if p.cursor != 2 {
		t.Errorf("after jj: cursor = %d, want 2", p.cursor)
	}

	// Move down at bottom — should stay
	m, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	p = m.(StatePicker)
	if p.cursor != 2 {
		t.Errorf("at bottom: cursor = %d, want 2", p.cursor)
	}

	// Move up
	m, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	p = m.(StatePicker)
	if p.cursor != 1 {
		t.Errorf("after k: cursor = %d, want 1", p.cursor)
	}
}

func TestStatePickerSelect(t *testing.T) {
	p := NewStatePicker("T-1", "Test", "Open", testStates)

	// Move to "In Progress" and select
	m, _ := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	result := m.(StatePicker).Result()
	if result.State != "In Progress" {
		t.Errorf("selected = %q, want %q", result.State, "In Progress")
	}
	if result.Cancelled {
		t.Error("should not be cancelled")
	}
}

func TestStatePickerCancel(t *testing.T) {
	p := NewStatePicker("T-1", "Test", "Open", testStates)
	m, _ := p.Update(tea.KeyMsg{Type: tea.KeyEsc})
	result := m.(StatePicker).Result()
	if !result.Cancelled {
		t.Error("expected cancelled on esc")
	}
}

func TestStatePickerViewClearsOnDone(t *testing.T) {
	p := NewStatePicker("T-1", "Test", "Open", testStates)
	m, _ := p.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := m.(StatePicker).View()
	if view != "" {
		t.Errorf("view should be empty after selection, got %q", view)
	}
}
