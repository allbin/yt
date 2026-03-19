package format

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

// RenderMarkdown renders markdown text as styled terminal output.
// Falls back to plain text on error. Caches the renderer per width.
func RenderMarkdown(text string, width int) string {
	r := getRenderer(width)
	if r == nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return out
}

// SplitRendered splits glamour output into individual lines,
// trimming leading/trailing blank lines.
func SplitRendered(s string) []string {
	s = strings.Trim(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

var (
	rendererMu    sync.Mutex
	rendererCache = map[int]*glamour.TermRenderer{}
)

func getRenderer(width int) *glamour.TermRenderer {
	rendererMu.Lock()
	defer rendererMu.Unlock()

	if r, ok := rendererCache[width]; ok {
		return r
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}
	rendererCache[width] = r
	return r
}
