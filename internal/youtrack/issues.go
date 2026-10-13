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

const linkFields = "links(id,direction," +
	"linkType(id,name,sourceToTarget,targetToSource,directed,aggregation)," +
	"issues(id,idReadable,summary))"

// issueDetailFields extends issueFields with links for single-issue views.
const issueDetailFields = issueFields + "," + linkFields

func (c *Client) GetIssue(id string) (*Issue, error) {
	params := url.Values{"fields": {issueDetailFields}}

	data, err := c.get("/api/issues/"+url.PathEscape(id), params)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", id, err)
	}

	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parse %s: %w", id, err)
	}
	issue.Links = nonEmptyLinks(issue.Links)

	return &issue, nil
}

// UpdateIssueFields writes core fields (summary, description) and custom
// fields in one request, so the server applies all or none of them.
func (c *Client) UpdateIssueFields(id string, core map[string]string, custom []FieldUpdate) error {
	body := make(map[string]any, len(core)+1)
	for k, v := range core {
		body[k] = v
	}
	if len(custom) > 0 {
		body["customFields"] = custom
	}
	if err := c.post("/api/issues/"+url.PathEscape(id), body); err != nil {
		return fmt.Errorf("update fields %s: %w", id, err)
	}
	return nil
}

// NewIssue is what CreateIssue sends: the issue's core fields, its custom
// fields and the IDs of the tags to apply.
type NewIssue struct {
	Project     string
	Summary     string
	Description string
	Fields      []FieldUpdate
	TagIDs      []string
}

// CreateIssue creates an issue with its core fields, custom fields and tags
// in one request, so a rejected value creates nothing. Tags are referenced by
// ID: the create body cannot name a tag.
func (c *Client) CreateIssue(n NewIssue) (*Issue, error) {
	type idRef struct {
		ID string `json:"id"`
	}
	body := struct {
		Project struct {
			ShortName string `json:"shortName"`
		} `json:"project"`
		Summary      string        `json:"summary"`
		Description  string        `json:"description,omitempty"`
		CustomFields []FieldUpdate `json:"customFields,omitempty"`
		Tags         []idRef       `json:"tags,omitempty"`
	}{
		Summary:      n.Summary,
		Description:  n.Description,
		CustomFields: n.Fields,
	}
	body.Project.ShortName = n.Project
	for _, id := range n.TagIDs {
		body.Tags = append(body.Tags, idRef{ID: id})
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
