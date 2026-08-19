package cmd

import "github.com/spf13/cobra"

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install shell completions and agent skills",
	Long: `Install supporting integrations for the yt CLI.

Subcommands install shell completions for tab-completion and the yt skill for
every supported coding agent found on this machine (Claude Code, Codex), so the
agent can interact with YouTrack via this CLI.`,
}

func init() {
	rootCmd.AddCommand(installCmd)
}
