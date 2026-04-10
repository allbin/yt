package cmd

import (
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestRunProjects(t *testing.T) {
	run := setupTest(t, &mockAPI{
		projects: []youtrack.Project{
			{ShortName: "PROJ", Name: "My Project"},
			{ShortName: "DEMO", Name: "Demo Project"},
		},
	})

	out, err := run("projects")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PROJ") {
		t.Errorf("output missing PROJ: %s", out)
	}
	if !strings.Contains(out, "DEMO") {
		t.Errorf("output missing DEMO: %s", out)
	}
}

func TestRunProjectsEmpty(t *testing.T) {
	run := setupTest(t, &mockAPI{})

	out, err := run("projects")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No projects found") {
		t.Errorf("expected empty message: %s", out)
	}
}

func TestRunProjectFields(t *testing.T) {
	mock := &mockAPI{
		projectFields: []youtrack.ProjectField{
			{Name: "State", Type: "state", Values: []youtrack.BundleValue{{Name: "Open"}, {Name: "Closed"}}},
			{Name: "Subsystem", Type: "owned", Values: []youtrack.BundleValue{{Name: "API"}, {Name: "UI"}}},
		},
	}
	run := setupTest(t, mock)

	out, err := run("project", "fields", "PROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "State") {
		t.Errorf("output missing State: %s", out)
	}
	if !strings.Contains(out, "Subsystem") {
		t.Errorf("output missing Subsystem: %s", out)
	}
	if !strings.Contains(out, "API") {
		t.Errorf("output missing API value: %s", out)
	}
}

func TestRunProjectFieldsJSON(t *testing.T) {
	mock := &mockAPI{
		projectFields: []youtrack.ProjectField{
			{Name: "Priority", Type: "enum", Values: []youtrack.BundleValue{{Name: "Critical"}}},
		},
	}
	run := setupTest(t, mock)

	out, err := run("project", "fields", "PROJ", "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"name": "Priority"`) {
		t.Errorf("JSON output missing field name: %s", out)
	}
	if !strings.Contains(out, `"name": "Critical"`) {
		t.Errorf("JSON output missing value: %s", out)
	}
}
