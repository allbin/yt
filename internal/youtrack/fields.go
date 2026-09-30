package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
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
	Bundle string        `json:"bundle,omitempty"`
	Values []BundleValue `json:"values,omitempty"`

	// ValueType is the field type without its cardinality, e.g. "enum",
	// "ownedField", "user", "date and time".
	ValueType string `json:"-"`
	Multi     bool   `json:"-"`
	// BundleID and BundleType ("EnumBundle", "OwnedBundle", ...) address the
	// bundle for admin writes.
	BundleID   string `json:"-"`
	BundleType string `json:"-"`
}

// projectFieldSchema selects a project custom field's schema. The same shape
// is served by a project's field list and by an issue's projectCustomField.
const projectFieldSchema = "field(name,fieldType(id,isMultiValue)),bundle($type,id,name,values(name,ordinal))"

// ListProjectFields returns all custom fields for a project with their
// allowed values.
func (c *Client) ListProjectFields(projectID string) ([]ProjectField, error) {
	params := url.Values{"fields": {projectFieldSchema}}

	path := "/api/admin/projects/" + url.PathEscape(projectID) + "/customFields"
	data, err := c.get(path, params)
	if err != nil {
		return nil, fmt.Errorf("fetch fields for project %s: %w", projectID, err)
	}

	var raw []rawProjectField
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse fields for project %s: %w", projectID, err)
	}
	return toProjectFields(raw), nil
}

// ListIssueFields returns the schema of every custom field on an issue, as
// configured on its project.
func (c *Client) ListIssueFields(issueID string) ([]ProjectField, error) {
	params := url.Values{"fields": {"projectCustomField(" + projectFieldSchema + ")"}}

	data, err := c.get("/api/issues/"+url.PathEscape(issueID)+"/customFields", params)
	if err != nil {
		return nil, fmt.Errorf("fetch fields for %s: %w", issueID, err)
	}

	var raw []struct {
		ProjectCustomField rawProjectField `json:"projectCustomField"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse fields for %s: %w", issueID, err)
	}
	fields := make([]rawProjectField, len(raw))
	for i, r := range raw {
		fields[i] = r.ProjectCustomField
	}
	return toProjectFields(fields), nil
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

type rawProjectField struct {
	Field *struct {
		Name      string `json:"name"`
		FieldType *struct {
			ID           string `json:"id"`
			IsMultiValue bool   `json:"isMultiValue"`
		} `json:"fieldType"`
	} `json:"field"`
	Bundle *struct {
		Type   string        `json:"$type"`
		ID     string        `json:"id"`
		Name   string        `json:"name"`
		Values []BundleValue `json:"values"`
	} `json:"bundle"`
}

func toProjectFields(raw []rawProjectField) []ProjectField {
	var result []ProjectField
	for _, r := range raw {
		if r.Field == nil {
			continue
		}
		pf := ProjectField{Name: r.Field.Name}
		if ft := r.Field.FieldType; ft != nil {
			pf.ValueType = strings.TrimSuffix(strings.TrimSuffix(ft.ID, "[1]"), "[*]")
			pf.Multi = ft.IsMultiValue
			pf.Type = friendlyFieldType(pf.ValueType, pf.Multi)
		}
		if b := r.Bundle; b != nil {
			pf.Bundle = b.Name
			pf.BundleID = b.ID
			pf.BundleType = b.Type
			pf.Values = b.Values
			sort.Slice(pf.Values, func(i, j int) bool {
				return pf.Values[i].Ordinal < pf.Values[j].Ordinal
			})
		}
		result = append(result, pf)
	}
	return result
}

// friendlyFieldType names a field type for display: "owned" for ownedField,
// with "[]" marking a multi-value field.
func friendlyFieldType(valueType string, multi bool) string {
	t := valueType
	if t == "ownedField" {
		t = "owned"
	}
	if multi {
		t += "[]"
	}
	return t
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
