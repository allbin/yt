package git

import "testing"

func TestIssueFromBranch(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{"PROJ-123", "PROJ-123"},
		{"proj-123", "PROJ-123"},
		{"PROJ-123-some-slug", "PROJ-123"},
		{"proj-123-fix-login-bug", "PROJ-123"},
		{"HK-1052-add-read-only-admin", "HK-1052"},
		{"AX-1", "AX-1"},
		{"main", ""},
		{"HEAD", ""},
		{"feature-branch", ""},
		{"123-no-prefix", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			got := IssueFromBranch(tt.branch)
			if got != tt.want {
				t.Errorf("IssueFromBranch(%q) = %q, want %q", tt.branch, got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "Fix the login bug", "fix-the-login-bug"},
		{"special_chars", "Add API endpoint (v2)", "add-api-endpoint-v2"},
		{"diacritics", "Hållkoll förbättringar", "hallkoll-forbattringar"},
		{"leading_trailing", "  --hello-- ", "hello"},
		{"multiple_spaces", "too   many   spaces", "too-many-spaces"},
		{"empty", "", ""},
		{"hash_and_quotes", "#ordet Ticka/Tickad", "ordet-ticka-tickad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSlugifyTruncation(t *testing.T) {
	long := "this is an extremely long issue summary that should be truncated to a reasonable branch name length"
	slug := Slugify(long)
	if len(slug) > 60 {
		t.Errorf("slug too long: %d chars", len(slug))
	}
	if slug[len(slug)-1] == '-' {
		t.Error("slug ends with dash after truncation")
	}
}
