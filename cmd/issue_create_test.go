package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func runCreateJSON(t *testing.T, api youtrack.API, args ...string) (youtrack.Issue, error) {
	t.Helper()
	run := setupTest(t, api)
	out, err := run(append([]string{"issue", "create", "-p", "AX", "-s", "Probe", "--json"}, args...)...)
	var issue youtrack.Issue
	if out != "" {
		if jerr := json.Unmarshal([]byte(out), &issue); jerr != nil {
			t.Fatalf("output is not JSON: %v\n%s", jerr, out)
		}
	}
	return issue, err
}

func TestRunIssueCreateExistingMultiWordTag(t *testing.T) {
	fake, api := newFakeServer(t, "Investigation", "To be discussed")

	issue, err := runCreateJSON(t, api, "-t", "to be discussed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.created) != 0 {
		t.Errorf("created tags %q, want none", fake.created)
	}
	if got := string(fake.createBody["tags"]); got != `[{"id":"6-b"}]` {
		t.Errorf("create body tags = %s, want the existing tag by ID", got)
	}
	if issue.TagNames() != "To be discussed" {
		t.Errorf("tags = %q, want To be discussed", issue.TagNames())
	}
}

func TestRunIssueCreateUnknownTag(t *testing.T) {
	fake, api := newFakeServer(t, "Investigation")

	_, err := runCreateJSON(t, api, "-t", "Investigaton")
	if err == nil || !strings.Contains(err.Error(), `no tag "Investigaton" (closest: "Investigation"); --create-tag`) {
		t.Fatalf("err = %v, want unknown tag", err)
	}
	if fake.createCalls != 0 || len(fake.created) != 0 {
		t.Errorf("create called %d times, created tags %q, want neither", fake.createCalls, fake.created)
	}

	issue, err := runCreateJSON(t, api, "-t", "Investigation", "-t", "needs review", "--create-tag")
	if err != nil {
		t.Fatal(err)
	}
	// The new tag is created and applied once the issue exists.
	if got := string(fake.createBody["tags"]); got != `[{"id":"6-a"}]` {
		t.Errorf("create body tags = %s", got)
	}
	if issue.TagNames() != "Investigation, needs review" {
		t.Errorf("tags = %q, want Investigation, needs review", issue.TagNames())
	}
}

func TestRunIssueCreateTagNotApplied(t *testing.T) {
	fake, api := newFakeServer(t, "To be discussed")
	fake.dropTags = true

	_, err := runCreateJSON(t, api, "-t", "To be discussed")
	if err == nil {
		t.Fatal("expected the missing tag to fail the create")
	}
	msg := formatError(err)
	for _, want := range []string{"AX-1 was created", `"To be discussed" (6-a) is not on AX-1`} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q", msg, want)
		}
	}
}

func TestRunIssueCreateRejectedCreatesNoTag(t *testing.T) {
	fake, api := newFakeServer(t)
	fake.rejectField = "Type"

	if _, err := runCreateJSON(t, api, "--type", "Bug", "-t", "needs review", "--create-tag"); err == nil {
		t.Fatal("expected the server's rejection")
	}
	if len(fake.created) != 0 {
		t.Errorf("created tags %q for a rejected issue", fake.created)
	}
}

func TestRunIssueCreateMultiWordTypeInCreateBody(t *testing.T) {
	fake, api := newFakeServer(t)

	if _, err := runCreateJSON(t, api, "--type", "user story"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `[{"$type":"SingleEnumIssueCustomField","name":"Type","value":{"name":"User Story"}}]`
	if got := string(fake.createBody["customFields"]); got != want {
		t.Errorf("customFields = %s, want %s", got, want)
	}
}

func TestRunIssueCreateRejectedFieldCreatesNothing(t *testing.T) {
	fake, api := newFakeServer(t, "Investigation")
	fake.rejectField = "Type"

	_, err := runCreateJSON(t, api, "--type", "Bug", "-t", "Investigation")
	if err == nil {
		t.Fatal("expected the server's rejection")
	}
	if !strings.Contains(err.Error(), "Bad value for field Type") {
		t.Errorf("err = %v", err)
	}
	if fake.createCalls != 1 {
		t.Errorf("create called %d times, want 1", fake.createCalls)
	}
	if strings.Contains(err.Error(), "was created") {
		t.Errorf("error claims an issue exists: %v", err)
	}
}

func TestRunIssueCreateUnknownValueCreatesNothing(t *testing.T) {
	fake, api := newFakeServer(t)

	_, err := runCreateJSON(t, api, "--type", "Epic")
	if err == nil || !strings.Contains(err.Error(), `Type has no value "Epic"`) {
		t.Fatalf("err = %v, want unknown value", err)
	}
	if fake.createCalls != 0 {
		t.Errorf("create called %d times, want 0", fake.createCalls)
	}
}

func TestRunIssueCreateFollowUpFailureKeepsIssue(t *testing.T) {
	mock := &mockAPI{
		issue:     &youtrack.Issue{IDReadable: "PROJ-999", Summary: "Test"},
		board:     testBoard(),
		sprintErr: fmt.Errorf("boom"),
	}
	run := setupTest(t, mock)

	out, err := run("issue", "create", "-p", "AX", "-s", "Test", "--board", "AllTix", "--json")
	if err == nil {
		t.Fatal("expected error when board placement fails")
	}
	if !strings.Contains(out, "PROJ-999") {
		t.Errorf("created issue not printed: %s", out)
	}
	msg := formatError(err)
	for _, want := range []string{"PROJ-999 was created", "yt issue update PROJ-999", "boom"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q", msg, want)
		}
	}
}

func TestRunIssueCreateInvalidFieldCreatesNothing(t *testing.T) {
	mock := &mockAPI{}
	run := setupTest(t, mock)

	_, err := run("issue", "create", "-p", "PROJ", "-s", "Test", "-d", "body", "--field", "bad-format")
	if err == nil || !strings.Contains(err.Error(), "invalid --field format") {
		t.Fatalf("err = %v, want invalid --field format", err)
	}
	if mock.createCalled {
		t.Error("issue was created despite the invalid --field")
	}
}
