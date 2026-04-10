package cmd

import (
	"strings"

	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

// completeFieldValues returns a cobra completion function that queries
// the YouTrack API for allowed values of the named custom field.
// It resolves the issue ID from the command args.
func completeFieldValues(fieldName string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		client, err := apiFactory()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		values, err := client.GetFieldValues(args[0], fieldName)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return filterValues(values, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
}

// completeProjectFieldValues returns a cobra completion function that queries
// the YouTrack API for allowed values of a named custom field in the project
// specified by the -p/--project flag.
func completeProjectFieldValues(fieldName string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		project, _ := cmd.Flags().GetString("project")
		if project == "" {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		client, err := apiFactory()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		values, err := client.GetProjectFieldValues(project, fieldName)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return filterValues(values, toComplete), cobra.ShellCompDirectiveNoFileComp
	}
}

// completeFieldFlag returns a cobra completion function for --field "Name=Value".
// Before "=": completes field names. After "=": completes values for that field.
func completeFieldFlag(issueIDFromArgs bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		client, err := apiFactory()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		name, valuePrefix, hasEq := strings.Cut(toComplete, "=")

		if !hasEq {
			if !issueIDFromArgs || len(args) == 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			names, err := client.ListFieldNames(args[0])
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			var completions []string
			for _, n := range names {
				if strings.HasPrefix(strings.ToLower(n), strings.ToLower(toComplete)) {
					completions = append(completions, n+"=")
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		}

		var values []youtrack.BundleValue
		if issueIDFromArgs && len(args) > 0 {
			values, _ = client.GetFieldValues(args[0], name)
		} else {
			project, _ := cmd.Flags().GetString("project")
			if project != "" {
				values, _ = client.GetProjectFieldValues(project, name)
			}
		}

		var completions []string
		for _, v := range values {
			if strings.HasPrefix(strings.ToLower(v.Name), strings.ToLower(valuePrefix)) {
				completions = append(completions, name+"="+v.Name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}

// completeProjectNames returns project short names for tab completion.
func completeProjectNames(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	client, err := apiFactory()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	projects, err := client.ListProjects()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	lower := strings.ToLower(toComplete)
	for _, p := range projects {
		if strings.HasPrefix(strings.ToLower(p.ShortName), lower) {
			completions = append(completions, p.ShortName)
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

func filterValues(values []youtrack.BundleValue, prefix string) []string {
	var completions []string
	for _, v := range values {
		if strings.HasPrefix(strings.ToLower(v.Name), strings.ToLower(prefix)) {
			completions = append(completions, v.Name)
		}
	}
	return completions
}
