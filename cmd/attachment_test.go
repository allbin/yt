package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunAttachmentUploadJSON(t *testing.T) {
	mock := &mockAPI{}
	run := setupTest(t, mock)

	out, err := run("attachment", "upload", "AX-1", writeTemp(t, "a.txt", "a"), "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got []youtrack.Attachment
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(got) != 1 || got[0].Name != "a.txt" {
		t.Errorf("got %+v", got)
	}
}

func TestRunAttachmentUploadName(t *testing.T) {
	mock := &mockAPI{}
	run := setupTest(t, mock)

	_, err := run("attachment", "upload", "AX-1", writeTemp(t, "out.csv", "x"), "--name", "report.csv")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"report.csv"}; !slices.Equal(mock.uploaded, want) {
		t.Errorf("uploaded = %q, want %q", mock.uploaded, want)
	}
}

func TestRunAttachmentUploadNameNeedsOneFile(t *testing.T) {
	mock := &mockAPI{}
	run := setupTest(t, mock)

	_, err := run("attachment", "upload", "AX-1", writeTemp(t, "a", "a"), writeTemp(t, "b", "b"), "--name", "x")
	if err == nil || !strings.Contains(err.Error(), "--name applies to one file") {
		t.Fatalf("err = %v, want --name rejection", err)
	}
	if len(mock.uploaded) != 0 {
		t.Errorf("uploaded %q despite the error", mock.uploaded)
	}
}

func TestRunAttachmentUpload(t *testing.T) {
	dir := t.TempDir()
	a, b := filepath.Join(dir, "shot.png"), filepath.Join(dir, "log.txt")
	for _, p := range []string{a, b} {
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	mock := &mockAPI{}
	run := setupTest(t, mock)

	out, err := run("attachment", "upload", "HK-1", a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"shot.png", "log.txt"}; !slices.Equal(mock.uploaded, want) {
		t.Errorf("uploaded = %q, want %q", mock.uploaded, want)
	}
	if !strings.Contains(out, "Uploaded shot.png") {
		t.Errorf("output: %s", out)
	}
}

func TestRunAttachmentUploadMissingFileUploadsNothing(t *testing.T) {
	mock := &mockAPI{}
	run := setupTest(t, mock)

	if _, err := run("attachment", "upload", "HK-1", filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("expected error for missing file")
	}
	if mock.uploaded != nil {
		t.Errorf("uploaded %q despite the missing file", mock.uploaded)
	}
}
