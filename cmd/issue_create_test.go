package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

// fakeServer mimics what issue create touches on a real YouTrack: the tag
// list, tag creation, the create body (tags by ID, fields checked), and the
// project's field schema.
type fakeServer struct {
	t           *testing.T
	tags        []youtrack.Tag // every tag the instance knows
	created     []string       // tags created through POST /api/tags
	createBody  map[string]json.RawMessage
	createCalls int
	rejectField string // when set, a create naming this field fails
	issueTags   []youtrack.Tag
}

func newFakeServer(t *testing.T, tags ...string) (*fakeServer, youtrack.API) {
	t.Helper()
	f := &fakeServer{t: t}
	for i, name := range tags {
		f.tags = append(f.tags, youtrack.Tag{ID: "6-" + string(rune('a'+i)), Name: name})
	}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	return f, youtrack.NewClient(srv.URL, "token")
}

func (f *fakeServer) serve(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/tags":
		q := strings.ToLower(r.URL.Query().Get("query"))
		out := []youtrack.Tag{}
		for _, t := range f.tags {
			if strings.Contains(strings.ToLower(t.Name), q) {
				out = append(out, t)
			}
		}
		f.write(w, out)
	case r.Method == http.MethodPost && r.URL.Path == "/api/tags":
		var body youtrack.Tag
		f.decode(r, &body)
		tag := youtrack.Tag{ID: "6-new-" + body.Name, Name: body.Name}
		f.tags = append(f.tags, tag)
		f.created = append(f.created, body.Name)
		f.write(w, tag)
	case r.Method == http.MethodGet && r.URL.Path == "/api/admin/projects/AX/customFields":
		f.write(w, []map[string]any{{
			"field": map[string]any{"name": "Type", "fieldType": map[string]any{"id": "enum[1]", "isMultiValue": false}},
			"bundle": map[string]any{"$type": "EnumBundle", "id": "b1", "name": "Types",
				"values": []map[string]any{{"name": "Bug", "ordinal": 0}, {"name": "User Story", "ordinal": 1}}},
		}})
	case r.Method == http.MethodPost && r.URL.Path == "/api/issues":
		f.createCalls++
		f.decode(r, &f.createBody)
		if f.rejectField != "" && strings.Contains(string(f.createBody["customFields"]), `"name":"`+f.rejectField+`"`) {
			f.fail(w, "Bad value for field "+f.rejectField)
			return
		}
		var refs []struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(f.createBody["tags"], &refs)
		for _, ref := range refs {
			for _, t := range f.tags {
				if t.ID == ref.ID {
					f.issueTags = append(f.issueTags, youtrack.Tag{Name: t.Name})
				}
			}
		}
		f.write(w, map[string]any{"idReadable": "AX-1", "summary": "Probe", "tags": f.issueTags})
	case r.Method == http.MethodGet && r.URL.Path == "/api/issues/AX-1":
		f.write(w, map[string]any{"idReadable": "AX-1", "summary": "Probe", "tags": f.issueTags})
	default:
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *fakeServer) decode(r *http.Request, v any) {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fakeServer) write(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fakeServer) fail(w http.ResponseWriter, description string) {
	w.WriteHeader(http.StatusBadRequest)
	f.write(w, map[string]string{"error": "bad_request", "error_description": description})
}

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

func TestRunIssueCreateNewTag(t *testing.T) {
	fake, api := newFakeServer(t, "Investigation")

	issue, err := runCreateJSON(t, api, "-t", "Investigation", "-t", "needs review")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"needs review"}; !slices.Equal(fake.created, want) {
		t.Errorf("created tags = %q, want %q", fake.created, want)
	}
	if got := string(fake.createBody["tags"]); got != `[{"id":"6-a"},{"id":"6-new-needs review"}]` {
		t.Errorf("create body tags = %s", got)
	}
	if issue.TagNames() != "Investigation, needs review" {
		t.Errorf("tags = %q, want Investigation, needs review", issue.TagNames())
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
