package youtrack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// testLinkTypes mirrors the live AX instance's configured link types.
var testLinkTypes = []LinkType{
	{ID: "105-0", Name: "Relates", SourceToTarget: "relates to", TargetToSource: "", Directed: false},
	{ID: "105-1", Name: "Depend", SourceToTarget: "is required for", TargetToSource: "depends on", Directed: true},
	{ID: "105-2", Name: "Duplicate", SourceToTarget: "is duplicated by", TargetToSource: "duplicates", Directed: true, Aggregation: true},
	{ID: "105-3", Name: "Subtask", SourceToTarget: "parent for", TargetToSource: "subtask of", Directed: true, Aggregation: true},
}

func TestResolveRelation(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		wantType  string
		wantDir   string
		wantPhr   string
		wantErr   bool
		errSubstr string
	}{
		{name: "subtask kebab", alias: "subtask-of", wantType: "Subtask", wantDir: DirInward, wantPhr: "subtask of"},
		{name: "subtask spaced", alias: "subtask of", wantType: "Subtask", wantDir: DirInward, wantPhr: "subtask of"},
		{name: "subtask squashed", alias: "subtaskof", wantType: "Subtask", wantDir: DirInward, wantPhr: "subtask of"},
		{name: "subtask uppercase", alias: "Subtask-Of", wantType: "Subtask", wantDir: DirInward, wantPhr: "subtask of"},
		{name: "subtask prefix", alias: "subtask", wantType: "Subtask", wantDir: DirInward, wantPhr: "subtask of"},
		{name: "parent for", alias: "parent-for", wantType: "Subtask", wantDir: DirOutward, wantPhr: "parent for"},
		{name: "parent prefix", alias: "parent", wantType: "Subtask", wantDir: DirOutward, wantPhr: "parent for"},
		{name: "relates", alias: "relates", wantType: "Relates", wantDir: DirBoth, wantPhr: "relates to"},
		{name: "relates to", alias: "relates to", wantType: "Relates", wantDir: DirBoth, wantPhr: "relates to"},
		{name: "relates kebab", alias: "relates-to", wantType: "Relates", wantDir: DirBoth, wantPhr: "relates to"},
		{name: "depends on", alias: "depends-on", wantType: "Depend", wantDir: DirInward, wantPhr: "depends on"},
		{name: "depends prefix", alias: "depends", wantType: "Depend", wantDir: DirInward, wantPhr: "depends on"},
		{name: "is required for", alias: "is-required-for", wantType: "Depend", wantDir: DirOutward, wantPhr: "is required for"},
		{name: "required substring", alias: "required-for", wantType: "Depend", wantDir: DirOutward, wantPhr: "is required for"},
		{name: "duplicates", alias: "duplicates", wantType: "Duplicate", wantDir: DirInward, wantPhr: "duplicates"},
		{name: "is duplicated by", alias: "is-duplicated-by", wantType: "Duplicate", wantDir: DirOutward, wantPhr: "is duplicated by"},
		{name: "unknown", alias: "frobnicate", wantErr: true, errSubstr: "unknown relation"},
		{name: "empty", alias: "", wantErr: true, errSubstr: "unknown relation"},
		{name: "ambiguous is", alias: "is", wantErr: true, errSubstr: "ambiguous"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := ResolveRelation(testLinkTypes, tt.alias)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %+v", rel)
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q missing %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rel.LinkType.Name != tt.wantType {
				t.Errorf("type = %q, want %q", rel.LinkType.Name, tt.wantType)
			}
			if rel.Direction != tt.wantDir {
				t.Errorf("direction = %q, want %q", rel.Direction, tt.wantDir)
			}
			if rel.Phrase != tt.wantPhr {
				t.Errorf("phrase = %q, want %q", rel.Phrase, tt.wantPhr)
			}
		})
	}
}

func TestUnknownRelationListsValidOnes(t *testing.T) {
	_, err := ResolveRelation(testLinkTypes, "nope")
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"relates to", "depends on", "subtask of", "parent for"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing valid relation %q: %v", want, err)
		}
	}
}

