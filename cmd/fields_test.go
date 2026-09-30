package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

// hkSchema mirrors a real project's field set: single and multi enums, a
// multi-value owned field, a user field and simple types.
func hkSchema() []youtrack.ProjectField {
	values := func(names ...string) []youtrack.BundleValue {
		out := make([]youtrack.BundleValue, len(names))
		for i, n := range names {
			out[i] = youtrack.BundleValue{Name: n, Ordinal: i}
		}
		return out
	}
	return []youtrack.ProjectField{
		{Name: "Type", ValueType: "enum", Values: values("Bug", "Task", "User Story"), BundleType: "EnumBundle", BundleID: "94-1"},
		{Name: "Priority", ValueType: "enum", Values: values("Critical", "Normal"), BundleType: "EnumBundle", BundleID: "94-0"},
		{Name: "Subsystem", ValueType: "ownedField", Multi: true, Values: values("API", "Management UI"), BundleType: "OwnedBundle", BundleID: "116-5"},
		{Name: "Customer", ValueType: "enum", Multi: true, Values: values("LTS", "VL"), BundleType: "EnumBundle", BundleID: "94-19", Bundle: "Mobilix: Customer"},
		{Name: "Assignee", ValueType: "user", Values: values("Registered Users"), BundleType: "UserBundle"},
		{Name: "Note", ValueType: "string"},
	}
}

// assertFields compares updates by their JSON payload, which is what the
// server sees.
func assertFields(t *testing.T, got, want []youtrack.FieldUpdate) {
	t.Helper()
	g, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	w, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(g) != string(w) {
		t.Errorf("fields =\n  %s\nwant\n  %s", g, w)
	}
}

func named(n string) map[string]any { return map[string]any{"name": n} }

func TestRunIssueUpdateMultiWordValues(t *testing.T) {
	mock := &mockAPI{issue: &youtrack.Issue{IDReadable: "HK-1"}, issueFields: hkSchema()}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "HK-1", "-t", "User Story", "--subsystem", "Management UI")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFields(t, mock.setFields, []youtrack.FieldUpdate{
		{Type: "SingleEnumIssueCustomField", Name: "Type", Value: named("User Story")},
		{Type: "MultiOwnedIssueCustomField", Name: "Subsystem", Value: []any{named("Management UI")}},
	})
	if mock.command != "" {
		t.Errorf("fields must not go through the command API, got %q", mock.command)
	}
}

func TestRunIssueUpdateMultiValueField(t *testing.T) {
	mock := &mockAPI{issue: &youtrack.Issue{IDReadable: "HK-1"}, issueFields: hkSchema()}
	run := setupTest(t, mock)

	// Case-insensitive names and values resolve to the canonical spelling.
	_, err := run("issue", "update", "HK-1", "--field", "subsystem=api", "--field", "Subsystem=management ui")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFields(t, mock.setFields, []youtrack.FieldUpdate{
		{Type: "MultiOwnedIssueCustomField", Name: "Subsystem", Value: []any{named("API"), named("Management UI")}},
	})
}

func TestRunIssueUpdateClearField(t *testing.T) {
	mock := &mockAPI{issue: &youtrack.Issue{IDReadable: "HK-1"}, issueFields: hkSchema()}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "HK-1", "--subsystem", "", "--field", "Priority=")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFields(t, mock.setFields, []youtrack.FieldUpdate{
		{Type: "MultiOwnedIssueCustomField", Name: "Subsystem", Value: []any{}},
		{Type: "SingleEnumIssueCustomField", Name: "Priority", Value: nil},
	})
}

func TestRunIssueUpdateFieldValueKeepsCommas(t *testing.T) {
	mock := &mockAPI{issue: &youtrack.Issue{IDReadable: "HK-1"}, issueFields: hkSchema()}
	run := setupTest(t, mock)

	_, err := run("issue", "update", "HK-1", "--field", "Note=first, second")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFields(t, mock.setFields, []youtrack.FieldUpdate{
		{Type: "SimpleIssueCustomField", Name: "Note", Value: "first, second"},
	})
}

func TestRunIssueUpdateBadValueWritesNothing(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"unknown_value", []string{"-S", "New title", "-t", "Story"}, `Type has no value "Story"; allowed: Bug, Task, User Story`},
		{"unknown_field", []string{"-S", "New title", "--field", "Severity=High"}, `no field "Severity"`},
		{"two_values_single_field", []string{"--field", "Type=Bug", "--field", "Type=Task"}, "Type takes a single value"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAPI{issue: &youtrack.Issue{IDReadable: "HK-1"}, issueFields: hkSchema()}
			run := setupTest(t, mock)

			_, err := run(append([]string{"issue", "update", "HK-1"}, tt.args...)...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("err = %v, want %q", err, tt.wantErr)
			}
			if mock.updatedFields != nil || mock.setFields != nil {
				t.Errorf("wrote %v / %v despite the bad field", mock.updatedFields, mock.setFields)
			}
		})
	}
}

