package youtrack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// issueFieldsJSON is the shape /api/issues/{id}/customFields returns for
// projectCustomField(field(name,fieldType(id,isMultiValue)),bundle(...)).
const issueFieldsJSON = `[
 {"projectCustomField":{"field":{"name":"Type","fieldType":{"id":"enum[1]","isMultiValue":false}},
  "bundle":{"$type":"EnumBundle","id":"94-1","name":"Types","values":[{"name":"User Story","ordinal":2},{"name":"Bug","ordinal":0}]}}},
 {"projectCustomField":{"field":{"name":"Subsystem","fieldType":{"id":"ownedField[*]","isMultiValue":true}},
  "bundle":{"$type":"OwnedBundle","id":"116-5","name":"Mobilix: Subsystem","values":[{"name":"API","ordinal":0}]}}},
 {"projectCustomField":{"field":{"name":"Due Date","fieldType":{"id":"date","isMultiValue":false}}}},
 {"projectCustomField":{"field":{"name":"Created at","fieldType":{"id":"date and time","isMultiValue":false}}}}
]`

func TestListIssueFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/HK-1/customFields" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, issueFieldsJSON)
	}))
	defer srv.Close()

	fields, err := NewClient(srv.URL, "token").ListIssueFields("HK-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 4 {
		t.Fatalf("got %d fields, want 4", len(fields))
	}
	typ, sub, due, at := fields[0], fields[1], fields[2], fields[3]
	if typ.ValueType != "enum" || typ.Multi || typ.Type != "enum" || typ.BundleType != "EnumBundle" || typ.BundleID != "94-1" {
		t.Errorf("Type = %+v", typ)
	}
	if typ.Values[0].Name != "Bug" {
		t.Errorf("values not sorted by ordinal: %+v", typ.Values)
	}
	if sub.ValueType != "ownedField" || !sub.Multi || sub.Type != "owned[]" || sub.Bundle != "Mobilix: Subsystem" {
		t.Errorf("Subsystem = %+v", sub)
	}
	if due.ValueType != "date" || at.ValueType != "date and time" {
		t.Errorf("date types = %q, %q", due.ValueType, at.ValueType)
	}
}

