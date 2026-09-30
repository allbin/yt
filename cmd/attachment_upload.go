package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var attachmentUploadCmd = &cobra.Command{
	Use:   "upload <issueID> <file>...",
	Short: "Attach files to an issue",
	Long: `Upload one or more local files as attachments on a YouTrack issue. Each
attachment is named after its file's base name. All files are sent in one
request; nothing is uploaded if a file cannot be opened.

Reference an uploaded image in a description or comment as ![](name.png).`,
	Example: `  # attach a screenshot
  yt attachment upload PROJ-123 screenshot.png

  # attach several files
  yt attachment upload PROJ-123 log.txt trace.json

  # JSON output (the created attachments)
  yt attachment upload PROJ-123 screenshot.png --json`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAttachmentUpload,
}

func init() {
	attachmentCmd.AddCommand(attachmentUploadCmd)
}

func runAttachmentUpload(cmd *cobra.Command, args []string) (err error) {
	issueID, paths := args[0], args[1:]

	client, err := apiFactory()
	if err != nil {
		return err
	}

	var opened []*os.File
	defer func() {
		for _, f := range opened {
			if cerr := f.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}
	}()
	files := make([]youtrack.UploadFile, 0, len(paths))
	for _, p := range paths {
		f, openErr := os.Open(p)
		if openErr != nil {
			return openErr
		}
		opened = append(opened, f)
		files = append(files, youtrack.UploadFile{Name: filepath.Base(p), Content: f})
	}

	attachments, err := client.UploadAttachments(issueID, files)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, attachments)
	}
	for _, a := range attachments {
		if _, err := fmt.Fprintf(w, "Uploaded %s (%s) to %s\n", a.Name, format.FormatSize(a.Size), issueID); err != nil {
			return err
		}
	}
	return nil
}
