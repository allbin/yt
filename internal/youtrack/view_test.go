package youtrack

import (
	"encoding/json"
	"testing"
)

func TestIssueView(t *testing.T) {
	desc := "Bug description"
	resolved := int64(1700000000000)
	issue := Issue{
		IDReadable:  "PROJ-42",
		Summary:     "Fix login",
		Description: &desc,
		Resolved:    &resolved,
		Tags:        []Tag{{Name: "urgent"}, {Name: "backend"}},
		CustomFields: []CustomField{
			{Name: "State", Value: json.RawMessage(`{"name":"In Progress"}`)},
			{Name: "Priority", Value: json.RawMessage(`{"name":"Critical"}`)},
			{Name: "Assignee", Value: json.RawMessage(`{"fullName":"Alice Smith"}`)},
			{Name: "Type", Value: json.RawMessage(`{"name":"Bug"}`)},
			{Name: "Subsystem", Value: json.RawMessage(`[{"name":"API"}]`)},
		},
	}

	v := issue.View()

	checks := map[string]string{
		"ID":          "PROJ-42",
		"Summary":     "Fix login",
		"Description": "Bug description",
		"State":       "In Progress",
		"Priority":    "Critical",
		"Assignee":    "Alice Smith",
		"Type":        "Bug",
		"Subsystem":   "API",
		"Tags":        "urgent, backend",
	}
	for field, want := range checks {
		var got string
		switch field {
		case "ID":
			got = v.ID
		case "Summary":
			got = v.Summary
		case "Description":
			got = v.Description
		case "State":
			got = v.State
		case "Priority":
			got = v.Priority
		case "Assignee":
			got = v.Assignee
		case "Type":
			got = v.Type
		case "Subsystem":
			got = v.Subsystem
		case "Tags":
			got = v.Tags
		}
		if got != want {
			t.Errorf("View().%s = %q, want %q", field, got, want)
		}
	}
	if !v.IsResolved {
		t.Error("View().IsResolved = false, want true")
	}
}

func TestIssueViewDefaults(t *testing.T) {
	issue := Issue{IDReadable: "PROJ-1", Summary: "Minimal"}
	v := issue.View()

	if v.Description != "" {
		t.Errorf("Description = %q, want empty", v.Description)
	}
	if v.State != "" {
		t.Errorf("State = %q, want empty", v.State)
	}
	if v.IsResolved {
		t.Error("IsResolved = true, want false")
	}
}

func TestCommentView(t *testing.T) {
	tests := []struct {
		name    string
		comment Comment
		want    string
	}{
		{"full name", Comment{Author: &User{FullName: "Alice", Login: "alice"}, Created: 1000, Text: "hi"}, "Alice"},
		{"login only", Comment{Author: &User{Login: "bob"}, Created: 2000, Text: "ok"}, "bob"},
		{"nil author", Comment{Created: 3000, Text: "anon"}, "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.comment.View()
			if v.Author != tt.want {
				t.Errorf("Author = %q, want %q", v.Author, tt.want)
			}
			if v.Created != tt.comment.Created {
				t.Errorf("Created = %d, want %d", v.Created, tt.comment.Created)
			}
			if v.Text != tt.comment.Text {
				t.Errorf("Text = %q, want %q", v.Text, tt.comment.Text)
			}
		})
	}
}