func payload(t *testing.T, u FieldUpdate) string {
	t.Helper()
	b, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestFieldUpdatePayloads(t *testing.T) {
	due := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	tests := []struct {
		name   string
		field  ProjectField
		values []string
		want   string
	}{
		{"enum", ProjectField{Name: "Type", ValueType: "enum"}, []string{"User Story"},
			`{"$type":"SingleEnumIssueCustomField","name":"Type","value":{"name":"User Story"}}`},
		{"enum_multi", ProjectField{Name: "Customer", ValueType: "enum", Multi: true}, []string{"LTS", "VL"},
			`{"$type":"MultiEnumIssueCustomField","name":"Customer","value":[{"name":"LTS"},{"name":"VL"}]}`},
		{"owned_multi_clear", ProjectField{Name: "Subsystem", ValueType: "ownedField", Multi: true}, []string{""},
			`{"$type":"MultiOwnedIssueCustomField","name":"Subsystem","value":[]}`},
		{"version", ProjectField{Name: "Milestone", ValueType: "version"}, []string{"MX2024-03"},
			`{"$type":"SingleVersionIssueCustomField","name":"Milestone","value":{"name":"MX2024-03"}}`},
		{"user", ProjectField{Name: "Assignee", ValueType: "user"}, []string{"jdoe"},
			`{"$type":"SingleUserIssueCustomField","name":"Assignee","value":{"login":"jdoe"}}`},
		{"user_clear", ProjectField{Name: "Assignee", ValueType: "user"}, nil,
			`{"$type":"SingleUserIssueCustomField","name":"Assignee","value":null}`},
		{"state", ProjectField{Name: "State", ValueType: "state"}, []string{"In Progress"},
			`{"$type":"StateIssueCustomField","name":"State","value":{"name":"In Progress"}}`},
		{"period", ProjectField{Name: "Estimation", ValueType: "period"}, []string{"2d 4h"},
			`{"$type":"PeriodIssueCustomField","name":"Estimation","value":{"presentation":"2d 4h"}}`},
		{"text", ProjectField{Name: "Notes", ValueType: "text"}, []string{"long text"},
			`{"$type":"TextIssueCustomField","name":"Notes","value":{"text":"long text"}}`},
		{"string", ProjectField{Name: "Ref", ValueType: "string"}, []string{"a b"},
			`{"$type":"SimpleIssueCustomField","name":"Ref","value":"a b"}`},
		{"integer", ProjectField{Name: "Points", ValueType: "integer"}, []string{"5"},
			`{"$type":"SimpleIssueCustomField","name":"Points","value":5}`},
		{"float", ProjectField{Name: "Ratio", ValueType: "float"}, []string{"0.5"},
			`{"$type":"SimpleIssueCustomField","name":"Ratio","value":0.5}`},
		{"date_and_time", ProjectField{Name: "Starts", ValueType: "date and time"}, []string{"2026-10-01T09:00:00Z"},
			`{"$type":"SimpleIssueCustomField","name":"Starts","value":` + jsonInt(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC).UnixMilli()) + `}`},
		{"date", ProjectField{Name: "Due Date", ValueType: "date"}, []string{"2026-10-01"},
			`{"$type":"DateIssueCustomField","name":"Due Date","value":` + jsonInt(due) + `}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := tt.field.Update(tt.values)
			if err != nil {
				t.Fatal(err)
			}
			if got := payload(t, u); got != tt.want {
				t.Errorf("payload =\n  %s\nwant\n  %s", got, tt.want)
			}
		})
	}
}

func jsonInt(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func TestFieldUpdateErrors(t *testing.T) {
	enum := ProjectField{Name: "Type", ValueType: "enum", Values: []BundleValue{{Name: "Bug"}, {Name: "Task"}}}
	tests := []struct {
		name    string
		field   ProjectField
		values  []string
		wantErr string
	}{
		{"unknown_value", enum, []string{"Story"}, `Type has no value "Story"; allowed: Bug, Task`},
		{"single_given_two", enum, []string{"Bug", "Task"}, "Type takes a single value, got 2"},
		{"bad_integer", ProjectField{Name: "Points", ValueType: "integer"}, []string{"five"}, "expects an integer"},
		{"nan_float", ProjectField{Name: "Ratio", ValueType: "float"}, []string{"NaN"}, "expects a number"},
		{"inf_float", ProjectField{Name: "Ratio", ValueType: "float"}, []string{"+Inf"}, "expects a number"},
		{"bad_date", ProjectField{Name: "Due", ValueType: "date"}, []string{"tomorrow"}, "invalid date"},
		{"unsupported", ProjectField{Name: "X", ValueType: "mystery"}, []string{"v"}, "not supported"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.field.Update(tt.values)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("err = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

func TestCreateIssueSendsCustomFields(t *testing.T) {
	var body map[string]json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		_, _ = io.WriteString(w, `{"idReadable":"HK-9"}`)
	}))
	defer srv.Close()

	u, err := ProjectField{Name: "Type", ValueType: "enum"}.Update([]string{"User Story"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewClient(srv.URL, "token").CreateIssue("HK", "S", "", []FieldUpdate{u}); err != nil {
		t.Fatal(err)
	}
	want := `[{"$type":"SingleEnumIssueCustomField","name":"Type","value":{"name":"User Story"}}]`
	if string(body["customFields"]) != want {
		t.Errorf("customFields = %s, want %s", body["customFields"], want)
	}
}

func TestUpdateIssueFieldsOneRequest(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
	}))
	defer srv.Close()

	u := FieldUpdate{Type: "SingleEnumIssueCustomField", Name: "Type", Value: namedValue{Name: "Bug"}}
	if err := NewClient(srv.URL, "token").UpdateIssueFields("HK-1", map[string]string{"summary": "S"}, []FieldUpdate{u}); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/issues/HK-1" {
		t.Errorf("path = %s", gotPath)
	}
	if want := `{"customFields":[{"$type":"SingleEnumIssueCustomField","name":"Type","value":{"name":"Bug"}}],"summary":"S"}`; gotBody != want {
		t.Errorf("body = %s, want %s", gotBody, want)
	}
}

func TestAddBundleValue(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	f := ProjectField{Name: "Customer", BundleType: "EnumBundle", BundleID: "94-19"}
	if err := client.AddBundleValue(f, "LTVB"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/admin/customFieldSettings/bundles/enum/94-19/values" {
		t.Errorf("path = %s", gotPath)
	}
	if gotBody != `{"name":"LTVB"}` {
		t.Errorf("body = %s", gotBody)
	}

	user := ProjectField{Name: "Assignee", Type: "user", BundleType: "UserBundle", BundleID: "9-19"}
	if err := client.AddBundleValue(user, "x"); err == nil {
		t.Error("expected error adding to a user bundle")
	}
}

func TestUploadAttachments(t *testing.T) {
	got := map[string]string{}
	types := map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issues/HK-1/attachments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		mr, err := r.MultipartReader()
		if err != nil {
			t.Fatal(err)
		}
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			b, _ := io.ReadAll(part)
			got[part.FileName()] = string(b)
			types[part.FileName()] = part.Header.Get("Content-Type")
		}
		_, _ = io.WriteString(w, `[{"id":"1","name":"a.png","size":3},{"id":"2","name":"b.txt","size":5}]`)
	}))
	defer srv.Close()

	atts, err := NewClient(srv.URL, "token").UploadAttachments("HK-1", []UploadFile{
		{Name: "a.png", Content: strings.NewReader("png")},
		{Name: "b.txt", Content: strings.NewReader("hello")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got["a.png"] != "png" || got["b.txt"] != "hello" {
		t.Errorf("server received %v", got)
	}
	if types["a.png"] != "image/png" || !strings.HasPrefix(types["b.txt"], "text/plain") {
		t.Errorf("content types = %v", types)
	}
	if len(atts) != 2 || atts[1].Name != "b.txt" {
		t.Errorf("attachments = %+v", atts)
	}
}

func TestListProjectFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/admin/projects/HK/customFields" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_, _ = io.WriteString(w, `[
 {"field":{"name":"Customer","fieldType":{"id":"enum[*]","isMultiValue":true}},
  "bundle":{"$type":"EnumBundle","id":"94-19","name":"Mobilix: Customer","values":[{"name":"VL","ordinal":1},{"name":"LTS","ordinal":0}]}},
 {"field":{"name":"Estimation","fieldType":{"id":"period","isMultiValue":false}}}
]`)
	}))
	defer srv.Close()

	fields, err := NewClient(srv.URL, "token").ListProjectFields("HK")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 {
		t.Fatalf("got %d fields", len(fields))
	}
	c, e := fields[0], fields[1]
	if c.Type != "enum[]" || !c.Multi || c.BundleID != "94-19" || c.BundleType != "EnumBundle" || c.Values[0].Name != "LTS" {
		t.Errorf("Customer = %+v", c)
	}
	if e.Type != "period" || e.BundleID != "" {
		t.Errorf("Estimation = %+v", e)
	}
}

func TestParseDateTime(t *testing.T) {
	local := time.Date(2026, 10, 1, 9, 30, 0, 0, time.Local).UnixMilli()
	tests := []struct {
		in   string
		want int64
	}{
		{"2026-10-01 09:30", local},
		{"2026-10-01T09:30:00Z", time.Date(2026, 10, 1, 9, 30, 0, 0, time.UTC).UnixMilli()},
		{"2026-10-01", time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local).UnixMilli()},
	}
	for _, tt := range tests {
		got, err := parseDate(tt.in, false)
		if err != nil || got != tt.want {
			t.Errorf("parseDate(%q) = %d, %v; want %d", tt.in, got, err, tt.want)
		}
	}
	if _, err := parseDate("2026-10-01 09:30", true); err == nil {
		t.Error("date-only field accepted a time")
	}
}

func TestCustomFieldValues(t *testing.T) {
	tests := []struct {
		raw  string
		want []string
	}{
		{`[{"name":"API"},{"name":"Management UI"}]`, []string{"API", "Management UI"}},
		{`[{"name":"Jane Doe","login":"jdoe"}]`, []string{"jdoe"}},
		{`{"name":"Bug"}`, []string{"Bug"}},
		{`[]`, nil},
		{`null`, nil},
		{`"text"`, nil},
	}
	for _, tt := range tests {
		cf := CustomField{Value: json.RawMessage(tt.raw)}
		got := cf.Values()
		if strings.Join(got, "|") != strings.Join(tt.want, "|") || len(got) != len(tt.want) {
			t.Errorf("Values(%s) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}
