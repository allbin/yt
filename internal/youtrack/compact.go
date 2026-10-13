package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// IssueCompact is the compact form of an issue: the fields a list is scanned
// by, flattened to scalars. It is what "yt issue list --json --compact"
// prints, small enough that hundreds of issues fit in one read.
type IssueCompact struct {
	ID       string   `json:"id"`
	Summary  string   `json:"summary"`
	State    string   `json:"state"`
	Type     string   `json:"type"`
	Priority string   `json:"priority"`
	Assignee string   `json:"assignee"`
	Tags     []string `json:"tags"`
	Updated  string   `json:"updated"`
}

// compactFields is the narrow field set a compact list asks the server for:
// no description, attachments or links.
const compactFields = "idReadable,summary,updated,tags(name)," +
	"customFields(name,value(name,login,fullName))"

// Compact flattens the issue into its compact form. Updated is RFC 3339 in
// local time.
func (i *Issue) Compact() IssueCompact {
	tags := make([]string, 0, len(i.Tags))
	for _, t := range i.Tags {
		tags = append(tags, t.Name)
	}
	updated := ""
	if i.Updated != 0 {
		updated = time.UnixMilli(i.Updated).Format(time.RFC3339)
	}
	return IssueCompact{
		ID:       i.IDReadable,
		Summary:  i.Summary,
		State:    i.Field("State"),
		Type:     i.Field("Type"),
		Priority: i.Field("Priority"),
		Assignee: i.Field("Assignee"),
		Tags:     tags,
		Updated:  updated,
	}
}

// ListIssuesCompact is ListIssues in compact form, fetching only the fields
// a summary needs.
func (c *Client) ListIssuesCompact(query string, limit int) ([]IssueCompact, error) {
	params := url.Values{"fields": {compactFields}}
	if query != "" {
		params.Set("query", query)
	}
	if limit > 0 {
		params.Set("$top", strconv.Itoa(limit))
	}

	data, err := c.get("/api/issues", params)
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}

	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("parse issues: %w", err)
	}
	return Compacts(issues), nil
}

// Compacts flattens each issue; the result is never nil.
func Compacts(issues []Issue) []IssueCompact {
	out := make([]IssueCompact, 0, len(issues))
	for i := range issues {
		out = append(out, issues[i].Compact())
	}
	return out
}
