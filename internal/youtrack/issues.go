package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

const issueFields = "idReadable,summary,description,resolved,created,updated," +
	"tags(name)," +
	"customFields(name,value(name,text,presentation,login,fullName))," +
	"attachments(id,name,url,size,mimeType,created)"

func (c *Client) GetIssue(id string) (*Issue, error) {
	params := url.Values{"fields": {issueFields}}

	data, err := c.get("/api/issues/"+url.PathEscape(id), params)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", id, err)
	}

	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parse %s: %w", id, err)
	}

	return &issue, nil
}

func (c *Client) UpdateIssue(id string, command string) error {
	body := struct {
		Query  string `json:"query"`
		Issues []struct {
			IDReadable string `json:"idReadable"`
		} `json:"issues"`
	}{
		Query:  command,
		Issues: []struct{ IDReadable string `json:"idReadable"` }{{IDReadable: id}},
	}
	if err := c.post("/api/commands", body); err != nil {
		return fmt.Errorf("update %s: %w", id, err)
	}
	return nil
}

func (c *Client) CreateIssue(project, summary, description string, tags []string) (*Issue, error) {
	type tagRef struct {
		Name string `json:"name"`
	}
	body := struct {
		Project     struct{ ShortName string `json:"shortName"` } `json:"project"`
		Summary     string                                        `json:"summary"`
		Description string                                        `json:"description,omitempty"`
		Tags        []tagRef                                      `json:"tags,omitempty"`
	}{
		Summary:     summary,
		Description: description,
	}
	body.Project.ShortName = project
	for _, t := range tags {
		body.Tags = append(body.Tags, tagRef{Name: t})
	}

	path := "/api/issues?fields=" + url.QueryEscape(issueFields)
	data, err := c.postJSON(path, body)
	if err != nil {
		return nil, fmt.Errorf("create issue: %w", err)
	}

	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parse created issue: %w", err)
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
