package youtrack

import "testing"

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		state    string
		assignee string
		query    string
		want     string
	}{
		{"empty", "", "", "", "", ""},
		{"project", "PROJ", "", "", "", "project: PROJ"},
		{"state", "", "Open", "", "", "State: {Open}"},
		{"multi_word_state", "", "In Progress", "", "", "State: {In Progress}"},
		{"assignee", "", "", "me", "", "Assignee: me"},
		{"raw_query", "", "", "", "tag: Important", "tag: Important"},
		{"all", "PROJ", "Open", "john", "sort by: updated", "project: PROJ State: {Open} Assignee: john sort by: updated"},
		{"project_and_query", "PROJ", "", "", "created: today", "project: PROJ created: today"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildQuery(tt.project, tt.state, tt.assignee, tt.query)
			if got != tt.want {
				t.Errorf("BuildQuery() = %q, want %q", got, tt.want)
			}
		})
	}
}
