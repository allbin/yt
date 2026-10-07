package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// activityPageSize is how many activities one activitiesPage request asks for.
	activityPageSize = 100

	activityPageFields = "afterCursor,hasAfter,activities(id,timestamp,author(login,fullName)," +
		"target(idReadable,summary),field(name)," +
		"added(name,login,fullName,presentation,text),removed(name,login,fullName,presentation,text))"
)

// FieldActivityFilter narrows custom-field change activities. Zero values
// leave a filter off.
type FieldActivityFilter struct {
	Author     string // login, database ID, or "me"
	Start      time.Time
	End        time.Time
	IssueQuery string // YouTrack search syntax, e.g. "project: AX"
}

// FieldActivity is one change to an issue's custom field, as recorded in the
// issue's activity history.
type FieldActivity struct {
	ID        string          `json:"id"`
	Timestamp int64           `json:"timestamp"`
	Author    *User           `json:"author"`
	Target    *ActivityTarget `json:"target"`
	Field     *ActivityField  `json:"field"`
	Added     json.RawMessage `json:"added"`
	Removed   json.RawMessage `json:"removed"`
}

// ActivityTarget is the issue an activity changed.
type ActivityTarget struct {
	IDReadable string `json:"idReadable"`
	Summary    string `json:"summary"`
}

// ActivityField names the property an activity changed.
type ActivityField struct {
	Name string `json:"name"`
}

type activityPage struct {
	Activities  []FieldActivity `json:"activities"`
	AfterCursor string          `json:"afterCursor"`
	HasAfter    bool            `json:"hasAfter"`
}

// ListFieldActivities returns custom-field changes oldest first, following the
// activity cursor until the server reports no further page.
func (c *Client) ListFieldActivities(f FieldActivityFilter) ([]FieldActivity, error) {
	params := url.Values{
		"categories": {"CustomFieldCategory"},
		"fields":     {activityPageFields},
		"$top":       {strconv.Itoa(activityPageSize)},
	}
	if f.Author != "" {
		params.Set("author", f.Author)
	}
	if !f.Start.IsZero() {
		params.Set("start", strconv.FormatInt(f.Start.UnixMilli(), 10))
	}
	if !f.End.IsZero() {
		params.Set("end", strconv.FormatInt(f.End.UnixMilli(), 10))
	}
	if f.IssueQuery != "" {
		params.Set("issueQuery", f.IssueQuery)
	}

	var out []FieldActivity
	for {
		data, err := c.get("/api/activitiesPage", params)
		if err != nil {
			return nil, fmt.Errorf("list activities: %w", err)
		}

		var page activityPage
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("parse activities: %w", err)
		}
		out = append(out, page.Activities...)

		if !page.HasAfter {
			return out, nil
		}
		if page.AfterCursor == "" || page.AfterCursor == params.Get("cursor") {
			return nil, fmt.Errorf("list activities: pagination cursor did not advance")
		}
		params.Set("cursor", page.AfterCursor)
	}
}

// Transition is one change of an issue's field from one value to another,
// e.g. State Planned -> In Progress. FromColumn/ToColumn are set when the
// change is read against a board whose columns map the field's values.
type Transition struct {
	Issue      string `json:"issue"`
	Summary    string `json:"summary"`
	Author     *User  `json:"author"`
	Timestamp  int64  `json:"timestamp"`
	Time       string `json:"time"`
	Field      string `json:"field"`
	From       string `json:"from"`
	To         string `json:"to"`
	FromColumn string `json:"fromColumn,omitempty"`
	ToColumn   string `json:"toColumn,omitempty"`
}

// Transitions picks the changes to the named field (case-insensitive) out of
// a list of field activities, preserving order.
func Transitions(acts []FieldActivity, field string) []Transition {
	out := []Transition{}
	for _, a := range acts {
		if a.Target == nil || a.Field == nil || !strings.EqualFold(a.Field.Name, field) {
			continue
		}
		out = append(out, Transition{
			Issue:     a.Target.IDReadable,
			Summary:   a.Target.Summary,
			Author:    a.Author,
			Timestamp: a.Timestamp,
			Time:      time.UnixMilli(a.Timestamp).Format(time.RFC3339),
			Field:     a.Field.Name,
			From:      displayValue(a.Removed),
			To:        displayValue(a.Added),
		})
	}
	return out
}

// ColumnField returns the name of the field the board's columns are bound to,
// or "" when the board has no column settings.
func (a *Agile) ColumnField() string {
	if a.ColumnSettings == nil || a.ColumnSettings.Field == nil {
		return ""
	}
	return a.ColumnSettings.Field.Name
}

// ColumnOf returns the presentation of the board column holding the given
// field value, or "" when no column holds it.
func (a *Agile) ColumnOf(value string) string {
	if a.ColumnSettings == nil || value == "" {
		return ""
	}
	for _, col := range a.ColumnSettings.Columns {
		for _, v := range col.FieldValues {
			if strings.EqualFold(v.Name, value) {
				return col.Presentation
			}
		}
	}
	return ""
}
