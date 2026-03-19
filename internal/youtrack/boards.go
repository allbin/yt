package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const (
	agileFields    = "id,name,projects(shortName),currentSprint(id,name)"
	agileDetail    = "id,name,projects(shortName),currentSprint(id,name),sprints(id,name)"
	agileBoardView = "id,name,projects(shortName)," +
		"currentSprint(id,name),sprints(id,name)," +
		"columnSettings(field(name),columns(presentation,ordinal,fieldValues(name,isResolved)))," +
		"swimlaneSettings(enabled,field(name),values(name))"
)

func (c *Client) ListBoards() ([]Agile, error) {
	params := url.Values{"fields": {agileFields}}

	data, err := c.get("/api/agiles", params)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}

	var boards []Agile
	if err := json.Unmarshal(data, &boards); err != nil {
		return nil, fmt.Errorf("parse boards: %w", err)
	}

	return boards, nil
}

// GetBoardByName finds a board by case-insensitive name match.
// Returns the board with its sprint list populated for sprint lookup.
func (c *Client) GetBoardByName(name string) (*Agile, error) {
	params := url.Values{"fields": {agileDetail}}

	data, err := c.get("/api/agiles", params)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}

	var boards []Agile
	if err := json.Unmarshal(data, &boards); err != nil {
		return nil, fmt.Errorf("parse boards: %w", err)
	}

	for _, b := range boards {
		if strings.EqualFold(b.Name, name) {
			return &b, nil
		}
	}

	return nil, fmt.Errorf("board %q not found", name)
}

// GetBoardForView finds a board by name with expanded column/swimlane settings.
func (c *Client) GetBoardForView(name string) (*Agile, error) {
	params := url.Values{"fields": {agileBoardView}}

	data, err := c.get("/api/agiles", params)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}

	var boards []Agile
	if err := json.Unmarshal(data, &boards); err != nil {
		return nil, fmt.Errorf("parse boards: %w", err)
	}

	for _, b := range boards {
		if strings.EqualFold(b.Name, name) {
			return &b, nil
		}
	}

	return nil, fmt.Errorf("board %q not found", name)
}

