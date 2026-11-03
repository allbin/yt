package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

const commentFields = "id,text,author(login,fullName),created,updated"

// commentPageSize is how many comments ListComments reads per request.
var commentPageSize = 100

// ListComments returns every comment on the issue or article, reading page
// by page: the server stops at a default page size otherwise.
func (c *Client) ListComments(id string) ([]Comment, error) {
	params := url.Values{"fields": {commentFields}, "$top": {strconv.Itoa(commentPageSize)}}
	var comments []Comment
	for skip := 0; ; skip += commentPageSize {
		params.Set("$skip", strconv.Itoa(skip))
		data, err := c.get(entityPath(id)+"/comments", params)
		if err != nil {
			return nil, fmt.Errorf("list comments for %s: %w", id, err)
		}
		var page []Comment
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("parse comments: %w", err)
		}
		comments = append(comments, page...)
		if len(page) < commentPageSize {
			return comments, nil
		}
	}
}

// GetComment returns one comment on the issue or article.
func (c *Client) GetComment(id, commentID string) (*Comment, error) {
	data, err := c.get(commentPath(id, commentID), url.Values{"fields": {commentFields}})
	if err != nil {
		return nil, fmt.Errorf("fetch comment %s on %s: %w", commentID, id, err)
	}
	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("parse comment: %w", err)
	}
	return &comment, nil
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

func commentPath(id, commentID string) string {
	return entityPath(id) + "/comments/" + url.PathEscape(commentID)
}

// UpdateComment replaces a comment's text and returns the comment as saved.
func (c *Client) UpdateComment(id, commentID, text string) (*Comment, error) {
	body := struct {
		Text string `json:"text"`
	}{Text: text}
	data, err := c.postJSON(commentPath(id, commentID)+"?fields="+url.QueryEscape(commentFields), body)
	if err != nil {
		return nil, fmt.Errorf("edit comment %s on %s: %w", commentID, id, err)
	}
	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("parse comment: %w", err)
	}
	return &comment, nil
}

// DeleteComment removes a comment permanently.
func (c *Client) DeleteComment(id, commentID string) error {
	if err := c.delete(commentPath(id, commentID)); err != nil {
		return fmt.Errorf("delete comment %s on %s: %w", commentID, id, err)
	}
	return nil
}
