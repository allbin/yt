package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

const projectFields = "id,shortName,name"

func (c *Client) ListProjects() ([]Project, error) {
	params := url.Values{"fields": {projectFields}}

	data, err := c.get("/api/admin/projects", params)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("parse projects: %w", err)
	}

	return projects, nil
}
