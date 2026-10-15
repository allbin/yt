package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
)

// tagLookup is a set of tag names resolved without writing anything: the
// tags that exist, and the names still to create.
type tagLookup struct {
	found   []youtrack.Tag
	missing []string
}

// lookupTags maps tag names to the instance's tags by whole name: the exact
// name first, then a case-insensitive match when exactly one tag has it. A
// name with no tag goes to missing when allowMissing is set, and is otherwise
// an error naming the closest tags, followed by missHint. Blank names are
// skipped. Nothing is written, so a caller creates missing tags only once
// every other check has passed.
func lookupTags(client youtrack.API, names []string, allowMissing bool, missHint string) (tagLookup, error) {
	var l tagLookup
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		tag, err := findTag(client, name)
		if err != nil {
			return tagLookup{}, err
		}
		switch {
		case tag != nil:
			l.found = append(l.found, *tag)
		case !allowMissing:
			return tagLookup{}, unknownTagError(client, name, missHint)
		case !slices.ContainsFunc(l.missing, func(m string) bool { return strings.EqualFold(m, name) }):
			l.missing = append(l.missing, name)
		}
	}
	return l, nil
}

// lookupAddTags looks up the names given to --tag; missing ones are allowed
// only with --create-tag. Braces are YouTrack query syntax, never part of a
// tag name: "{To be discussed}" is refused rather than matched to, or created
// as, a tag of that literal name.
func lookupAddTags(client youtrack.API, names []string, create bool) (tagLookup, error) {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if strings.HasPrefix(name, "{") || strings.HasSuffix(name, "}") {
			return tagLookup{}, fmt.Errorf("tag names take no braces: use --tag %q", strings.Trim(name, "{}"))
		}
	}
	return lookupTags(client, names, create, "; --create-tag creates it")
}

// lookupRemoveTags looks up the names given to --remove-tag, which must all
// exist.
func lookupRemoveTags(client youtrack.API, names []string) ([]youtrack.Tag, error) {
	l, err := lookupTags(client, names, false, "")
	return l.found, err
}

// createMissingTags creates the named tags.
func createMissingTags(client youtrack.API, names []string) ([]youtrack.Tag, error) {
	tags := make([]youtrack.Tag, 0, len(names))
	for _, name := range names {
		tag, err := client.CreateTag(name)
		if err != nil {
			return tags, err
		}
		tags = append(tags, *tag)
	}
	return tags, nil
}

// findTag returns the tag whose whole name is name, or nil. The server's tag
// query is a case-insensitive substring match, so the match is picked here:
// the exact name wins, and a case-insensitive one only when it is unique.
func findTag(client youtrack.API, name string) (*youtrack.Tag, error) {
	candidates, err := client.ListTags(name)
	if err != nil {
		return nil, err
	}
	var folded []youtrack.Tag
	for _, t := range candidates {
		if t.Name == name {
			return &t, nil
		}
		if strings.EqualFold(t.Name, name) {
			folded = append(folded, t)
		}
	}
	switch len(folded) {
	case 0:
		return nil, nil
	case 1:
		return &folded[0], nil
	}
	return nil, fmt.Errorf("tag %q matches %s, which differ only in case; give the exact name", name, quoteTagNames(folded))
}

// unknownTagError names the tags closest to name, so a typo is caught instead
// of becoming a new tag.
func unknownTagError(client youtrack.API, name, hint string) error {
	all, err := client.ListTags("")
	if err != nil {
		return fmt.Errorf("no tag %q%s", name, hint)
	}
	if closest := closestTags(all, name, 3); len(closest) > 0 {
		return fmt.Errorf("no tag %q (closest: %s)%s", name, quoteTagNames(closest), hint)
	}
	return fmt.Errorf("no tag %q%s", name, hint)
}

// closestTags returns up to n tags whose names are within a few edits of name
// or contain it, nearest first. Case is ignored.
func closestTags(tags []youtrack.Tag, name string, n int) []youtrack.Tag {
	type scored struct {
		tag  youtrack.Tag
		dist int
	}
	want := strings.ToLower(name)
	limit := max(2, len(want)/3)
	var near []scored
	for _, t := range tags {
		have := strings.ToLower(t.Name)
		d := editDistance(want, have)
		if d <= limit || strings.Contains(have, want) || strings.Contains(want, have) {
			near = append(near, scored{t, d})
		}
	}
	slices.SortStableFunc(near, func(a, b scored) int { return a.dist - b.dist })
	out := make([]youtrack.Tag, 0, n)
	for i := 0; i < len(near) && i < n; i++ {
		out = append(out, near[i].tag)
	}
	return out
}

// editDistance is the Levenshtein distance between a and b, by rune.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}

func quoteTagNames(tags []youtrack.Tag) string {
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = fmt.Sprintf("%q", t.Name)
	}
	return strings.Join(names, ", ")
}

// applyTags adds and removes tags on an issue, in that order.
func applyTags(client youtrack.API, issueID string, add, remove []youtrack.Tag) error {
	for _, t := range add {
		if err := client.TagIssue(issueID, t.ID); err != nil {
			return err
		}
	}
	for _, t := range remove {
		if err := client.UntagIssue(issueID, t.ID); err != nil {
			return err
		}
	}
	return nil
}

// verifyTags checks the issue as read back from the server: every added tag
// is on it and every removed one is gone, compared by ID. A write the server
// accepted but did not apply is an error, never a silent success.
func verifyTags(issue *youtrack.Issue, added, removed []youtrack.Tag) error {
	has := func(id string) bool {
		return slices.ContainsFunc(issue.Tags, func(t youtrack.Tag) bool { return t.ID == id })
	}
	for _, t := range added {
		if !has(t.ID) {
			return fmt.Errorf("tag %q (%s) is not on %s after tagging", t.Name, t.ID, issue.IDReadable)
		}
	}
	for _, t := range removed {
		if has(t.ID) {
			return fmt.Errorf("tag %q (%s) is still on %s after untagging", t.Name, t.ID, issue.IDReadable)
		}
	}
	return nil
}

// tagIDs returns the tags' IDs.
func tagIDs(tags []youtrack.Tag) []string {
	ids := make([]string, len(tags))
	for i, t := range tags {
		ids[i] = t.ID
	}
	return ids
}
