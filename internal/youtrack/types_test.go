package youtrack

import (
	"encoding/json"
	"testing"
)

func TestCustomFieldDisplayValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"null", "null", ""},
		{"empty", "", ""},
		{"enum", `{"name":"Open"}`, "Open"},
		{"user_fullname", `{"fullName":"John Doe","login":"john"}`, "John Doe"},
		{"user_login_only", `{"login":"john"}`, "john"},
		{"presentation", `{"presentation":"2h 30m"}`, "2h 30m"},
		{"text", `{"text":"some notes"}`, "some notes"},
		{"array", `[{"name":"Backend"},{"name":"Frontend"}]`, "Backend, Frontend"},
		{"array_single", `[{"name":"API"}]`, "API"},
		{"empty_array", `[]`, ""},
		{"plain_string", `"hello"`, "hello"},
		{"number", `42`, ""},
		{"object_empty_fields", `{"$type":"EnumBundleElement"}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := CustomField{Name: "test", Value: json.RawMessage(tt.value)}
			got := cf.DisplayValue()
			if got != tt.want {
				t.Errorf("DisplayValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIssueField(t *testing.T) {
	issue := Issue{
		CustomFields: []CustomField{
			{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
			{Name: "Priority", Value: json.RawMessage(`{"name":"Major"}`)},
		},
	}

	if got := issue.Field("State"); got != "Open" {
		t.Errorf("Field(State) = %q, want Open", got)
	}
	if got := issue.Field("Priority"); got != "Major" {
		t.Errorf("Field(Priority) = %q, want Major", got)
	}
	if got := issue.Field("Missing"); got != "" {
		t.Errorf("Field(Missing) = %q, want empty", got)
	}
}

func TestIssueTagNames(t *testing.T) {
	tests := []struct {
		name string
		tags []Tag
		want string
	}{
		{"nil", nil, ""},
		{"empty", []Tag{}, ""},
		{"single", []Tag{{Name: "urgent"}}, "urgent"},
		{"multiple", []Tag{{Name: "urgent"}, {Name: "backend"}}, "urgent, backend"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := Issue{Tags: tt.tags}
			if got := issue.TagNames(); got != tt.want {
				t.Errorf("TagNames() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIssueDesc(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		issue := Issue{}
		if got := issue.Desc(); got != "" {
			t.Errorf("Desc() = %q, want empty", got)
		}
	})
	t.Run("non_nil", func(t *testing.T) {
		desc := "hello"
		issue := Issue{Description: &desc}
		if got := issue.Desc(); got != "hello" {
			t.Errorf("Desc() = %q, want hello", got)
		}
	})
}
