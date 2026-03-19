package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

const commentFields = "id,text,author(login,fullName),created,updated"

func (c *Client) ListComments(issueID string) ([]Comment, error) {
	params := url.Values{"fields": {commentFields}}

	data, err := c.get("/api/issues/"+issueID+"/comments", params)
	if err != nil {
		return nil, fmt.Errorf("list comments for %s: %w", issueID, err)
	}

	var comments []Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, fmt.Errorf("parse comments: %w", err)
	}

	return comments, nil
}

func (c *Client) AddComment(issueID, text string) (*Comment, error) {
	body := struct {
		Text string `json:"text"`
	}{Text: text}

	params := url.Values{"fields": {commentFields}}
	path := "/api/issues/" + issueID + "/comments?" + params.Encode()

	data, err := c.postJSON(path, body)
	if err != nil {
		return nil, fmt.Errorf("add comment to %s: %w", issueID, err)
	}

	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("parse comment: %w", err)
	}

	return &comment, nil
}
