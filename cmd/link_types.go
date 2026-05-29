package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var linkTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List available issue link types",
	Long: `List the instance's issue link types and their directed phrases.

Use the listed phrases (in kebab, spaced, or squashed form) as the relation
argument to "yt link" and "yt unlink".`,
	Example: `  # list link types
  yt link types

  # JSON output
  yt link types --json`,
	Args: cobra.NoArgs,
	RunE: runLinkTypes,
}

func init() {
	linkCmd.AddCommand(linkTypesCmd)
}

func runLinkTypes(cmd *cobra.Command, args []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	types, err := client.ListLinkTypes()
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, types)
	}
	return format.LinkTypes(w, types)
}
