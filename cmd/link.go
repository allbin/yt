package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link <id> <relation> <target-id>...",
	Short: "Create links between issues",
	Long: `Link an issue to one or more target issues using a directed relation phrase.

The relation is matched against the instance's link types and is accepted in
kebab, spaced, or squashed form (e.g. "subtask-of", "subtask of", "subtaskof").
Run "yt link types" to list the available relations.

Linking is idempotent: a link that already exists is reported as such and left
unchanged.`,
	Example: `  # make AX-804 a subtask of AX-332
  yt link AX-804 subtask-of AX-332

  # relate two issues
  yt link AX-1 relates AX-2

  # declare a dependency
  yt link AX-1 depends-on AX-3

  # mark a duplicate
  yt link AX-1 duplicates AX-4

  # link to several targets at once
  yt link AX-1 relates AX-2 AX-3 AX-4

  # JSON output (the source issue's links after the change)
  yt link AX-804 subtask-of AX-332 --json`,
	Args:              cobra.MinimumNArgs(3),
	RunE:              runLink,
	ValidArgsFunction: completeRelation,
}

func init() {
	rootCmd.AddCommand(linkCmd)
}

func runLink(cmd *cobra.Command, args []string) error {
	sourceID, alias, targets := args[0], args[1], args[2:]

	client, err := apiFactory()
	if err != nil {
		return err
	}

	rel, err := resolveRelation(client, alias)
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(sourceID)
	if err != nil {
		return err
	}

	created := make(map[string]bool, len(targets))
	for _, t := range targets {
		if link, _ := youtrack.FindLink(issue.Links, rel, t); link != nil {
			created[t] = false
			continue
		}
		if err := client.CreateLink(sourceID, rel.Phrase, t); err != nil {
			return err
		}
		created[t] = true
	}

	issue, err = client.GetIssue(sourceID)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, issue.Links)
	}

	arrow := format.StyleDim.Render("→")
	for _, t := range targets {
		note := ""
		if !created[t] {
			note = " " + format.StyleDim.Render("(already linked)")
		}
		if _, err := fmt.Fprintf(w, "%s %s %s %s%s\n",
			format.StyleID.Render(sourceID), rel.Phrase, arrow, t, note); err != nil {
			return err
		}
	}
	return nil
}

// resolveRelation fetches the instance's link types and resolves a relation
// alias against them.
func resolveRelation(client youtrack.API, alias string) (youtrack.Relation, error) {
	types, err := client.ListLinkTypes()
	if err != nil {
		return youtrack.Relation{}, err
	}
	return youtrack.ResolveRelation(types, alias)
}

// completeRelation completes the relation argument (the second positional) with
// kebab-cased forms of the instance's directed phrases.
func completeRelation(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	client, err := apiFactory()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	types, err := client.ListLinkTypes()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var out []string
	for _, p := range youtrack.RelationPhrases(types) {
		kebab := strings.ReplaceAll(p, " ", "-")
		if strings.HasPrefix(kebab, toComplete) {
			out = append(out, kebab)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}
