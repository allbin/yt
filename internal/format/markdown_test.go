package format

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	out := RenderMarkdown("**bold** text", 80)
	if !strings.Contains(out, "bold") {
		t.Errorf("expected rendered output to contain 'bold', got %q", out)
	}
}

func TestRenderMarkdownFallback(t *testing.T) {
	// Width 0 should still produce output (fallback to plain text at worst)
	out := RenderMarkdown("hello", 0)
	if !strings.Contains(out, "hello") {
		t.Errorf("expected fallback to contain 'hello', got %q", out)
	}
}

func TestSplitRendered(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"newlines only", "\n\n\n", 0},
		{"single line", "\nhello\n", 1},
		{"multi line", "\nline1\nline2\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitRendered(tt.input)
			if len(got) != tt.want {
				t.Errorf("SplitRendered() returned %d lines, want %d", len(got), tt.want)
			}
		})
	}
}

func TestRendererCaching(t *testing.T) {
	// Call twice with same width — should hit cache
	r1 := RenderMarkdown("test", 80)
	r2 := RenderMarkdown("test", 80)
	if r1 != r2 {
		t.Error("expected identical output from cached renderer")
	}
}
