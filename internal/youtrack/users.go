package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type User struct {
	Login    string `json:"login"`
	FullName string `json:"fullName"`
}

// CurrentUser returns the user the configured token authenticates as.
// Used to validate credentials during login.
func (c *Client) CurrentUser() (*User, error) {
	params := url.Values{"fields": {"login,fullName"}}

	data, err := c.get("/api/users/me", params)
	if err != nil {
		return nil, fmt.Errorf("current user: %w", err)
	}

	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}

	return &u, nil
}

// ResolveUser finds a user by query (login, full name, or partial match).
// Returns the login name for use in YouTrack queries.
// Passes through special values like "me" unchanged.
func (c *Client) ResolveUser(query string) (string, error) {
	if query == "me" || query == "unassigned" {
		return query, nil
	}

	params := url.Values{
		"fields": {"login,fullName"},
		"query":  {query},
	}

	data, err := c.get("/api/users", params)
	if err != nil {
		return "", fmt.Errorf("resolve user: %w", err)
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return "", fmt.Errorf("parse users: %w", err)
	}

	if len(users) == 0 {
		return "", fmt.Errorf("user %q not found", query)
	}

	return users[0].Login, nil
}
