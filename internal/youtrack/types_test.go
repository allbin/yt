package youtrack

import (
	"encoding/json"
	"reflect"
	"strings"
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

func TestIssueFieldMap(t *testing.T) {
	issue := Issue{
		CustomFields: []CustomField{
			{Name: "State", Value: json.RawMessage(`{"name":"Planned","$type":"StateBundleElement"}`)},
			{Name: "Assignee", Value: json.RawMessage(`{"login":"jdoe","fullName":"Jane Doe"}`)},
			{Name: "Subsystem", Value: json.RawMessage(`[{"name":"Webticket"},{"name":"API"}]`)},
			{Name: "Estimation", Value: json.RawMessage(`{"presentation":"2h","minutes":120}`)},
			{Name: "Due Date", Value: json.RawMessage(`1760000000000`)},
			{Name: "Notes", Value: json.RawMessage(`{"text":"some notes"}`)},
			{Name: "Planning", Value: json.RawMessage(`null`)},
			{Name: "Empty list", Value: json.RawMessage(`[]`)},
		},
	}

	got := issue.FieldMap()
	want := map[string]any{
		"State":      "Planned",
		"Assignee":   "Jane Doe",
		"Subsystem":  []any{"Webticket", "API"},
		"Estimation": "2h",
		"Due Date":   json.Number("1760000000000"),
		"Notes":      "some notes",
		"Planning":   nil,
		"Empty list": []any{},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FieldMap() = %#v\nwant %#v", got, want)
	}
}

func TestIssueMarshalJSONAddsFields(t *testing.T) {
	desc := "body"
	issue := Issue{
		IDReadable:  "AX-1",
		Summary:     "One",
		Description: &desc,
		CustomFields: []CustomField{
			{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
			{Name: "Subsystem", Value: json.RawMessage(`[{"name":"API"}]`)},
		},
	}
	data, err := json.Marshal(&issue)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out["idReadable"] != "AX-1" || out["description"] != "body" {
		t.Errorf("raw keys lost: %v", out)
	}
	fields, _ := out["fields"].(map[string]any)
	if fields["State"] != "Open" {
		t.Errorf("fields.State = %v, want Open", fields["State"])
	}
	if sub, ok := fields["Subsystem"].([]any); !ok || len(sub) != 1 || sub[0] != "API" {
		t.Errorf("fields.Subsystem = %v, want [API]", fields["Subsystem"])
	}
	if _, ok := out["customFields"]; !ok {
		t.Error("customFields dropped from JSON")
	}

	// The map is derived: reading the JSON back ignores it.
	var back Issue
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Field("State") != "Open" {
		t.Errorf("round trip lost State: %+v", back)
	}
}

func TestIssueMarshalJSONEmptyFields(t *testing.T) {
	data, err := json.Marshal(Issue{IDReadable: "AX-1"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"fields":{}`) {
		t.Errorf("want an empty fields object, got %s", data)
	}
}

func TestAgileKind(t *testing.T) {
	sync := &struct {
		Name string `json:"name"`
	}{Name: "Sprint"}
	tests := []struct {
		name       string
		board      Agile
		kind       string
		automatic  bool
		membership string
	}{
		{"no_settings", Agile{Name: "A"}, BoardSprints, false, "issues assigned to one of the board's sprints"},
		{"sprints_explicit", Agile{SprintsSettings: &SprintsSettings{IsExplicit: true}}, BoardSprints, false, "issues assigned to one of the board's sprints"},
		{"sprints_synced", Agile{SprintsSettings: &SprintsSettings{SprintSyncField: sync}}, BoardSprints, false, "issues whose Sprint names one of the board's sprints"},
		{"query", Agile{SprintsSettings: &SprintsSettings{DisableSprints: true, ExplicitQuery: "Subsystem: Webticket "}}, BoardQuery, true, `issues matching the query "Subsystem: Webticket"`},
		{"projects", Agile{Projects: []Project{{ShortName: "AID"}, {ShortName: "KOM"}}, SprintsSettings: &SprintsSettings{DisableSprints: true}}, BoardProjects, true, "every issue of AID, KOM"},
		{"manual", Agile{SprintsSettings: &SprintsSettings{DisableSprints: true, IsExplicit: true, ExplicitQuery: "ignored"}}, BoardManual, false, "issues added to the board by hand"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.board.Kind(); got != tt.kind {
				t.Errorf("Kind() = %q, want %q", got, tt.kind)
			}
			if got := tt.board.Automatic(); got != tt.automatic {
				t.Errorf("Automatic() = %v, want %v", got, tt.automatic)
			}
			if got := tt.board.Membership(); got != tt.membership {
				t.Errorf("Membership() = %q, want %q", got, tt.membership)
			}
		})
	}
}

func TestAgileSearchQuery(t *testing.T) {
	sprints := Agile{Name: "AllTix"}
	if got := sprints.SearchQuery("2026-02 Skolmodulen 2"); got != "Board AllTix: {2026-02 Skolmodulen 2}" {
		t.Errorf("sprint board query = %q", got)
	}
	sprintless := Agile{Name: "AllTix Webticket", SprintsSettings: &SprintsSettings{DisableSprints: true, ExplicitQuery: "Subsystem: Webticket"}}
	if got := sprintless.SearchQuery(""); got != "has: {Board AllTix Webticket}" {
		t.Errorf("sprintless board query = %q", got)
	}
}
