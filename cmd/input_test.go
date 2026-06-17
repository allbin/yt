package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTextArg(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "body.md")
	if err := os.WriteFile(file, []byte("from file\nline two"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		val   string
		stdin string
		want  string
	}{
		{"literal", "plain text", "", "plain text"},
		{"empty", "", "ignored", ""},
		{"stdin", "-", "from stdin\nmulti", "from stdin\nmulti"},
		{"file", "@" + file, "", "from file\nline two"},
		{"literal_at_short", "@", "", "@"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readTextArg(tt.val, strings.NewReader(tt.stdin))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("readTextArg(%q) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

func TestReadTextArgMissingFile(t *testing.T) {
	_, err := readTextArg("@/no/such/file/here.md", strings.NewReader(""))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "read ") {
		t.Errorf("unexpected error: %v", err)
	}
}
