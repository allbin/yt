package cmd

import (
	"slices"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestResolveTags(t *testing.T) {
	instance := []youtrack.Tag{{ID: "6-63", Name: "To be discussed"}, {ID: "6-1", Name: "QA"}}

	t.Run("existing_by_whole_name_any_case", func(t *testing.T) {
		mock := &mockAPI{tags: instance}
		got, err := resolveTags(mock, []string{"to be discussed", "qa"}, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 || got[0].ID != "6-63" || got[1].ID != "6-1" {
			t.Errorf("got %+v", got)
		}
		if len(mock.createdTags) != 0 {
			t.Errorf("created %q, want none", mock.createdTags)
		}
	})

	t.Run("substring_is_not_a_match", func(t *testing.T) {
		mock := &mockAPI{tags: instance}
		got, err := resolveTags(mock, []string{"discussed"}, true)
		if err != nil {
			t.Fatal(err)
		}
		if want := []string{"discussed"}; !slices.Equal(mock.createdTags, want) {
			t.Errorf("created %q, want %q", mock.createdTags, want)
		}
		if len(got) != 1 || got[0].Name != "discussed" {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("missing_without_create_errors", func(t *testing.T) {
		mock := &mockAPI{tags: instance}
		_, err := resolveTags(mock, []string{"nope"}, false)
		if err == nil || !strings.Contains(err.Error(), `no tag "nope"`) {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("keeps_order_skips_blank", func(t *testing.T) {
		mock := &mockAPI{tags: instance}
		got, err := resolveTags(mock, []string{"QA", " ", "To be discussed"}, false)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 || got[0].ID != "6-1" || got[1].ID != "6-63" {
			t.Errorf("got %+v", got)
		}
	})
}

func TestParseFieldArgs(t *testing.T) {
	tests := []struct {
		name    string
		in      []string
		want    []fieldArg
		wantErr bool
	}{
		{"simple", []string{"Severity=Critical"}, []fieldArg{{"Severity", "Critical"}}, false},
		{"multi_word", []string{"Type=User Story"}, []fieldArg{{"Type", "User Story"}}, false},
		{"multi_word_name", []string{"Fix versions=2.0"}, []fieldArg{{"Fix versions", "2.0"}}, false},
		{"trims", []string{" Type = Bug "}, []fieldArg{{"Type", "Bug"}}, false},
		{"equals_in_value", []string{"URL=https://x.com/a=b"}, []fieldArg{{"URL", "https://x.com/a=b"}}, false},
		{"empty_value", []string{"Subsystem="}, []fieldArg{{"Subsystem", ""}}, false},
		{"repeated", []string{"Subsystem=API", "Subsystem=Management UI"},
			[]fieldArg{{"Subsystem", "API"}, {"Subsystem", "Management UI"}}, false},
		{"invalid", []string{"bad-format"}, nil, true},
		{"empty_name", []string{"=Value"}, nil, true},
		{"whitespace_name", []string{"  =Value"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFieldArgs(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseFieldArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !slices.Equal(got, tt.want) {
				t.Errorf("parseFieldArgs() = %q, want %q", got, tt.want)
			}
		})
	}
}
