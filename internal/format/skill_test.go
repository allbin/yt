package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/skill"
)

func TestSkillInstall(t *testing.T) {
	var buf bytes.Buffer
	err := SkillInstall(&buf, []skill.Result{
		{
			Name:    "Claude Code",
			Status:  skill.StatusInstalled,
			Path:    "/home/u/.claude/skills/yt/SKILL.md",
			Removed: []string{"/home/u/.claude/commands/yt.md"},
		},
		{Name: "Codex", Status: skill.StatusSkipped, Reason: "not found"},
	})
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines (two results plus one removal), got %d: %q", len(lines), buf.String())
	}
	if !strings.HasPrefix(lines[0], "installed  Claude Code  /home/u/.claude/skills/yt/SKILL.md") {
		t.Errorf("install line: %q", lines[0])
	}
	if !strings.Contains(lines[1], "removed legacy /home/u/.claude/commands/yt.md") {
		t.Errorf("removal line: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "skipped    Codex        not found") {
		t.Errorf("skip line should show the reason, padded to match: %q", lines[2])
	}
}
