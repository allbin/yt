package cmd

import "github.com/spf13/cobra"

var attachmentCmd = &cobra.Command{
	Use:   "attachment",
	Short: "Manage issue attachments",
	Long:  `Download and manage attachments on YouTrack issues.`,
}

func init() {
	rootCmd.AddCommand(attachmentCmd)
}
