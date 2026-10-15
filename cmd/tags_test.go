package cmd

import (
	"slices"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func runUpdate(t *testing.T, api youtrack.API, args ...string) error {
	t.Helper()
	run := setupTest(t, api)
	_, err := run(append([]string{"issue", "update", "AX-1", "--json"}, args...)...)
	return err
}

func TestUpdateMultiWordTagAppliedAndRemoved(t *testing.T) {
	fake, api := newFakeServer(t, "To be discussed", "Investigation")
	fake.issueTags = []youtrack.Tag{fake.tag("Investigation")}

	if err := runUpdate(t, api, "--tag", "To be discussed", "--remove-tag", "Investigation"); err != nil {
		t.Fatal(err)
	}
	if got, want := fake.issueTagNames(), []string{"To be discussed"}; !slices.Equal(got, want) {
		t.Errorf("issue tags = %q, want %q", got, want)
	}
	if len(fake.created) != 0 {
		t.Errorf("created tags %q, want none", fake.created)
	}
}

func TestUpdateUnknownTag(t *testing.T) {
	fake, api := newFakeServer(t, "To be discussed", "Investigation")

	err := runUpdate(t, api, "--tag", "To be discused")
	if err == nil {
		t.Fatal("expected an unknown tag error")
	}
	for _, want := range []string{`no tag "To be discused"`, `closest: "To be discussed"`, "--create-tag"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}
	if len(fake.created) != 0 || fake.tagWrites != 0 {
		t.Errorf("unknown tag wrote: created %q, %d tag writes", fake.created, fake.tagWrites)
	}

	if err := runUpdate(t, api, "--tag", "needs triage", "--create-tag"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"needs triage"}; !slices.Equal(fake.created, want) || !slices.Equal(fake.issueTagNames(), want) {
		t.Errorf("created %q, issue tags %q, want %q", fake.created, fake.issueTagNames(), want)
	}
}

func TestUpdateTagCase(t *testing.T) {
	tests := []struct {
		name     string
		instance []string
		arg      string
		want     string // tag applied; empty when the name is ambiguous
	}{
		{"unique_case_mismatch", []string{"To be discussed"}, "to be discussed", "To be discussed"},
		{"exact_beats_case_mismatch", []string{"QA", "qa"}, "qa", "qa"},
		{"ambiguous_case_mismatch", []string{"QA", "qa"}, "Qa", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake, api := newFakeServer(t, tt.instance...)
			err := runUpdate(t, api, "--tag", tt.arg, "--create-tag")
			if tt.want == "" {
				if err == nil || !strings.Contains(err.Error(), "differ only in case") {
					t.Errorf("err = %v, want ambiguous", err)
				}
				if len(fake.created) != 0 || fake.tagWrites != 0 {
					t.Errorf("ambiguous tag wrote: created %q, %d tag writes", fake.created, fake.tagWrites)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := fake.issueTagNames(); !slices.Equal(got, []string{tt.want}) || len(fake.created) != 0 {
				t.Errorf("issue tags %q, created %q, want %q applied", got, fake.created, tt.want)
			}
		})
	}
}

func TestUpdateBracedTag(t *testing.T) {
	fake, api := newFakeServer(t, "To be discussed", "{To be discussed}")
	fake.issueTags = []youtrack.Tag{fake.tag("{To be discussed}")}

	for _, name := range []string{"{To be discussed}", "{To be discussed"} {
		err := runUpdate(t, api, "--tag", name, "--create-tag")
		if err == nil || !strings.Contains(err.Error(), `use --tag "To be discussed"`) {
			t.Fatalf("%s: err = %v, want braces refused", name, err)
		}
	}
	if fake.tagWrites != 0 {
		t.Errorf("%d tag writes, want none", fake.tagWrites)
	}

	// A braced tag left by an older yt can still be removed by its name.
	if err := runUpdate(t, api, "--remove-tag", "{To be discussed}"); err != nil {
		t.Fatal(err)
	}
	if len(fake.issueTags) != 0 {
		t.Errorf("issue tags = %q, want none", fake.issueTagNames())
	}
}

func TestUpdateTagNotApplied(t *testing.T) {
	fake, api := newFakeServer(t, "To be discussed")
	fake.dropTags = true

	err := runUpdate(t, api, "--tag", "To be discussed")
	if err == nil || !strings.Contains(err.Error(), `"To be discussed" (6-a) is not on AX-1`) {
		t.Errorf("err = %v, want the missing tag named", err)
	}
}

func TestUpdateTagChecksBeforeCreate(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"added_and_removed", []string{"--tag", "QA", "--remove-tag", "QA"}, "both --tag and --remove-tag"},
		{"new_and_removed", []string{"--tag", "new", "--remove-tag", "new"}, `no tag "new"`},
		{"unknown_remove", []string{"--tag", "new", "--remove-tag", "nope"}, `no tag "nope"`},
		{"ambiguous_add", []string{"--tag", "new", "--tag", "Qa"}, "differ only in case"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake, api := newFakeServer(t, "QA", "qa")
			err := runUpdate(t, api, append(tt.args, "--create-tag")...)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want %q", err, tt.want)
			}
			if strings.Contains(tt.want, "no tag") && strings.Contains(err.Error(), "--create-tag") {
				t.Errorf("--remove-tag error offers --create-tag: %v", err)
			}
			if len(fake.created) != 0 || fake.tagWrites != 0 {
				t.Errorf("failed update wrote: created %q, %d tag writes", fake.created, fake.tagWrites)
			}
		})
	}
}

func TestUpdateBlankTagSkipped(t *testing.T) {
	fake, api := newFakeServer(t, "Investigation")

	if err := runUpdate(t, api, "--tag", " ", "--tag", "Investigation"); err != nil {
		t.Fatal(err)
	}
	if got, want := fake.issueTagNames(), []string{"Investigation"}; !slices.Equal(got, want) {
		t.Errorf("issue tags = %q, want %q", got, want)
	}
}
