package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

const issueFields = "idReadable,summary,description,resolved,created,updated," +
	"tags(name)," +
	"customFields(name,value(name,text,presentation,login,fullName))"

func (c *Client) GetIssue(id string) (*Issue, error) {
	params := url.Values{"fields": {issueFields}}

	data, err := c.get("/api/issues/"+id, params)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", id, err)
	}

	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parse %s: %w", id, err)
	}

	return &issue, nil
}

func (c *Client) ListIssues(query string, limit int) ([]Issue, error) {
	params := url.Values{"fields": {issueFields}}
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

	return issues, nil
}
