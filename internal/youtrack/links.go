package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Link direction values returned by the YouTrack API.
const (
	DirOutward = "OUTWARD"
	DirInward  = "INWARD"
	DirBoth    = "BOTH"
)

// LinkType is an instance-defined issue link type with its directed phrases.
type LinkType struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	SourceToTarget string `json:"sourceToTarget"`
	TargetToSource string `json:"targetToSource"`
	Directed       bool   `json:"directed"`
	Aggregation    bool   `json:"aggregation"`
}

// LinkedIssue is the minimal issue reference returned inside a link. ID is the
// internal database id (e.g. "2-6261"), required by the link DELETE endpoint.
type LinkedIssue struct {
	ID         string `json:"id,omitempty"`
	IDReadable string `json:"idReadable"`
	Summary    string `json:"summary"`
}

// IssueLink is one type+direction slot of an issue's links. The API returns a
// slot per type/direction; populated ones have a non-empty Issues list.
type IssueLink struct {
	ID        string        `json:"id"`
	Direction string        `json:"direction"`
	LinkType  LinkType      `json:"linkType"`
	Issues    []LinkedIssue `json:"issues"`
}

// Phrase returns the directed phrase matching this link's direction, e.g.
// "subtask of" for an inward Subtask link, "parent for" for an outward one.
func (l IssueLink) Phrase() string {
	if l.Direction == DirInward && l.LinkType.TargetToSource != "" {
		return l.LinkType.TargetToSource
	}
	return l.LinkType.SourceToTarget
}

// Relation is a resolved link type + direction, identified by a directed phrase.
type Relation struct {
	LinkType  LinkType
	Phrase    string // canonical directed phrase, e.g. "subtask of"
	Direction string // OUTWARD, INWARD, or BOTH
}

const linkTypeFields = "id,name,sourceToTarget,targetToSource,directed,aggregation"

// ListLinkTypes returns the instance's available issue link types.
func (c *Client) ListLinkTypes() ([]LinkType, error) {
	params := url.Values{"fields": {linkTypeFields}}

	data, err := c.get("/api/issueLinkTypes", params)
	if err != nil {
		return nil, fmt.Errorf("list link types: %w", err)
	}

	var types []LinkType
	if err := json.Unmarshal(data, &types); err != nil {
		return nil, fmt.Errorf("parse link types: %w", err)
	}
	return types, nil
}

// CreateLink links sourceID to targetID using the given directed phrase via the
// commands API, which infers direction from the phrase. Adding an existing link
// is a no-op on the server side.
func (c *Client) CreateLink(sourceID, phrase, targetID string) error {
	if err := c.runCommand(phrase+" "+targetID, sourceID); err != nil {
		return fmt.Errorf("link %s to %s: %w", sourceID, targetID, err)
	}
	return nil
}

// RemoveLink deletes the link identified by linkID (an IssueLink id such as
// "105-3t") between sourceID and the linked issue. targetRef must be the linked
// issue's internal database id (e.g. "2-6261"); the endpoint rejects idReadable.
func (c *Client) RemoveLink(sourceID, linkID, targetRef string) error {
	path := "/api/issues/" + url.PathEscape(sourceID) +
		"/links/" + url.PathEscape(linkID) +
		"/issues/" + url.PathEscape(targetRef)
	if err := c.delete(path); err != nil {
		return fmt.Errorf("unlink from %s: %w", sourceID, err)
	}
	return nil
}

// relationCandidates expands link types into one Relation per usable phrase.
func relationCandidates(types []LinkType) []Relation {
	var out []Relation
	for _, t := range types {
		if t.SourceToTarget != "" {
			dir := DirOutward
			if !t.Directed {
				dir = DirBoth
			}
			out = append(out, Relation{LinkType: t, Phrase: t.SourceToTarget, Direction: dir})
		}
		if t.Directed && t.TargetToSource != "" {
			out = append(out, Relation{LinkType: t, Phrase: t.TargetToSource, Direction: DirInward})
		}
	}
	return out
}

// RelationPhrases returns the valid directed phrases for the given link types,
// in API order, for display in error messages.
func RelationPhrases(types []LinkType) []string {
	cands := relationCandidates(types)
	phrases := make([]string, len(cands))
	for i, c := range cands {
		phrases[i] = c.Phrase
	}
	return phrases
}

// normalizeRelation reduces a phrase or alias to a comparison key, stripping
// everything but lowercase letters and digits. "subtask-of", "subtask of" and
// "subtaskof" all map to "subtaskof".
func normalizeRelation(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ResolveRelation maps a user-supplied relation alias to a link type and
// direction by matching against the fetched link types. Matching is
// kebab/space-insensitive: exact normalized match first, then unique prefix,
// then unique substring. Ambiguous or unknown aliases return an error listing
// the valid relations.
func ResolveRelation(types []LinkType, alias string) (Relation, error) {
	key := normalizeRelation(alias)
	cands := relationCandidates(types)
	if key == "" {
		return Relation{}, unknownRelationErr(alias, cands)
	}

	for _, c := range cands {
		if normalizeRelation(c.Phrase) == key {
			return c, nil
		}
	}

	if m, err := uniqueMatch(alias, cands, func(phrase string) bool {
		return strings.HasPrefix(phrase, key) || strings.HasPrefix(key, phrase)
	}); err != nil || m != nil {
		return deref(m), err
	}

	if m, err := uniqueMatch(alias, cands, func(phrase string) bool {
		return strings.Contains(phrase, key)
	}); err != nil || m != nil {
		return deref(m), err
	}

	return Relation{}, unknownRelationErr(alias, cands)
}

// uniqueMatch returns the single candidate satisfying match, an ambiguity
// error if several match, or (nil, nil) if none match.
func uniqueMatch(alias string, cands []Relation, match func(phrase string) bool) (*Relation, error) {
	var found []Relation
	for _, c := range cands {
		if match(normalizeRelation(c.Phrase)) {
			found = append(found, c)
		}
	}
	switch len(found) {
	case 0:
		return nil, nil
	case 1:
		return &found[0], nil
	default:
		phrases := make([]string, len(found))
		for i, c := range found {
			phrases[i] = c.Phrase
		}
		return nil, fmt.Errorf("ambiguous relation %q matches: %s", alias, strings.Join(phrases, ", "))
	}
}

func deref(r *Relation) Relation {
	if r == nil {
		return Relation{}
	}
	return *r
}

func unknownRelationErr(alias string, cands []Relation) error {
	phrases := make([]string, len(cands))
	for i, c := range cands {
		phrases[i] = c.Phrase
	}
	return fmt.Errorf("unknown relation %q; valid relations: %s", alias, strings.Join(phrases, ", "))
}

// nonEmptyLinks drops the empty type/direction slots the API returns.
func nonEmptyLinks(links []IssueLink) []IssueLink {
	var out []IssueLink
	for _, l := range links {
		if len(l.Issues) > 0 {
			out = append(out, l)
		}
	}
	return out
}

// FindLink returns the populated link matching the relation and the specific
// linked issue matching targetID. Both are nil if no such link exists.
func FindLink(links []IssueLink, rel Relation, targetID string) (*IssueLink, *LinkedIssue) {
	for i := range links {
		l := &links[i]
		if l.LinkType.Name != rel.LinkType.Name || l.Direction != rel.Direction {
			continue
		}
		for j := range l.Issues {
			if strings.EqualFold(l.Issues[j].IDReadable, targetID) {
				return l, &l.Issues[j]
			}
		}
	}
	return nil, nil
}
