package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

// fakeTagServer mimics what issue create touches on a real YouTrack: the
// create body cannot reference a tag by name, and the tag command creates a
// tag that does not exist yet.
type fakeTagServer struct {
	t          *testing.T
	tags       []string // every tag the instance knows
	created    []string // tags the tag command had to create
	issueTags  []string
	commands   []string
	commandErr string // when set, /api/commands fails with this description
}

func newFakeTagServer(t *testing.T, tags ...string) (*fakeTagServer, youtrack.API) {
	t.Helper()
	f := &fakeTagServer{t: t, tags: tags}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	return f, youtrack.NewClient(srv.URL, "token")
}

func (f *fakeTagServer) serve(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/api/issues":
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			f.t.Fatal(err)
		}
		if _, ok := body["tags"]; ok {
			f.fail(w, "YouTrack is unable to locate an Tag-type entity unless its ID is also provided")
			return
		}
		f.write(w, map[string]any{"idReadable": "AX-1", "summary": body["summary"]})
	case r.Method == http.MethodPost && r.URL.Path == "/api/commands":
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			f.t.Fatal(err)
		}
		f.commands = append(f.commands, body.Query)
		if f.commandErr != "" {
			f.fail(w, f.commandErr)
			return
		}
		for name := range strings.SplitSeq(strings.TrimPrefix(body.Query, "tag "), " tag ") {
			if !slices.Contains(f.tags, name) {
				f.tags = append(f.tags, name)
				f.created = append(f.created, name)
			}
			f.issueTags = append(f.issueTags, name)
		}
		f.write(w, map[string]any{})
	case r.Method == http.MethodGet && r.URL.Path == "/api/issues/AX-1":
		tags := []map[string]string{}
		for _, name := range f.issueTags {
			tags = append(tags, map[string]string{"name": name})
		}
		f.write(w, map[string]any{"idReadable": "AX-1", "summary": "Probe", "tags": tags})
	default:
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *fakeTagServer) write(w http.ResponseWriter, v any) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fakeTagServer) fail(w http.ResponseWriter, description string) {
	w.WriteHeader(http.StatusBadRequest)
	f.write(w, map[string]string{"error": "bad_request", "error_description": description})
}

func runCreateJSON(t *testing.T, api youtrack.API, args ...string) (youtrack.Issue, error) {
	t.Helper()
	run := setupTest(t, api)
	out, err := run(append([]string{"issue", "create", "-p", "AX", "-s", "Probe", "--json"}, args...)...)
	var issue youtrack.Issue
	if jerr := json.Unmarshal([]byte(out), &issue); jerr != nil {
		t.Fatalf("output is not JSON: %v\n%s", jerr, out)
	}
	return issue, err
}

func TestRunIssueCreateExistingTag(t *testing.T) {
	fake, api := newFakeTagServer(t, "Investigation")

	issue, err := runCreateJSON(t, api, "-t", "Investigation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"tag Investigation"}; !slices.Equal(fake.commands, want) {
		t.Errorf("commands = %q, want %q", fake.commands, want)
	}
	if len(fake.created) != 0 {
		t.Errorf("created tags %q, want none", fake.created)
	}
	if issue.TagNames() != "Investigation" {
		t.Errorf("tags = %q, want Investigation", issue.TagNames())
	}
}

func TestRunIssueCreateNewTag(t *testing.T) {
	fake, api := newFakeTagServer(t, "Investigation")

	issue, err := runCreateJSON(t, api, "-t", "Investigation", "-t", "needs review")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"tag Investigation tag needs review"}; !slices.Equal(fake.commands, want) {
		t.Errorf("commands = %q, want %q", fake.commands, want)
	}
	if want := []string{"needs review"}; !slices.Equal(fake.created, want) {
		t.Errorf("created tags = %q, want %q", fake.created, want)
	}
	if issue.TagNames() != "Investigation, needs review" {
		t.Errorf("tags = %q, want Investigation, needs review", issue.TagNames())
	}
}

func TestRunIssueCreateTagFailureKeepsIssue(t *testing.T) {
	fake, api := newFakeTagServer(t)
	fake.commandErr = "Command is not applicable"

	issue, err := runCreateJSON(t, api, "-t", "Investigation")
	if err == nil {
		t.Fatal("expected error when tagging fails")
	}
	if issue.IDReadable != "AX-1" {
		t.Errorf("printed issue ID = %q, want AX-1", issue.IDReadable)
	}
	msg := formatError(err)
	for _, want := range []string{"AX-1 was created", "yt issue update AX-1", "Command is not applicable"} {
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
	if mock.createdDescription != "" {
		t.Error("issue was created despite the invalid --field")
	}
}
