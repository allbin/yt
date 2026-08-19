package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSkill = `---
name: yt
description: "Interact with YouTrack: issues, boards."
argument-hint: <issue-id>
allowed-tools: Bash(yt *)
---

# YouTrack CLI

Body mentioning allowed-tools stays put.
`

// isolate empties $PATH so presence depends only on what the test creates.
func isolate(t *testing.T) string {
	t.Helper()
	t.Setenv("PATH", "")
	return t.TempDir()
}

// markPresent gives a target's config dir a file of its own, standing in for a
// real agent installation.
func markPresent(t *testing.T, home, configDir string) {
	t.Helper()
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func resultFor(t *testing.T, results []Result, name string) Result {
	t.Helper()
	for _, r := range results {
		if r.Name == name {
			return r
		}
	}
	t.Fatalf("no result for %q in %+v", name, results)
	return Result{}
}

func TestInstallSkipsAbsentAgents(t *testing.T) {
	home := isolate(t)
	markPresent(t, home, ".codex")

	results := Install(testSkill, home)

	codex := resultFor(t, results, "Codex")
	if codex.Status != StatusInstalled {
		t.Errorf("codex status = %q, want installed (%+v)", codex.Status, codex)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "skills", "yt", "SKILL.md")); err != nil {
		t.Errorf("codex skill not written: %v", err)
	}

	claude := resultFor(t, results, "Claude Code")
	if claude.Status != StatusSkipped {
		t.Errorf("claude status = %q, want skipped", claude.Status)
	}
	if claude.Reason == "" {
		t.Error("skipped result should carry a reason")
	}
	if _, err := os.Stat(filepath.Join(home, ".claude")); !os.IsNotExist(err) {
		t.Error("absent agent's config dir should not be created")
	}
}

func TestInstallReportsUnchangedThenUpdated(t *testing.T) {
	home := isolate(t)
	markPresent(t, home, ".codex")
	path := filepath.Join(home, ".codex", "skills", "yt", "SKILL.md")

	if got := resultFor(t, Install(testSkill, home), "Codex").Status; got != StatusInstalled {
		t.Fatalf("first install = %q, want installed", got)
	}
	if got := resultFor(t, Install(testSkill, home), "Codex").Status; got != StatusUnchanged {
		t.Errorf("second install = %q, want unchanged", got)
	}

	if err := os.WriteFile(path, []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resultFor(t, Install(testSkill, home), "Codex").Status; got != StatusUpdated {
		t.Errorf("install over stale content = %q, want updated", got)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) == "stale\n" {
		t.Error("stale skill was not overwritten")
	}
}

// A config dir holding nothing but the skill we installed is evidence of a past
// install, not of the agent -- otherwise the first install would make every
// later run report a phantom agent.
func TestOurOwnSkillIsNotProofOfPresence(t *testing.T) {
	home := isolate(t)
	skillDir := filepath.Join(home, ".codex", "skills", "yt")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(testSkill), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := resultFor(t, Install(testSkill, home), "Codex").Status; got != StatusSkipped {
		t.Errorf("status = %q, want skipped for a config dir holding only our skill", got)
	}

	// A second skill next to ours means something else put it there.
	if err := os.MkdirAll(filepath.Join(home, ".codex", "skills", "other"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := resultFor(t, Install(testSkill, home), "Codex").Status; !got.Written() {
		t.Errorf("status = %q, want an install once another skill is present", got)
	}
}

func TestInstallStripsFrontmatterPerTarget(t *testing.T) {
	home := isolate(t)
	markPresent(t, home, ".claude")
	markPresent(t, home, ".codex")

	Install(testSkill, home)

	codex, err := os.ReadFile(filepath.Join(home, ".codex", "skills", "yt", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	want := `---
name: yt
description: "Interact with YouTrack: issues, boards."
---

# YouTrack CLI

Body mentioning allowed-tools stays put.
`
	if string(codex) != want {
		t.Errorf("codex skill:\ngot:\n%s\nwant:\n%s", codex, want)
	}

	claude, err := os.ReadFile(filepath.Join(home, ".claude", "skills", "yt", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(claude) != testSkill {
		t.Error("claude skill should be installed verbatim")
	}
}

func TestInstallRemovesLegacyClaudeCommand(t *testing.T) {
	home := isolate(t)
	markPresent(t, home, ".claude")
	legacy := filepath.Join(home, ".claude", "commands", "yt.md")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	claude := resultFor(t, Install(testSkill, home), "Claude Code")
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Error("legacy command not removed")
	}
	if len(claude.Removed) != 1 || !strings.HasSuffix(claude.Removed[0], "yt.md") {
		t.Errorf("removed = %v, want the legacy command", claude.Removed)
	}
}

func TestInstallReportsWriteFailure(t *testing.T) {
	home := isolate(t)
	markPresent(t, home, ".codex")
	// A directory where SKILL.md belongs makes the write fail.
	if err := os.MkdirAll(filepath.Join(home, ".codex", "skills", "yt", "SKILL.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	codex := resultFor(t, Install(testSkill, home), "Codex")
	if codex.Status != StatusFailed {
		t.Errorf("status = %q, want failed", codex.Status)
	}
	if codex.Reason == "" {
		t.Error("failed result should carry a reason")
	}
	if codex.Status.Written() {
		t.Error("failed must not count as written")
	}
}

func TestFilterFrontmatter(t *testing.T) {
	doc := `---
name: yt
description: "a: b"
allowed-tools: Bash(yt *)
list:
  - one
  - two
---

# Body
`
	got, err := filterFrontmatter(doc, []string{"name", "list"})
	if err != nil {
		t.Fatal(err)
	}
	want := `---
name: yt
list:
  - one
  - two
---

# Body
`
	if got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}

	if got, err := filterFrontmatter(doc, nil); err != nil || got != doc {
		t.Errorf("nil keep list should pass the document through unchanged (err %v)", err)
	}
	if got, err := filterFrontmatter("no frontmatter\n", []string{"name"}); err != nil || got != "no frontmatter\n" {
		t.Errorf("document without frontmatter should be unchanged, got %q (err %v)", got, err)
	}
	if _, err := filterFrontmatter("---\nname: [unclosed\n---\n\nbody\n", []string{"name"}); err == nil {
		t.Error("malformed frontmatter should be an error, not silent output")
	}
}
