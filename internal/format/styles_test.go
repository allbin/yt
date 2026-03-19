package format

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStateColorSubstringCollision(t *testing.T) {
	tests := []struct {
		state string
		want  lipgloss.TerminalColor
	}{
		{"Incomplete", ColorDim},
		{"Complete", ColorGreen},
		{"Done", ColorGreen},
		{"Duplicate", ColorDim},
		{"Obsolete", ColorDim},
		{"In Review", ColorAccent},
		{"In Dev", ColorCyan},
		{"In Progress", ColorCyan},
		{"Open", ColorYellow},
		{"Submitted", ColorYellow},
		{"Planned", ColorSlate},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			if got := StateColor(tt.state); got != tt.want {
				t.Errorf("StateColor(%q) = %v, want %v", tt.state, got, tt.want)
			}
		})
	}
}

func TestPriorityColor(t *testing.T) {
	tests := []struct {
		priority string
		want     lipgloss.TerminalColor
	}{
		{"Critical", ColorRed},
		{"Show-stopper", ColorRed},
		{"Major", ColorYellow},
		{"Normal", lipgloss.NoColor{}},
		{"Minor", ColorDim},
		{"Nice to have", ColorDim},
	}
	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			if got := PriorityColor(tt.priority); got != tt.want {
				t.Errorf("PriorityColor(%q) = %v, want %v", tt.priority, got, tt.want)
			}
		})
	}
}
