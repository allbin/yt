package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var projectFieldsAddCmd = &cobra.Command{
	Use:   "add <project> <field> <value>...",
	Short: "Add allowed values to a project field",
	Long: `Add values to the bundle behind a project's custom field, so they can be
set with --field. Works for enum, owned, version, build and state fields.

A value the field already has is reported and left unchanged. Bundles can be
shared by several projects: a value added here appears in every project that
uses the same bundle, which the output names. Requires permission to edit the
bundle.`,
	Example: `  # add a customer to the Customer field
  yt project fields add HK Customer LTVB

  # add several values at once
  yt project fields add HK Subsystem "Admin UI" Billing`,
	Args:              cobra.MinimumNArgs(3),
	RunE:              runProjectFieldsAdd,
	ValidArgsFunction: completeProjectFieldsAdd,
}

func init() {
	projectFieldsCmd.AddCommand(projectFieldsAddCmd)
}

func runProjectFieldsAdd(cmd *cobra.Command, args []string) error {
	project, name, values := args[0], args[1], args[2:]

	client, err := apiFactory()
	if err != nil {
		return err
	}

	fields, err := client.ListProjectFields(project)
	if err != nil {
		return err
	}
	field := findField(fields, name)
	if field == nil {
		return fmt.Errorf("project %s has no field %q; available: %s", project, name, fieldNames(fields))
	}

	w := cmd.OutOrStdout()
	for _, v := range values {
		v = strings.TrimSpace(v)
		note := ""
		if existing, ok := bundleValue(field.Values, v); ok {
			v = existing
			note = " " + format.StyleDim.Render("(already exists)")
		} else {
			if err := client.AddBundleValue(*field, v); err != nil {
				return err
			}
			field.Values = append(field.Values, youtrack.BundleValue{Name: v})
		}
		if _, err := fmt.Fprintf(w, "%s %s %s%s\n",
			format.StyleBold.Render(field.Name), format.StyleDim.Render("+"), v, note); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w, format.StyleDim.Render("bundle: "+field.Bundle))
	return err
}

// bundleValue finds name in values case-insensitively and returns its
// canonical spelling.
func bundleValue(values []youtrack.BundleValue, name string) (string, bool) {
	for _, v := range values {
		if strings.EqualFold(v.Name, name) {
			return v.Name, true
		}
	}
	return "", false
}

// completeProjectFieldsAdd completes the project, then its field names.
func completeProjectFieldsAdd(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return completeProjectNames(cmd, args, toComplete)
	case 1:
		client, err := apiFactory()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		fields, err := client.ListProjectFields(args[0])
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var out []string
		for _, f := range fields {
			if f.BundleID != "" && strings.HasPrefix(strings.ToLower(f.Name), strings.ToLower(toComplete)) {
				out = append(out, f.Name)
			}
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
