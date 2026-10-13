package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var attachmentUploadName string

var attachmentUploadCmd = &cobra.Command{
	Use:   "upload <issueID> <file>...",
	Short: "Attach files to an issue",
	Long: `Upload one or more local files as attachments on a YouTrack issue. Each
attachment is named after its file's base name; --name stores a single file
under another name. All files are opened before anything is sent, so a
missing file uploads nothing.

Reference an uploaded image in a description or comment as ![](name.png).`,
	Example: `  # attach a screenshot
  yt attachment upload PROJ-123 screenshot.png

  # attach several files
  yt attachment upload PROJ-123 log.txt trace.json

  # store /tmp/out.csv as report.csv
  yt attachment upload PROJ-123 /tmp/out.csv --name report.csv

  # JSON output (id, name and size of each created attachment)
  yt attachment upload PROJ-123 screenshot.png --json`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAttachmentUpload,
}

func init() {
	attachmentCmd.AddCommand(attachmentUploadCmd)
	attachmentUploadCmd.Flags().StringVar(&attachmentUploadName, "name", "", "store the file under this name (one file only)")
}

func runAttachmentUpload(cmd *cobra.Command, args []string) (err error) {
	issueID, paths := args[0], args[1:]
	if attachmentUploadName != "" && len(paths) > 1 {
		return fmt.Errorf("--name applies to one file, got %d", len(paths))
	}

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
		name := filepath.Base(p)
		if attachmentUploadName != "" {
			name = attachmentUploadName
		}
		files = append(files, youtrack.UploadFile{Name: name, Content: f})
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
