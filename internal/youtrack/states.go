package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
)

const stateFields = "name,value(name),projectCustomField(bundle(values(name,ordinal,isResolved)))"

func (c *Client) GetIssueStates(issueID string) ([]StateBundleElement, error) {
	params := url.Values{"fields": {stateFields}}

	data, err := c.get(fmt.Sprintf("/api/issues/%s/customFields", issueID), params)
	if err != nil {
		return nil, fmt.Errorf("fetch states for %s: %w", issueID, err)
	}

	var fields []struct {
		Name               string `json:"name"`
		ProjectCustomField *struct {
			Bundle *struct {
				Values []StateBundleElement `json:"values"`
			} `json:"bundle"`
		} `json:"projectCustomField"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("parse states for %s: %w", issueID, err)
	}

	for _, f := range fields {
		if f.Name != "State" {
			continue
		}
		if f.ProjectCustomField == nil || f.ProjectCustomField.Bundle == nil {
			return nil, fmt.Errorf("no state bundle found for %s", issueID)
		}
		states := f.ProjectCustomField.Bundle.Values
		sort.Slice(states, func(i, j int) bool {
			return states[i].Ordinal < states[j].Ordinal
		})
		return states, nil
	}

	return nil, fmt.Errorf("no State field found for %s", issueID)
}

func (c *Client) SetIssueState(issueID, stateName string) error {
	body := map[string]any{
		"customFields": []map[string]any{
			{
				"$type": "StateIssueCustomField",
				"name":  "State",
				"value": map[string]any{
					"name": stateName,
				},
			},
		},
	}
	if err := c.post(fmt.Sprintf("/api/issues/%s", issueID), body); err != nil {
		return fmt.Errorf("set state on %s: %w", issueID, err)
	}
	return nil
}
