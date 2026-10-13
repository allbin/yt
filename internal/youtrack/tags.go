package youtrack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const tagFields = "id,name"

// ListTags returns the tags whose name contains query (every tag when query
// is empty). Matching is the server's: a substring, case-insensitive.
func (c *Client) ListTags(query string) ([]Tag, error) {
	params := url.Values{"fields": {tagFields}, "$top": {"1000"}}
	if query != "" {
		params.Set("query", query)
	}

	data, err := c.get("/api/tags", params)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	var tags []Tag
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("parse tags: %w", err)
	}
	return tags, nil
}

// CreateTag creates a tag owned by the token's user and returns it with its
// ID.
func (c *Client) CreateTag(name string) (*Tag, error) {
	body := struct {
		Name string `json:"name"`
	}{Name: name}

	data, err := c.postJSON("/api/tags?fields="+url.QueryEscape(tagFields), body)
	if err != nil {
		return nil, fmt.Errorf("create tag %q: %w", name, err)
	}

	var tag Tag
	if err := json.Unmarshal(data, &tag); err != nil {
		return nil, fmt.Errorf("parse created tag: %w", err)
	}
	return &tag, nil
}

// TagIssue adds a tag to an issue by the tag's ID. Adding a tag the issue
// already has is a no-op on the server side.
func (c *Client) TagIssue(issueID, tagID string) error {
	body := struct {
		ID string `json:"id"`
	}{ID: tagID}
	if err := c.post("/api/issues/"+url.PathEscape(issueID)+"/tags", body); err != nil {
		return fmt.Errorf("tag %s: %w", issueID, err)
	}
	return nil
}

// UntagIssue removes a tag from an issue by the tag's ID. A tag the issue
// does not carry is a no-op: the server answers 404 for it.
func (c *Client) UntagIssue(issueID, tagID string) error {
	path := "/api/issues/" + url.PathEscape(issueID) + "/tags/" + url.PathEscape(tagID)
	err := c.delete(path)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return fmt.Errorf("untag %s: %w", issueID, err)
	}
	return nil
}
