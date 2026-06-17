package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ListSprintIssues returns the readable IDs of the issues on a board's sprint.
func (c *Client) ListSprintIssues(agileID, sprintID string) ([]string, error) {
	path := fmt.Sprintf("/api/agiles/%s/sprints/%s/issues",
		url.PathEscape(agileID), url.PathEscape(sprintID))
	params := url.Values{"fields": {"idReadable"}, "$top": {"500"}}

	data, err := c.get(path, params)
	if err != nil {
		return nil, fmt.Errorf("list sprint issues: %w", err)
	}

	var refs []struct {
		IDReadable string `json:"idReadable"`
	}
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, fmt.Errorf("parse sprint issues: %w", err)
	}

	ids := make([]string, 0, len(refs))
	for _, r := range refs {
		if r.IDReadable != "" {
			ids = append(ids, r.IDReadable)
		}
	}
	return ids, nil
}

// AddIssueToSprint places an issue on a board's sprint. Adding an issue that is
// already on the sprint is a no-op on the server side.
func (c *Client) AddIssueToSprint(agileID, sprintID, idReadable string) error {
	path := fmt.Sprintf("/api/agiles/%s/sprints/%s/issues",
		url.PathEscape(agileID), url.PathEscape(sprintID))
	body := struct {
		IDReadable string `json:"idReadable"`
	}{IDReadable: idReadable}
	if err := c.post(path, body); err != nil {
		return fmt.Errorf("add %s to sprint: %w", idReadable, err)
	}
	return nil
}

// RemoveIssueFromSprint removes an issue from a board's sprint.
func (c *Client) RemoveIssueFromSprint(agileID, sprintID, issueID string) error {
	path := fmt.Sprintf("/api/agiles/%s/sprints/%s/issues/%s",
		url.PathEscape(agileID), url.PathEscape(sprintID), url.PathEscape(issueID))
	if err := c.delete(path); err != nil {
		return fmt.Errorf("remove %s from sprint: %w", issueID, err)
	}
	return nil
}

// IssueBoards derives which boards and sprints an issue is on. It lists boards
// whose projects include the issue's project, then scans those boards' sprints
// for the issue. Returns an empty slice when the issue is on no scanned board.
func (c *Client) IssueBoards(issueID string) ([]BoardMembership, error) {
	boards, err := c.ListBoards()
	if err != nil {
		return nil, err
	}

	prefix := projectPrefix(issueID)
	var out []BoardMembership
	for i := range boards {
		b := &boards[i]
		if prefix != "" && !b.HasProject(prefix) {
			continue
		}
		for _, s := range b.SprintList() {
			ids, err := c.ListSprintIssues(b.ID, s.ID)
			if err != nil {
				return nil, err
			}
			if containsFold(ids, issueID) {
				out = append(out, BoardMembership{Board: b.Name, Sprint: s.Name})
			}
		}
	}
	return out, nil
}

// projectPrefix extracts the project short name from a readable issue ID, e.g.
// "AX-812" -> "AX". Returns "" when the ID has no separator.
func projectPrefix(issueID string) string {
	if i := strings.LastIndex(issueID, "-"); i > 0 {
		return issueID[:i]
	}
	return ""
}

func containsFold(ids []string, target string) bool {
	for _, id := range ids {
		if strings.EqualFold(id, target) {
			return true
		}
	}
	return false
}
