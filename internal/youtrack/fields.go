package youtrack

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
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

// projectFieldSchema selects a project custom field's schema, as served by an
// issue's projectCustomField.
const projectFieldSchema = "field(name,fieldType(id,isMultiValue)),bundle($type,id,name,values(name,ordinal))"

// globalFieldSchema selects every custom field with its per-project instances.
const globalFieldSchema = "name,fieldType(id,isMultiValue),instances(ordinal,project(id,shortName),bundle($type,id,name,values(name,ordinal)))"

// ListProjectFields returns a project's custom fields (project matched by
// short name or ID) in the project's order, with their allowed values. It
// reads the global field list because the per-project list is silently
// filtered to 200 [] for users who lack permission to read project settings.
// Finding no fields is an error: the token cannot tell an empty project from
// a hidden one.
func (c *Client) ListProjectFields(projectID string) ([]ProjectField, error) {
	params := url.Values{"fields": {globalFieldSchema}, "$top": {"1000"}}

	data, err := c.get("/api/admin/customFieldSettings/customFields", params)
	if err != nil {
		return nil, fmt.Errorf("fetch fields for project %s: %w", projectID, err)
	}

	var global []struct {
		rawField
		Instances []struct {
			Ordinal int `json:"ordinal"`
			Project *struct {
				ID        string `json:"id"`
				ShortName string `json:"shortName"`
			} `json:"project"`
			Bundle *rawBundle `json:"bundle"`
		} `json:"instances"`
	}
	if err := json.Unmarshal(data, &global); err != nil {
		return nil, fmt.Errorf("parse fields for project %s: %w", projectID, err)
	}

	var raw []rawProjectField
	for _, f := range global {
		for _, in := range f.Instances {
			if p := in.Project; p != nil && (p.ID == projectID || strings.EqualFold(p.ShortName, projectID)) {
				raw = append(raw, rawProjectField{Ordinal: in.Ordinal, Field: &f.rawField, Bundle: in.Bundle})
			}
		}
	}

	if len(raw) == 0 {
		if _, err := c.get("/api/admin/projects/"+url.PathEscape(projectID), url.Values{"fields": {"id"}}); err != nil {
			return nil, fmt.Errorf("fetch fields for project %s: %w", projectID, err)
		}
		return nil, fmt.Errorf("project %s has no custom fields visible to this token; it may have none, or the token may not be allowed to read them", projectID)
	}

	slices.SortStableFunc(raw, func(a, b rawProjectField) int { return cmp.Compare(a.Ordinal, b.Ordinal) })
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

// rawField is a custom field's definition as the API serves it.
type rawField struct {
	Name      string `json:"name"`
	FieldType *struct {
		ID           string `json:"id"`
		IsMultiValue bool   `json:"isMultiValue"`
	} `json:"fieldType"`
}

// rawBundle is the value bundle a field uses in one project.
type rawBundle struct {
	Type   string        `json:"$type"`
	ID     string        `json:"id"`
	Name   string        `json:"name"`
	Values []BundleValue `json:"values"`
}

// rawProjectField is a field as configured on one project. Ordinal is its
// position in the project's field order; only ListProjectFields sets it.
type rawProjectField struct {
	Ordinal int        `json:"ordinal"`
	Field   *rawField  `json:"field"`
	Bundle  *rawBundle `json:"bundle"`
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
