package cmd

import (
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestRunProjectList(t *testing.T) {
	run := setupTest(t, &mockAPI{
		projects: []youtrack.Project{
			{ShortName: "PROJ", Name: "My Project"},
			{ShortName: "DEMO", Name: "Demo Project"},
		},
	})

	out, err := run("project", "list")
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

func TestRunProjectListEmpty(t *testing.T) {
	run := setupTest(t, &mockAPI{})

	out, err := run("project", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No projects found") {
		t.Errorf("expected empty message: %s", out)
	}
}
