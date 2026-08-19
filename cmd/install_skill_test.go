package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolateInstall points $HOME at a temp dir and empties $PATH so agent
// detection depends only on what the test creates.
func isolateInstall(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "")
	return home
}

func markAgentPresent(t *testing.T, home, configDir string) {
	t.Helper()
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunInstallSkillSummary(t *testing.T) {
	home := isolateInstall(t)
	markAgentPresent(t, home, ".codex")

	run := setupTest(t, &mockAPI{})
	out, err := run("install", "skill")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	codexSkill := filepath.Join(home, ".codex", "skills", "yt", "SKILL.md")
	if !strings.Contains(out, "installed") || !strings.Contains(out, codexSkill) {
		t.Errorf("summary missing codex install line: %s", out)
	}
	if !strings.Contains(out, "skipped") || !strings.Contains(out, "Claude Code") {
		t.Errorf("summary missing claude skip line: %s", out)
	}
}

func TestRunInstallSkillNoAgentFound(t *testing.T) {
	isolateInstall(t)

	run := setupTest(t, &mockAPI{})
	out, err := run("install", "skill")
	if err == nil {
		t.Fatal("expected an error when no agent is present")
	}
	if !strings.Contains(err.Error(), "no supported agent found") {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.Count(out, "skipped") != 2 {
		t.Errorf("expected every agent skipped: %s", out)
	}
}

func TestRunInstallSkillFailureExitsNonZero(t *testing.T) {
	home := isolateInstall(t)
	markAgentPresent(t, home, ".claude")
	markAgentPresent(t, home, ".codex")
	// A directory where Codex's SKILL.md belongs makes only that write fail.
	if err := os.MkdirAll(filepath.Join(home, ".codex", "skills", "yt", "SKILL.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	run := setupTest(t, &mockAPI{})
	out, err := run("install", "skill")
	if err == nil {
		t.Fatal("a failed write must not exit successfully")
	}
	if !strings.Contains(err.Error(), "Codex") {
		t.Errorf("error should name the failing agent, got %v", err)
	}
	if strings.Contains(err.Error(), "no supported agent found") {
		t.Errorf("write failure misreported as a missing agent: %v", err)
	}
	if !strings.Contains(out, "installed") || !strings.Contains(out, "failed") {
		t.Errorf("summary should show both outcomes: %s", out)
	}
}

// The embedded skill.md must survive per-agent frontmatter filtering.
func TestEmbeddedSkillInstallsForBothAgents(t *testing.T) {
	home := isolateInstall(t)
	markAgentPresent(t, home, ".claude")
	markAgentPresent(t, home, ".codex")

	run := setupTest(t, &mockAPI{})
	if _, err := run("install", "skill"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	codex, err := os.ReadFile(filepath.Join(home, ".codex", "skills", "yt", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	head, _, _ := strings.Cut(string(codex), "\n---\n")
	if strings.Contains(head, "allowed-tools") || strings.Contains(head, "argument-hint") {
		t.Errorf("codex frontmatter kept Claude-only keys:\n%s", head)
	}
	if !strings.Contains(head, "name: yt") || !strings.Contains(head, "description:") {
		t.Errorf("codex frontmatter lost required keys:\n%s", head)
	}

	claude, err := os.ReadFile(filepath.Join(home, ".claude", "skills", "yt", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(claude) != skillContent {
		t.Error("claude skill should be installed verbatim")
	}
}
