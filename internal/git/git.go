package git

import (
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var issueRe = regexp.MustCompile(`(?i)^([A-Z][A-Z0-9]+-\d+)`)

// CurrentBranch returns the current git branch name.
func CurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// IssueFromBranch extracts a YouTrack issue ID from a branch name.
// Returns empty string if no match. The ID is always uppercased.
func IssueFromBranch(branch string) string {
	m := issueRe.FindString(branch)
	return strings.ToUpper(m)
}

// Checkout creates and switches to a new branch.
func Checkout(name string) error {
	return exec.Command("git", "checkout", "-b", name).Run()
}

// Slugify converts a string into a branch-safe slug.
// Strips diacritics, lowercases, replaces non-alphanumeric runs with dashes.
func Slugify(s string) string {
	// Normalize to NFD and strip combining marks (diacritics)
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}

	s = strings.ToLower(b.String())

	// Replace non-alphanumeric runs with single dash
	slug := regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(s, "-")
	slug = strings.Trim(slug, "-")

	// Truncate to keep branch names reasonable
	const maxLen = 60
	if len(slug) > maxLen {
		slug = slug[:maxLen]
		slug = strings.TrimRight(slug, "-")
	}

	return slug
}
