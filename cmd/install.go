package cmd

import "github.com/spf13/cobra"

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install shell completions and Claude Code skill",
	Long: `Install supporting integrations for the yt CLI.

Subcommands install shell completions for tab-completion and a Claude Code
skill that lets Claude interact with YouTrack via this CLI.`,
}

func init() {
	rootCmd.AddCommand(installCmd)
}
