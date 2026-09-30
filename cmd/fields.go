package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/pflag"
)

// fieldArg is one "Name=Value" custom field assignment.
type fieldArg struct {
	name, value string
}

// parseFieldArgs parses --field values of the form "Name=Value".
func parseFieldArgs(raw []string) ([]fieldArg, error) {
	args := make([]fieldArg, 0, len(raw))
	for _, f := range raw {
		name, value, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --field format %q: expected Name=Value", f)
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, fmt.Errorf("invalid --field format %q: name must be non-empty", f)
		}
		args = append(args, fieldArg{name: name, value: strings.TrimSpace(value)})
	}
	return args, nil
}

// fieldUpdates resolves assignments against a field schema into REST writes.
// Assignments to the same field collect into one value list, so a
// multi-value field is set by repeating it; user values resolve to logins.
func fieldUpdates(client youtrack.API, schema []youtrack.ProjectField, args []fieldArg) ([]youtrack.FieldUpdate, error) {
	var order []*youtrack.ProjectField
	values := make(map[*youtrack.ProjectField][]string)
	for _, a := range args {
		f := findField(schema, a.name)
		if f == nil {
			return nil, fmt.Errorf("no field %q; available: %s", a.name, fieldNames(schema))
		}
		if _, seen := values[f]; !seen {
			order = append(order, f)
		}
		values[f] = append(values[f], a.value)
	}

	updates := make([]youtrack.FieldUpdate, 0, len(order))
	for _, f := range order {
		vs := values[f]
		if f.ValueType == "user" {
			var err error
			if vs, err = resolveLogins(client, vs); err != nil {
				return nil, err
			}
		}
		u, err := f.Update(vs)
		if err != nil {
			return nil, err
		}
		updates = append(updates, u)
	}
	return updates, nil
}

// resolveLogins maps user values to logins: "me" is the token's user,
// "unassigned" clears the field, anything else is looked up.
func resolveLogins(client youtrack.API, values []string) ([]string, error) {
	logins := make([]string, 0, len(values))
	for _, v := range values {
		switch v {
		case "", "unassigned":
			continue
		case "me":
			u, err := client.CurrentUser()
			if err != nil {
				return nil, err
			}
			logins = append(logins, u.Login)
		default:
			login, err := client.ResolveUser(v)
			if err != nil {
				return nil, err
			}
			logins = append(logins, login)
		}
	}
	return logins, nil
}

func findField(schema []youtrack.ProjectField, name string) *youtrack.ProjectField {
	for i := range schema {
		if strings.EqualFold(schema[i].Name, name) {
			return &schema[i]
		}
	}
	return nil
}

func fieldNames(schema []youtrack.ProjectField) string {
	names := make([]string, len(schema))
	for i, f := range schema {
		names[i] = f.Name
	}
	return strings.Join(names, ", ")
}

// flagFieldArgs turns the dedicated field flags that were passed (--type,
// --subsystem, ...) into assignments, in the order given. A flag passed as ""
// clears its field. Each pair is {flag name, field name}.
func flagFieldArgs(flags *pflag.FlagSet, pairs ...[2]string) []fieldArg {
	var args []fieldArg
	for _, p := range pairs {
		if !flags.Changed(p[0]) {
			continue
		}
		v, _ := flags.GetString(p[0])
		args = append(args, fieldArg{name: p[1], value: strings.TrimSpace(v)})
	}
	return args
}

// tagCommand builds the command that adds and removes tags. Tag names stay
// bare: YouTrack keeps braces as part of a new tag's name and rejects them on
// untag, while a bare multi-word name parses up to the next keyword.
func tagCommand(tags, removeTags []string) string {
	var parts []string
	for _, t := range tags {
		parts = append(parts, "tag "+t)
	}
	for _, t := range removeTags {
		parts = append(parts, "untag "+t)
	}
	return strings.Join(parts, " ")
}
