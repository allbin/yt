package cmd

import (
	"slices"
	"testing"
)

func TestTagCommand(t *testing.T) {
	tests := []struct {
		name       string
		tags       []string
		removeTags []string
		want       string
	}{
		{"empty", nil, nil, ""},
		{"single", []string{"tech-debt"}, nil, "tag tech-debt"},
		{"multi", []string{"tech-debt", "scheduler"}, nil, "tag tech-debt tag scheduler"},
		{"remove", nil, []string{"obsolete"}, "untag obsolete"},
		{"add_and_remove", []string{"new-tag"}, []string{"old-tag"}, "tag new-tag untag old-tag"},
		{"multi_word", []string{"needs review"}, nil, "tag needs review"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tagCommand(tt.tags, tt.removeTags); got != tt.want {
				t.Errorf("tagCommand() = %q, want %q", got, tt.want)
			}
		})
	}
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
