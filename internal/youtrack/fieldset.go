package youtrack

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// FieldUpdate is one entry of an issue's customFields in a REST write. A nil
// Value clears a single-value field.
type FieldUpdate struct {
	Type  string `json:"$type"`
	Name  string `json:"name"`
	Value any    `json:"value"`
}

type namedValue struct {
	Name string `json:"name"`
}

type loginValue struct {
	Login string `json:"login"`
}

// bundleIssueTypes maps a bundle-backed value type to its issue field $type
// stem; Single or Multi is prefixed by cardinality.
var bundleIssueTypes = map[string]string{
	"enum":       "EnumIssueCustomField",
	"ownedField": "OwnedIssueCustomField",
	"version":    "VersionIssueCustomField",
	"build":      "BuildIssueCustomField",
	"group":      "GroupIssueCustomField",
	"user":       "UserIssueCustomField",
}

// Update builds the REST payload that sets f to values. Bundle values are
// matched case-insensitively against the bundle and sent by their canonical
// name; user values must already be logins. No values, or a single empty
// one, clears the field.
func (f ProjectField) Update(values []string) (FieldUpdate, error) {
	values = nonEmpty(values)
	if len(values) > 1 && !f.Multi {
		return FieldUpdate{}, fmt.Errorf("%s takes a single value, got %d: %s",
			f.Name, len(values), strings.Join(values, ", "))
	}

	u := FieldUpdate{Name: f.Name}
	if stem, ok := bundleIssueTypes[f.ValueType]; ok {
		u.Type = "Single" + stem
		if f.Multi {
			u.Type = "Multi" + stem
		}
		items := make([]any, 0, len(values))
		for _, v := range values {
			item, err := f.bundleItem(v)
			if err != nil {
				return FieldUpdate{}, err
			}
			items = append(items, item)
		}
		switch {
		case f.Multi:
			u.Value = items
		case len(items) == 1:
			u.Value = items[0]
		}
		return u, nil
	}

	var value string
	if len(values) == 1 {
		value = values[0]
	}
	switch f.ValueType {
	case "state":
		u.Type = "StateIssueCustomField"
		if value != "" {
			item, err := f.bundleItem(value)
			if err != nil {
				return FieldUpdate{}, err
			}
			u.Value = item
		}
	case "period":
		u.Type = "PeriodIssueCustomField"
		if value != "" {
			u.Value = map[string]string{"presentation": value}
		}
	case "text":
		u.Type = "TextIssueCustomField"
		if value != "" {
			u.Value = map[string]string{"text": value}
		}
	case "date", "date and time":
		// Date-only fields have their own $type; date-and-time fields are
		// simple fields holding a timestamp.
		u.Type = "DateIssueCustomField"
		if f.ValueType == "date and time" {
			u.Type = "SimpleIssueCustomField"
		}
		if value != "" {
			ms, err := parseDate(value, f.ValueType == "date")
			if err != nil {
				return FieldUpdate{}, fmt.Errorf("%s: %w", f.Name, err)
			}
			u.Value = ms
		}
	case "string":
		u.Type = "SimpleIssueCustomField"
		if value != "" {
			u.Value = value
		}
	case "integer":
		u.Type = "SimpleIssueCustomField"
		if value != "" {
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return FieldUpdate{}, fmt.Errorf("%s expects an integer, got %q", f.Name, value)
			}
			u.Value = n
		}
	case "float":
		u.Type = "SimpleIssueCustomField"
		if value != "" {
			n, err := strconv.ParseFloat(value, 64)
			if err != nil || math.IsNaN(n) || math.IsInf(n, 0) {
				return FieldUpdate{}, fmt.Errorf("%s expects a number, got %q", f.Name, value)
			}
			u.Value = n
		}
	default:
		return FieldUpdate{}, fmt.Errorf("%s: setting %q fields is not supported", f.Name, f.ValueType)
	}
	return u, nil
}

// bundleItem resolves v against the field's bundle. User fields carry a
// login and are not checked: a user bundle lists groups, not every member.
func (f ProjectField) bundleItem(v string) (any, error) {
	if f.ValueType == "user" {
		return loginValue{Login: v}, nil
	}
	if len(f.Values) == 0 {
		return namedValue{Name: v}, nil
	}
	for _, bv := range f.Values {
		if strings.EqualFold(bv.Name, v) {
			return namedValue{Name: bv.Name}, nil
		}
	}
	names := make([]string, len(f.Values))
	for i, bv := range f.Values {
		names[i] = bv.Name
	}
	return nil, &UnknownValueError{Field: f.Name, Value: v, Allowed: names}
}

// UnknownValueError reports a value that is not in a field's bundle.
type UnknownValueError struct {
	Field   string
	Value   string
	Allowed []string
}

func (e *UnknownValueError) Error() string {
	return fmt.Sprintf("%s has no value %q; allowed: %s", e.Field, e.Value, strings.Join(e.Allowed, ", "))
}

// parseDate turns a date into epoch milliseconds. A date-only field is stored
// at noon UTC, as YouTrack does, so no time zone shifts it to another day.
func parseDate(s string, dateOnly bool) (int64, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		if dateOnly {
			t = t.Add(12 * time.Hour)
		} else {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
		}
		return t.UnixMilli(), nil
	}
	if !dateOnly {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t.UnixMilli(), nil
		}
		if t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local); err == nil {
			return t.UnixMilli(), nil
		}
		return 0, fmt.Errorf("invalid date %q: use YYYY-MM-DD, \"YYYY-MM-DD HH:MM\" or RFC 3339", s)
	}
	return 0, fmt.Errorf("invalid date %q: use YYYY-MM-DD", s)
}

func nonEmpty(values []string) []string {
	var out []string
	for _, v := range values {
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
