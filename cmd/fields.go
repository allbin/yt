package cmd

import (
	"fmt"
	"slices"
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

// fieldChanges are custom field edits from the command line: set replaces a
// field's value, add and remove edit a multi-value field's current list.
type fieldChanges struct {
	set, add, remove []fieldArg
}

func (c fieldChanges) empty() bool {
	return len(c.set) == 0 && len(c.add) == 0 && len(c.remove) == 0
}

// fieldUpdates resolves changes against a field schema into REST writes.
// Assignments to the same field collect into one value list, so a
// multi-value field is set by repeating it; user values resolve to logins.
// current returns a field's present values and is only called for add and
// remove.
func fieldUpdates(client youtrack.API, schema []youtrack.ProjectField, changes fieldChanges, current func(field string) []string) ([]youtrack.FieldUpdate, error) {
	var order []*youtrack.ProjectField
	edits := make(map[*youtrack.ProjectField]*fieldChanges)
	group := func(args []fieldArg, list func(*fieldChanges) *[]fieldArg) error {
		for _, a := range args {
			f := findField(schema, a.name)
			if f == nil {
				return fmt.Errorf("no field %q; available: %s", a.name, fieldNames(schema))
			}
			e := edits[f]
			if e == nil {
				e = &fieldChanges{}
				edits[f] = e
				order = append(order, f)
			}
			*list(e) = append(*list(e), a)
		}
		return nil
	}
	if err := group(changes.set, func(e *fieldChanges) *[]fieldArg { return &e.set }); err != nil {
		return nil, err
	}
	if err := group(changes.add, func(e *fieldChanges) *[]fieldArg { return &e.add }); err != nil {
		return nil, err
	}
	if err := group(changes.remove, func(e *fieldChanges) *[]fieldArg { return &e.remove }); err != nil {
		return nil, err
	}

	updates := make([]youtrack.FieldUpdate, 0, len(order))
	for _, f := range order {
		values, err := fieldValues(client, f, edits[f], current)
		if err != nil {
			return nil, err
		}
		u, err := f.Update(values)
		if err != nil {
			return nil, err
		}
		updates = append(updates, u)
	}
	return updates, nil
}

// fieldValues computes a field's new value list: the set values, or the
// current values with add and remove applied, compared case-insensitively.
func fieldValues(client youtrack.API, f *youtrack.ProjectField, e *fieldChanges, current func(string) []string) ([]string, error) {
	if len(e.add) == 0 && len(e.remove) == 0 {
		return userLogins(client, f, e.set)
	}
	if len(e.set) > 0 {
		return nil, fmt.Errorf("%s: use either --field or --add-field/--remove-field, not both", f.Name)
	}
	if !f.Multi {
		return nil, fmt.Errorf("%s takes a single value; set it with --field", f.Name)
	}
	add, err := userLogins(client, f, e.add)
	if err != nil {
		return nil, err
	}
	remove, err := userLogins(client, f, e.remove)
	if err != nil {
		return nil, err
	}
	contains := func(list []string, v string) bool {
		return slices.ContainsFunc(list, func(x string) bool { return strings.EqualFold(x, v) })
	}
	var result []string
	for _, v := range current(f.Name) {
		if !contains(remove, v) {
			result = append(result, v)
		}
	}
	for _, v := range add {
		if !contains(result, v) {
			result = append(result, v)
		}
	}
	return result, nil
}

// userLogins returns the args' values, resolved to logins on a user field.
func userLogins(client youtrack.API, f *youtrack.ProjectField, args []fieldArg) ([]string, error) {
	values := make([]string, len(args))
	for i, a := range args {
		values[i] = a.value
	}
	if f.ValueType != "user" {
		return values, nil
	}
	return resolveLogins(client, values)
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