func TestIssueLinkPhrase(t *testing.T) {
	subtask := LinkType{Name: "Subtask", SourceToTarget: "parent for", TargetToSource: "subtask of", Directed: true}
	tests := []struct {
		dir  string
		want string
	}{
		{DirInward, "subtask of"},
		{DirOutward, "parent for"},
		{DirBoth, "parent for"},
	}
	for _, tt := range tests {
		l := IssueLink{Direction: tt.dir, LinkType: subtask}
		if got := l.Phrase(); got != tt.want {
			t.Errorf("Phrase(%s) = %q, want %q", tt.dir, got, tt.want)
		}
	}
}

func TestNonEmptyLinksAndFind(t *testing.T) {
	links := []IssueLink{
		{ID: "105-0", Direction: DirBoth, LinkType: testLinkTypes[0], Issues: nil},
		{ID: "105-3t", Direction: DirInward, LinkType: testLinkTypes[3], Issues: []LinkedIssue{{ID: "2-6261", IDReadable: "AX-332", Summary: "Story"}}},
		{ID: "105-3s", Direction: DirOutward, LinkType: testLinkTypes[3], Issues: nil},
	}

	got := nonEmptyLinks(links)
	if len(got) != 1 || got[0].ID != "105-3t" {
		t.Fatalf("nonEmptyLinks = %+v, want only 105-3t", got)
	}

	rel := Relation{LinkType: testLinkTypes[3], Phrase: "subtask of", Direction: DirInward}
	if l, m := FindLink(got, rel, "ax-332"); l == nil || l.ID != "105-3t" || m == nil || m.ID != "2-6261" {
		t.Errorf("FindLink (case-insensitive) = %+v / %+v, want link 105-3t target 2-6261", l, m)
	}
	if l, m := FindLink(got, rel, "AX-999"); l != nil || m != nil {
		t.Errorf("FindLink for absent target = %+v / %+v, want nil", l, m)
	}
	outward := Relation{LinkType: testLinkTypes[3], Phrase: "parent for", Direction: DirOutward}
	if l, _ := FindLink(got, outward, "AX-332"); l != nil {
		t.Errorf("FindLink wrong direction = %+v, want nil", l)
	}
}

func TestListLinkTypes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/issueLinkTypes" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode(testLinkTypes); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	types, err := client.ListLinkTypes()
	if err != nil {
		t.Fatal(err)
	}
	if len(types) != 4 || types[3].Name != "Subtask" {
		t.Fatalf("got %+v", types)
	}
}

func TestGetIssueParsesAndFiltersLinks(t *testing.T) {
	raw := `{
		"idReadable": "AX-804",
		"summary": "Child",
		"links": [
			{"id":"105-0","direction":"BOTH","linkType":{"name":"Relates","sourceToTarget":"relates to","targetToSource":""},"issues":[]},
			{"id":"105-3t","direction":"INWARD","linkType":{"name":"Subtask","sourceToTarget":"parent for","targetToSource":"subtask of"},"issues":[{"idReadable":"AX-332","summary":"Story"}]}
		]
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte(raw)); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	issue, err := client.GetIssue("AX-804")
	if err != nil {
		t.Fatal(err)
	}
	if len(issue.Links) != 1 {
		t.Fatalf("got %d links, want 1 (empty slot filtered)", len(issue.Links))
	}
	l := issue.Links[0]
	if l.Phrase() != "subtask of" {
		t.Errorf("phrase = %q, want 'subtask of'", l.Phrase())
	}
	if len(l.Issues) != 1 || l.Issues[0].IDReadable != "AX-332" {
		t.Errorf("linked issue = %+v, want AX-332", l.Issues)
	}
}

func TestCreateLinkSendsCommand(t *testing.T) {
	var got struct {
		Query  string `json:"query"`
		Issues []struct {
			IDReadable string `json:"idReadable"`
		} `json:"issues"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/commands" || r.Method != http.MethodPost {
			t.Errorf("got %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	if err := client.CreateLink("AX-804", "subtask of", "AX-332"); err != nil {
		t.Fatal(err)
	}
	if got.Query != "subtask of AX-332" {
		t.Errorf("query = %q, want 'subtask of AX-332'", got.Query)
	}
	if len(got.Issues) != 1 || got.Issues[0].IDReadable != "AX-804" {
		t.Errorf("issues = %+v, want [AX-804]", got.Issues)
	}
}

func TestRemoveLinkDeletesEndpoint(t *testing.T) {
	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	if err := client.RemoveLink("AX-804", "105-3t", "AX-332"); err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
	want := "/api/issues/AX-804/links/105-3t/issues/AX-332"
	if gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
}
