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
		fields     []string
		want       string
		wantErr    bool
	}{
		{"empty", "", "", "", "", nil, nil, nil, "", false},
		{"state_single", "Open", "", "", "", nil, nil, nil, "State Open", false},
		{"state_multi_word", "In Progress", "", "", "", nil, nil, nil, "State {In Progress}", false},
		{"assignee", "", "john", "", "", nil, nil, nil, "Assignee john", false},
		{"assignee_me", "", "me", "", "", nil, nil, nil, "Assignee me", false},
		{"priority", "", "", "Critical", "", nil, nil, nil, "Priority Critical", false},
		{"type", "", "", "", "Bug", nil, nil, nil, "Type Bug", false},
		{"type_multi_word", "", "", "", "Feature Request", nil, nil, nil, "Type {Feature Request}", false},
		{"all", "Open", "john", "Critical", "Bug", nil, nil, nil, "State Open Assignee john Priority Critical Type Bug", false},
		{"state_and_assignee", "In Progress", "me", "", "", nil, nil, nil, "State {In Progress} Assignee me", false},
		{"single_tag", "", "", "", "", []string{"tech-debt"}, nil, nil, "tag tech-debt", false},
		{"multi_tag", "", "", "", "", []string{"tech-debt", "scheduler"}, nil, nil, "tag tech-debt tag scheduler", false},
		{"remove_tag", "", "", "", "", nil, []string{"obsolete"}, nil, "untag obsolete", false},
		{"tag_and_remove", "", "", "", "", []string{"new-tag"}, []string{"old-tag"}, nil, "tag new-tag untag old-tag", false},
		{"tag_multi_word", "", "", "", "", []string{"needs review"}, nil, nil, "tag {needs review}", false},
		{"state_with_tags", "Open", "", "", "", []string{"tech-debt"}, nil, nil, "State Open tag tech-debt", false},
		{"field_simple", "", "", "", "", nil, nil, []string{"Severity=Critical"}, "Severity Critical", false},
		{"field_multi_word_value", "", "", "", "", nil, nil, []string{"Severity=Show Stopper"}, "Severity {Show Stopper}", false},
		{"field_multi_word_name", "", "", "", "", nil, nil, []string{"Fix versions=2.0"}, "{Fix versions} 2.0", false},
		{"field_multiple", "", "", "", "", nil, nil, []string{"Severity=Critical", "Component=Auth"}, "Severity Critical Component Auth", false},
		{"field_with_state", "Open", "", "", "", nil, nil, []string{"Severity=Critical"}, "State Open Severity Critical", false},
		{"field_value_with_equals", "", "", "", "", nil, nil, []string{"URL=https://x.com/a=b"}, "URL https://x.com/a=b", false},
		{"field_subsystem", "", "", "", "", nil, nil, []string{"Subsystem=API"}, "Subsystem API", false},
		{"field_subsystem_multi_word", "", "", "", "", nil, nil, []string{"Subsystem=Back End"}, "Subsystem {Back End}", false},
		{"field_empty_value", "", "", "", "", nil, nil, []string{"Subsystem="}, "Subsystem ", false},
		{"field_invalid", "", "", "", "", nil, nil, []string{"bad-format"}, "", true},
		{"field_empty_name", "", "", "", "", nil, nil, []string{"=Value"}, "", true},
		{"field_whitespace_name", "", "", "", "", nil, nil, []string{"  =Value"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildCommand(tt.state, tt.assignee, tt.priority, tt.typ, tt.tags, tt.removeTags, tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}
