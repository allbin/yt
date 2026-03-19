package tui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func testIssue(priority, assignee string) youtrack.Issue {
	fields := []youtrack.CustomField{
		{Name: "State", Value: json.RawMessage(`{"name":"Open"}`)},
	}
	if priority != "" {
		fields = append(fields, youtrack.CustomField{
			Name: "Priority", Value: json.RawMessage(`{"name":"` + priority + `"}`),
		})
	}
	if assignee != "" {
		fields = append(fields, youtrack.CustomField{
			Name: "Assignee", Value: json.RawMessage(`{"fullName":"` + assignee + `"}`),
		})
	}
	return youtrack.Issue{
		IDReadable:   "PROJ-123",
		Summary:      "Fix the login bug",
		CustomFields: fields,
	}
}

func TestRenderCardBasic(t *testing.T) {
	issue := testIssue("Major", "alice")
	out := renderCard(issue, 30, false, false)

	for _, want := range []string{"PROJ-123", "Fix the login bug", "alice"} {
		if !strings.Contains(out, want) {
			t.Errorf("card missing %q", want)
		}
	}
	// Major should have icon
	if !strings.Contains(out, "\uf139") {
		t.Error("card missing major priority icon")
	}
}

func TestRenderCardFocusedBorder(t *testing.T) {
	issue := testIssue("Normal", "bob")
	out := renderCard(issue, 30, true, false)

	if !strings.Contains(out, "┃") {
		t.Error("focused card should use thick border (┃)")
	}
}

func TestRenderCardLongSummaryWraps(t *testing.T) {
	issue := testIssue("Normal", "")
	issue.Summary = "This is a very long summary that should definitely wrap across multiple lines in the card"
	out := renderCard(issue, 30, false, false)

	lines := strings.Split(out, "\n")
	// Should have more than 3 lines (top border + id + wrapped summary + bottom border)
	if len(lines) < 5 {
		t.Errorf("expected wrapped summary to produce >5 lines, got %d", len(lines))
	}
}

func TestRenderCardEmptyAssignee(t *testing.T) {
	issue := testIssue("Normal", "")
	out := renderCard(issue, 30, false, false)

	if strings.Contains(out, "\uf007") {
		t.Error("card should not contain assignee icon when assignee is empty")
	}
}

func TestRenderCardRespectsWidth(t *testing.T) {
	issue := testIssue("Normal", "")
	narrow := renderCard(issue, 24, false, false)
	wide := renderCard(issue, 50, false, false)

	narrowMax := maxLineWidth(narrow)
	wideMax := maxLineWidth(wide)
	if narrowMax >= wideMax {
		t.Errorf("narrow card (%d) should be narrower than wide card (%d)", narrowMax, wideMax)
	}
}

func TestRenderCardPriorityIcons(t *testing.T) {
	tests := []struct {
		priority string
		icon     string
	}{
		{"Critical", "\uf139"},
		{"Show-stopper", "\uf139"},
		{"Major", "\uf139"},
		{"Minor", "\uf13a"},
		{"Nice to have", "\uf13a"},
		{"Normal", ""},
	}
	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			issue := testIssue(tt.priority, "")
			out := renderCard(issue, 30, false, false)
			if tt.icon == "" {
				// Normal: no priority icon, just verify no icon chars
				for _, ic := range []string{"\uf139", "\uf13a"} {
					if strings.Contains(out, ic) {
						t.Errorf("Normal priority should not have icon %q", ic)
					}
				}
			} else {
				if !strings.Contains(out, tt.icon) {
					t.Errorf("priority %q: missing icon %q", tt.priority, tt.icon)
				}
			}
		})
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  int // expected number of lines
	}{
		{"short", "hello", 20, 1},
		{"wraps", "hello world foo bar baz", 10, 3},
		{"newlines", "line one\nline two", 40, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.width)
			if len(got) != tt.want {
				t.Errorf("wrapText(%q, %d) = %d lines %v, want %d", tt.text, tt.width, len(got), got, tt.want)
			}
		})
	}
}

func maxLineWidth(s string) int {
	max := 0
	for line := range strings.SplitSeq(s, "\n") {
		if len(line) > max {
			max = len(line)
		}
	}
	return max
}
