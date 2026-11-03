package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

const commentFields = "id,text,author(login,fullName),created,updated"

func (c *Client) ListComments(id string) ([]Comment, error) {
	params := url.Values{"fields": {commentFields}}

	data, err := c.get(entityPath(id)+"/comments", params)
	if err != nil {
		return nil, fmt.Errorf("list comments for %s: %w", id, err)
	}

	var comments []Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("parse comments: %w", err)
	}

	return comments, nil
}

func (c *Client) AddComment(id, text string) (*Comment, error) {
	body := struct {
		Text string `json:"text"`
	}{Text: text}

	params := url.Values{"fields": {commentFields}}
	path := entityPath(id) + "/comments?" + params.Encode()

	data, err := c.postJSON(path, body)
	if err != nil {
		return nil, fmt.Errorf("add comment to %s: %w", id, err)
	}

	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("parse comment: %w", err)
	}

	return &comment, nil
}
