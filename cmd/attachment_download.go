package cmd

import (
	"fmt"
	"os"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var attachmentOutput string

var attachmentDownloadCmd = &cobra.Command{
	Use:   "download <issueID> <filename>",
	Short: "Download an attachment from an issue",
	Long: `Download an attachment by filename from a YouTrack issue.
If multiple attachments share the same name, the first match is downloaded.`,
	Example: `  # download to current directory
  yt attachment download PROJ-123 report.csv

  # download to specific path
  yt attachment download PROJ-123 photo.png --output /tmp/photo.png`,
	Args: cobra.ExactArgs(2),
	RunE: runAttachmentDownload,
}

func init() {
	attachmentCmd.AddCommand(attachmentDownloadCmd)
	attachmentDownloadCmd.Flags().StringVarP(&attachmentOutput, "output", "o", "", "output file path (default: filename in current directory)")
}

func runAttachmentDownload(cmd *cobra.Command, args []string) (err error) {
	issueID, filename := args[0], args[1]

	client, err := apiFactory()
	if err != nil {
		return err
	}

	attachments, err := client.ListAttachments(issueID)
	if err != nil {
		return err
	}

	var dlURL string
	var size int64
	for _, a := range attachments {
		if a.Name == filename {
			dlURL = a.URL
			size = a.Size
			break
		}
	}
	if dlURL == "" {
		return fmt.Errorf("attachment %q not found on %s", filename, issueID)
	}

	outPath := filename
	if attachmentOutput != "" {
		outPath = attachmentOutput
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if dlErr := client.DownloadAttachment(dlURL, f); dlErr != nil {
		_ = f.Close()
		_ = os.Remove(outPath)
		return fmt.Errorf("download %s: %w", filename, dlErr)
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Downloaded %s (%s)\n", filename, format.FormatSize(size))
	return err
}
