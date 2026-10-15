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

// fakeServer mimics what issue create and the tag flags touch on a real
// YouTrack: the tag list, tag creation, the create body (tags by ID, fields
// checked), the project's field schema, and issue AX-1's tags. It fails the
// test if a tag whose name is wrapped in braces is ever created.
type fakeServer struct {
	t           *testing.T
	tags        []youtrack.Tag // every tag the instance knows
	created     []string       // tags created through POST /api/tags
	createBody  map[string]json.RawMessage
	createCalls int
	rejectField string // when set, a create naming this field fails
	issueTags   []youtrack.Tag
	tagWrites   int  // POST and DELETE on AX-1's tags
	dropTags    bool // accept tag writes without applying them
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

// tag returns the instance's tag named name.
func (f *fakeServer) tag(name string) youtrack.Tag {
	f.t.Helper()
	for _, t := range f.tags {
		if t.Name == name {
			return t
		}
	}
	f.t.Fatalf("fake has no tag %q", name)
	return youtrack.Tag{}
}

func (f *fakeServer) issueTagNames() []string {
	names := make([]string, len(f.issueTags))
	for i, t := range f.issueTags {
		names[i] = t.Name
	}
	return names
}

func (f *fakeServer) addIssueTag(id string) {
	if f.dropTags || slices.ContainsFunc(f.issueTags, func(t youtrack.Tag) bool { return t.ID == id }) {
		return
	}
	for _, t := range f.tags {
		if t.ID == id {
			f.issueTags = append(f.issueTags, t)
		}
	}
}

func (f *fakeServer) serve(w http.ResponseWriter, r *http.Request) {
	const issueTagsPath = "/api/issues/AX-1/tags"
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
		if strings.HasPrefix(body.Name, "{") || strings.HasSuffix(body.Name, "}") {
			f.t.Errorf("created tag %q, a name in braces", body.Name)
		}
		tag := youtrack.Tag{ID: "6-new-" + body.Name, Name: body.Name}
		f.tags = append(f.tags, tag)
		f.created = append(f.created, body.Name)
		f.write(w, tag)
	case r.Method == http.MethodPost && r.URL.Path == issueTagsPath:
		var body youtrack.Tag
		f.decode(r, &body)
		f.tagWrites++
		f.addIssueTag(body.ID)
		f.write(w, body)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, issueTagsPath+"/"):
		id := strings.TrimPrefix(r.URL.Path, issueTagsPath+"/")
		f.tagWrites++
		if !f.dropTags {
			f.issueTags = slices.DeleteFunc(f.issueTags, func(t youtrack.Tag) bool { return t.ID == id })
		}
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
		var refs []youtrack.Tag
		_ = json.Unmarshal(f.createBody["tags"], &refs)
		for _, ref := range refs {
			f.addIssueTag(ref.ID)
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
