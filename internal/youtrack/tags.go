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

// AddTag adds a tag to an issue or article by the tag's ID. Adding a tag it
// already has is a no-op on the server side.
func (c *Client) AddTag(id, tagID string) error {
	body := struct {
		ID string `json:"id"`
	}{ID: tagID}
	if err := c.post(entityPath(id)+"/tags", body); err != nil {
		return fmt.Errorf("tag %s: %w", id, err)
	}
	return nil
}

// RemoveTag removes a tag from an issue or article by the tag's ID. A tag it
// does not carry is a no-op: the server answers 404 for it.
func (c *Client) RemoveTag(id, tagID string) error {
	err := c.delete(entityPath(id) + "/tags/" + url.PathEscape(tagID))
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return nil
	}
	if err != nil {
		return fmt.Errorf("untag %s: %w", id, err)
	}
	return nil
}
