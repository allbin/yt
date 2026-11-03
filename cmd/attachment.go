package cmd

import "github.com/spf13/cobra"

var attachmentCmd = &cobra.Command{
	Use:   "attachment",
	Short: "Manage issue and article attachments",
	Long: `Upload and download attachments on YouTrack issues and knowledge base
articles. Every subcommand takes an issue ID (PROJ-123) or an article ID
(PROJ-A-12).`,
}

func init() {
	rootCmd.AddCommand(attachmentCmd)
}
