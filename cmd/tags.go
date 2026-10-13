package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
)

// resolveTags maps tag names to the instance's tags, matched by whole name
// case-insensitively, so "to be discussed" is the existing "To be discussed"
// and never a new tag. A name with no tag is created when create is set and
// an error otherwise. The result keeps the input order.
func resolveTags(client youtrack.API, names []string, create bool) ([]youtrack.Tag, error) {
	tags := make([]youtrack.Tag, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		tag, err := findTag(client, name)
		if err != nil {
			return nil, err
		}
		if tag == nil && !create {
			return nil, fmt.Errorf("no tag %q", name)
		}
		if tag == nil {
			if tag, err = client.CreateTag(name); err != nil {
				return nil, err
			}
		}
		tags = append(tags, *tag)
	}
	return tags, nil
}

// findTag returns the tag whose whole name equals name, ignoring case, or
// nil. The server's tag query is a substring match, so the exact match is
// picked here.
func findTag(client youtrack.API, name string) (*youtrack.Tag, error) {
	candidates, err := client.ListTags(name)
	if err != nil {
		return nil, err
	}
	for i := range candidates {
		if strings.EqualFold(candidates[i].Name, name) {
			return &candidates[i], nil
		}
	}
	return nil, nil
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

// tagIDs returns the tags' IDs.
func tagIDs(tags []youtrack.Tag) []string {
	ids := make([]string, len(tags))
	for i, t := range tags {
		ids[i] = t.ID
	}
	return ids
}
