package youtrack

import (
	"encoding/json"
	"strings"
)

type Issue struct {
	IDReadable   string        `json:"idReadable"`
	Summary      string        `json:"summary"`
	Description  *string       `json:"description"`
	Resolved     *int64        `json:"resolved"`
	Created      int64         `json:"created"`
	Updated      int64         `json:"updated"`
	Tags         []Tag         `json:"tags"`
	CustomFields []CustomField `json:"customFields"`
}

type Tag struct {
	Name string `json:"name"`
}

type CustomField struct {
	Name  string          `json:"name"`
	Value json.RawMessage `json:"value"`
}

// DisplayValue returns a human-readable string for the field value.
// Handles single objects (enum/user), arrays (multi-value), and plain strings.
func (cf *CustomField) DisplayValue() string {
	if len(cf.Value) == 0 || string(cf.Value) == "null" {
		return ""
	}

	var obj struct {
		Name         string `json:"name"`
		FullName     string `json:"fullName"`
		Login        string `json:"login"`
		Text         string `json:"text"`
		Presentation string `json:"presentation"`
	}
	if err := json.Unmarshal(cf.Value, &obj); err == nil {
		switch {
		case obj.Name != "":
			return obj.Name
		case obj.FullName != "":
			return obj.FullName
		case obj.Presentation != "":
			return obj.Presentation
		case obj.Text != "":
			return obj.Text
		case obj.Login != "":
			return obj.Login
		}
	}

	var arr []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(cf.Value, &arr); err == nil && len(arr) > 0 {
		names := make([]string, 0, len(arr))
		for _, item := range arr {
			if item.Name != "" {
				names = append(names, item.Name)
			}
		}
		return strings.Join(names, ", ")
	}

	var s string
	if err := json.Unmarshal(cf.Value, &s); err == nil {
		return s
	}

	return ""
}

// Field returns the display value of the named custom field.
func (i *Issue) Field(name string) string {
	for _, cf := range i.CustomFields {
		if cf.Name == name {
			return cf.DisplayValue()
		}
	}
	return ""
}

// TagNames returns a comma-separated string of tag names.
func (i *Issue) TagNames() string {
	if len(i.Tags) == 0 {
		return ""
	}
	names := make([]string, len(i.Tags))
	for idx, t := range i.Tags {
		names[idx] = t.Name
	}
	return strings.Join(names, ", ")
}

// Desc returns the description or empty string if nil.
func (i *Issue) Desc() string {
	if i.Description == nil {
		return ""
	}
	return *i.Description
}

type StateBundleElement struct {
	Name       string `json:"name"`
	Ordinal    int    `json:"ordinal"`
	IsResolved bool   `json:"isResolved"`
}

type Comment struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Author  *User  `json:"author"`
	Created int64  `json:"created"`
	Updated *int64 `json:"updated"`
}

type Agile struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Projects      []Project `json:"projects"`
	CurrentSprint *Sprint   `json:"currentSprint"`
	Sprints       []Sprint  `json:"sprints,omitempty"`
}

type Project struct {
	ID        string `json:"id,omitempty"`
	ShortName string `json:"shortName"`
	Name      string `json:"name,omitempty"`
}

type Sprint struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Start  *int64  `json:"start,omitempty"`
	Finish *int64  `json:"finish,omitempty"`
	Issues []Issue `json:"issues,omitempty"`
}