func TestRunIssueCreateSendsFieldsWithCreate(t *testing.T) {
	mock := &mockAPI{
		issue:         &youtrack.Issue{IDReadable: "PROJ-999"},
		projectFields: hkSchema(),
		currentUser:   &youtrack.User{Login: "jdoe"},
	}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "HK", "-s", "Test",
		"--type", "User Story", "--assignee", "me",
		"--field", "Customer=LTS", "--field", "Customer=VL")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFields(t, mock.createdFields, []youtrack.FieldUpdate{
		{Type: "SingleEnumIssueCustomField", Name: "Type", Value: named("User Story")},
		{Type: "SingleUserIssueCustomField", Name: "Assignee", Value: map[string]any{"login": "jdoe"}},
		{Type: "MultiEnumIssueCustomField", Name: "Customer", Value: []any{named("LTS"), named("VL")}},
	})
	if mock.command != "" {
		t.Errorf("fields must not go through the command API, got %q", mock.command)
	}
}

func TestRunIssueCreateBadFieldCreatesNothing(t *testing.T) {
	mock := &mockAPI{projectFields: hkSchema()}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "HK", "-s", "Test", "--field", "Customer=LTVB")
	if err == nil || !strings.Contains(err.Error(), `Customer has no value "LTVB"`) {
		t.Fatalf("err = %v, want unknown value", err)
	}
	if mock.createCalled {
		t.Error("issue was created despite the bad field value")
	}
}

func TestRunIssueCreateWithLinks(t *testing.T) {
	mock := &mockAPI{issue: &youtrack.Issue{IDReadable: "PROJ-999"}, linkTypes: mockLinkTypes}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "PROJ", "-s", "Test", "--link", "depends-on=PROJ-3", "--link", "relates=PROJ-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"PROJ-999|depends on|PROJ-3", "PROJ-999|relates to|PROJ-4"}
	if !slices.Equal(mock.createdLinks, want) {
		t.Errorf("links = %q, want %q", mock.createdLinks, want)
	}
}

func TestRunIssueCreateBadLinkCreatesNothing(t *testing.T) {
	for _, link := range []string{"depends-on", "blocks=PROJ-3"} {
		t.Run(link, func(t *testing.T) {
			mock := &mockAPI{linkTypes: mockLinkTypes}
			run := setupTest(t, mock)

			if _, err := run("issue", "create", "-p", "PROJ", "-s", "Test", "--link", link); err == nil {
				t.Fatal("expected error for bad --link")
			}
			if mock.createCalled {
				t.Error("issue was created despite the bad --link")
			}
		})
	}
}

func TestRunProjectFieldsAdd(t *testing.T) {
	mock := &mockAPI{projectFields: hkSchema()}
	run := setupTest(t, mock)

	out, err := run("project", "fields", "add", "HK", "customer", "LTVB", "lts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"Customer|LTVB"}; !slices.Equal(mock.addedBundle, want) {
		t.Errorf("added = %q, want %q", mock.addedBundle, want)
	}
	for _, want := range []string{"LTVB", "LTS (already exists)", "Mobilix: Customer"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRunProjectFieldsAddUnknownField(t *testing.T) {
	mock := &mockAPI{projectFields: hkSchema()}
	run := setupTest(t, mock)

	_, err := run("project", "fields", "add", "HK", "Severity", "High")
	if err == nil || !strings.Contains(err.Error(), `no field "Severity"`) {
		t.Fatalf("err = %v, want unknown field", err)
	}
}

func TestRunAttachmentUpload(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "shot.png"), filepath.Join(dir, "log.txt")
	for _, p := range []string{a, b} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	mock := &mockAPI{}
	run := setupTest(t, mock)

	out, err := run("attachment", "upload", "HK-1", a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"shot.png", "log.txt"}; !slices.Equal(mock.uploaded, want) {
		t.Errorf("uploaded = %q, want %q", mock.uploaded, want)
	}
	if !strings.Contains(out, "Uploaded shot.png") {
		t.Errorf("output: %s", out)
	}
}

func TestRunAttachmentUploadMissingFileUploadsNothing(t *testing.T) {
	mock := &mockAPI{}
	run := setupTest(t, mock)

	if _, err := run("attachment", "upload", "HK-1", filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("expected error for missing file")
	}
	if mock.uploaded != nil {
		t.Errorf("uploaded %q despite the missing file", mock.uploaded)
	}
}
