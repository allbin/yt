package cmd

import "testing"

func TestBuildCommand(t *testing.T) {
	tests := []struct {
		name       string
		state      string
		assignee   string
		priority   string
		typ        string
		tags       []string
		removeTags []string
		want       string
	}{
		{"empty", "", "", "", "", nil, nil, ""},
		{"state_single", "Open", "", "", "", nil, nil, "State Open"},
		{"state_multi_word", "In Progress", "", "", "", nil, nil, "State {In Progress}"},
		{"assignee", "", "john", "", "", nil, nil, "Assignee john"},
		{"assignee_me", "", "me", "", "", nil, nil, "Assignee me"},
		{"priority", "", "", "Critical", "", nil, nil, "Priority Critical"},
		{"type", "", "", "", "Bug", nil, nil, "Type Bug"},
		{"type_multi_word", "", "", "", "Feature Request", nil, nil, "Type {Feature Request}"},
		{"all", "Open", "john", "Critical", "Bug", nil, nil, "State Open Assignee john Priority Critical Type Bug"},
		{"state_and_assignee", "In Progress", "me", "", "", nil, nil, "State {In Progress} Assignee me"},
		{"single_tag", "", "", "", "", []string{"tech-debt"}, nil, "tag tech-debt"},
		{"multi_tag", "", "", "", "", []string{"tech-debt", "scheduler"}, nil, "tag tech-debt tag scheduler"},
		{"remove_tag", "", "", "", "", nil, []string{"obsolete"}, "untag obsolete"},
		{"tag_and_remove", "", "", "", "", []string{"new-tag"}, []string{"old-tag"}, "tag new-tag untag old-tag"},
		{"tag_multi_word", "", "", "", "", []string{"needs review"}, nil, "tag {needs review}"},
		{"state_with_tags", "Open", "", "", "", []string{"tech-debt"}, nil, "State Open tag tech-debt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCommand(tt.state, tt.assignee, tt.priority, tt.typ, tt.tags, tt.removeTags)
			if got != tt.want {
				t.Errorf("buildCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
