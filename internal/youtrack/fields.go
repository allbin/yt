package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
)

// BundleValue represents a single value in a custom field bundle.
type BundleValue struct {
	Name    string `json:"name"`
	Ordinal int    `json:"ordinal"`
}

const bundleFields = "name,value(name),projectCustomField(bundle(values(name,ordinal)))"

// GetFieldValues returns the allowed bundle values for a named custom field
// on the given issue. Works for enum, state, owned, and version bundle fields.
func (c *Client) GetFieldValues(issueID, fieldName string) ([]BundleValue, error) {
	params := url.Values{"fields": {bundleFields}}

	data, err := c.get("/api/issues/"+url.PathEscape(issueID)+"/customFields", params)
	if err != nil {
		return nil, fmt.Errorf("fetch fields for %s: %w", issueID, err)
	}

	return extractBundleValues(data, fieldName, issueID)
}

// ProjectField represents a custom field configured on a project,
// including its allowed bundle values (if any).
type ProjectField struct {
	Name   string        `json:"name"`
	Type   string        `json:"type,omitempty"`
	Values []BundleValue `json:"values,omitempty"`
}

const projectBundleFields = "field(name,fieldType($type)),bundle(values(name,ordinal))"

// ListProjectFields returns all custom fields for a project with their
// allowed values.
func (c *Client) ListProjectFields(projectID string) ([]ProjectField, error) {
	params := url.Values{"fields": {projectBundleFields}}

	path := "/api/admin/projects/" + url.PathEscape(projectID) + "/customFields"
	data, err := c.get(path, params)
	if err != nil {
		return nil, fmt.Errorf("fetch fields for project %s: %w", projectID, err)
	}

	return parseProjectFields(data, projectID)
}

// GetProjectFieldValues returns the allowed bundle values for a named custom
// field in the given project.
func (c *Client) GetProjectFieldValues(projectID, fieldName string) ([]BundleValue, error) {
	fields, err := c.ListProjectFields(projectID)
	if err != nil {
		return nil, err
	}
	for _, f := range fields {
		if f.Name == fieldName {
			return f.Values, nil
		}
	}
	return nil, nil
}

func parseProjectFields(data []byte, context string) ([]ProjectField, error) {
	var raw []struct {
		Field *struct {
			Name      string  `json:"name"`
			FieldType *struct {
				Type string `json:"$type"`
			} `json:"fieldType"`
		} `json:"field"`
		Bundle *struct {
			Values []BundleValue `json:"values"`
		} `json:"bundle"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse fields for project %s: %w", context, err)
	}

	var result []ProjectField
	for _, r := range raw {
		if r.Field == nil {
			continue
		}
		pf := ProjectField{Name: r.Field.Name}
		if r.Field.FieldType != nil {
			pf.Type = friendlyFieldType(r.Field.FieldType.Type)
		}
		if r.Bundle != nil {
			pf.Values = r.Bundle.Values
			sort.Slice(pf.Values, func(i, j int) bool {
				return pf.Values[i].Ordinal < pf.Values[j].Ordinal
			})
		}
		result = append(result, pf)
	}
	return result, nil
}

// friendlyFieldType maps YouTrack $type values to short human-readable names.
func friendlyFieldType(t string) string {
	switch t {
	case "StateIssueCustomField", "StateBundleCustomFieldDefaults":
		return "state"
	case "EnumBundleCustomFieldDefaults", "SingleEnumIssueCustomField":
		return "enum"
	case "OwnedBundleCustomFieldDefaults", "SingleOwnedIssueCustomField":
		return "owned"
	case "VersionBundleCustomFieldDefaults", "SingleVersionIssueCustomField":
		return "version"
	case "UserCustomFieldDefaults", "SingleUserIssueCustomField":
		return "user"
	case "BuildBundleCustomFieldDefaults", "SingleBuildIssueCustomField":
		return "build"
	case "PeriodIssueCustomField":
		return "period"
	case "DateIssueCustomField":
		return "date"
	case "TextIssueCustomField":
		return "text"
	case "SimpleIssueCustomField":
		return "simple"
	case "MultiBundleCustomFieldDefaults", "MultiEnumIssueCustomField":
		return "enum[]"
	case "MultiOwnedIssueCustomField":
		return "owned[]"
	case "MultiVersionIssueCustomField":
		return "version[]"
	case "MultiUserIssueCustomField":
		return "user[]"
	default:
		return t
	}
}

// ListFieldNames returns the names of all custom fields on the given issue.
func (c *Client) ListFieldNames(issueID string) ([]string, error) {
	params := url.Values{"fields": {"name"}}

	data, err := c.get("/api/issues/"+url.PathEscape(issueID)+"/customFields", params)
	if err != nil {
		return nil, fmt.Errorf("fetch field names for %s: %w", issueID, err)
	}

	var fields []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("parse field names for %s: %w", issueID, err)
	}

	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	return names, nil
}

func extractBundleValues(data []byte, fieldName, context string) ([]BundleValue, error) {
	var fields []struct {
		Name               string `json:"name"`
		ProjectCustomField *struct {
			Bundle *struct {
				Values []BundleValue `json:"values"`
			} `json:"bundle"`
		} `json:"projectCustomField"`
	}
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("parse fields for %s: %w", context, err)
	}

	for _, f := range fields {
		if f.Name != fieldName {
			continue
		}
		if f.ProjectCustomField == nil || f.ProjectCustomField.Bundle == nil {
			return nil, nil
		}
		values := f.ProjectCustomField.Bundle.Values
		sort.Slice(values, func(i, j int) bool { return values[i].Ordinal < values[j].Ordinal })
		return values, nil
	}

	return nil, nil
}
