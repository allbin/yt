package cmd

import "testing"

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name     string
		state    string
		assignee string
		priority string
		typ      string
		want     string
	}{
		{"empty", "", "", "", "", ""},
		{"state_single", "Open", "", "", "", "State Open"},
		{"state_multi_word", "In Progress", "", "", "", "State {In Progress}"},
		{"assignee", "", "john", "", "", "Assignee john"},
		{"assignee_me", "", "me", "", "", "Assignee me"},
		{"priority", "", "", "Critical", "", "Priority Critical"},
		{"type", "", "", "", "Bug", "Type Bug"},
		{"type_multi_word", "", "", "", "Feature Request", "Type {Feature Request}"},
		{"all", "Open", "john", "Critical", "Bug", "State Open Assignee john Priority Critical Type Bug"},
		{"state_and_assignee", "In Progress", "me", "", "", "State {In Progress} Assignee me"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCommand(tt.state, tt.assignee, tt.priority, tt.typ)
			if got != tt.want {
				t.Errorf("buildCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
