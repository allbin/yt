// Package skill installs the yt skill definition for coding agents that read
// SKILL.md files, such as Claude Code and Codex.
package skill

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Dir is the skill's directory name inside each agent's skills directory.
const Dir = "yt"

// Target is an agent that reads SKILL.md-style skill definitions.
type Target struct {
	Name string // human label, shown in the summary

	configDir string   // agent config dir, relative to $HOME
	binary    string   // executable that proves the agent is installed
	keep      []string // frontmatter keys this agent understands; nil keeps all
	legacy    []string // stale files (relative to $HOME) removed on install
}

// Targets are ordered as they appear in the install summary.
func Targets() []Target {
	return []Target{
		{
			Name:      "Claude Code",
			configDir: ".claude",
			binary:    "claude",
			keep:      nil, // Claude Code understands every key in skill.md
			legacy:    []string{filepath.Join(".claude", "commands", "yt.md")},
		},
		{
			Name:      "Codex",
			configDir: ".codex",
			binary:    "codex",
			keep:      []string{"name", "description"},
		},
	}
}

// Status is the outcome of installing for one target.
type Status string

const (
	StatusInstalled Status = "installed" // no skill file was there before
	StatusUpdated   Status = "updated"   // file existed with different content
	StatusUnchanged Status = "unchanged" // file already had the exact content
	StatusSkipped   Status = "skipped"   // agent not present on this machine
	StatusFailed    Status = "failed"    // write error
)

// Written reports whether the target ended up holding the current skill.
func (s Status) Written() bool {
	return s == StatusInstalled || s == StatusUpdated || s == StatusUnchanged
}

// Result is the outcome for one target.
type Result struct {
	Name    string   `json:"name"`
	Status  Status   `json:"status"`
	Path    string   `json:"path,omitempty"`
	Reason  string   `json:"reason,omitempty"`
	Removed []string `json:"removed,omitempty"`
}

// Install writes content as the yt skill for every target present under home,
// returning one result per target in Targets() order. Absent agents are
// reported as skipped rather than installed into.
func Install(content, home string) []Result {
	targets := Targets()
	results := make([]Result, 0, len(targets))
	for _, t := range targets {
		results = append(results, t.install(content, home))
	}
	return results
}

func (t Target) install(content, home string) Result {
	res := Result{Name: t.Name}
	if !t.present(home) {
		res.Status = StatusSkipped
		res.Reason = fmt.Sprintf("not found (no ~/%s and no %q on PATH)", t.configDir, t.binary)
		return res
	}

	res.Path = t.skillPath(home)
	want, err := filterFrontmatter(content, t.keep)
	if err != nil {
		res.Status = StatusFailed
		res.Reason = err.Error()
		return res
	}

	res.Status, err = writeIfChanged(res.Path, want)
	if err != nil {
		res.Status = StatusFailed
		res.Reason = err.Error()
		return res
	}

	res.Removed = t.removeLegacy(home)
	return res
}

func (t Target) skillPath(home string) string {
	return filepath.Join(home, t.configDir, "skills", Dir, "SKILL.md")
}

// present reports whether the agent is installed on this machine. The config
// dir alone is not proof: installing the skill creates it, so a dir holding
// nothing but our own skill is evidence of a past install, not of the agent.
func (t Target) present(home string) bool {
	if _, err := exec.LookPath(t.binary); err == nil {
		return true
	}
	return hasEntryBesides(filepath.Join(home, t.configDir), "skills") ||
		hasEntryBesides(filepath.Join(home, t.configDir, "skills"), Dir)
}

// hasEntryBesides reports whether dir holds any entry other than except.
func hasEntryBesides(dir, except string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.Name() != except {
			return true
		}
	}
	return false
}

// writeIfChanged writes want to path unless it is already there byte for byte,
// reporting which of the three happened.
func writeIfChanged(path, want string) (Status, error) {
	status := StatusInstalled
	switch existing, err := os.ReadFile(path); {
	case err == nil && string(existing) == want:
		return StatusUnchanged, nil
	case err == nil:
		status = StatusUpdated
	case !os.IsNotExist(err):
		return StatusFailed, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return StatusFailed, err
	}
	if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
		return StatusFailed, err
	}
	return status, nil
}

func (t Target) removeLegacy(home string) []string {
	var removed []string
	for _, rel := range t.legacy {
		path := filepath.Join(home, rel)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if err := os.Remove(path); err == nil {
			removed = append(removed, path)
		}
	}
	return removed
}

const fence = "---\n"

// filterFrontmatter drops every top-level frontmatter key not in keep. A nil
// keep list returns the document untouched.
//
// Keys are located by parsing the frontmatter, but the retained lines are
// copied from the original text so that quoting, wrapping, and long values
// survive verbatim -- agents parse this file with varying strictness, and
// re-emitting YAML would rewrite bytes we have no reason to touch.
func filterFrontmatter(doc string, keep []string) (string, error) {
	if keep == nil || !strings.HasPrefix(doc, fence) {
		return doc, nil
	}
	rest := doc[len(fence):]
	end := strings.Index(rest, "\n"+fence)
	if end < 0 {
		return doc, nil
	}
	front, body := rest[:end], rest[end+len("\n"+fence):]

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(front), &root); err != nil {
		return "", fmt.Errorf("parse skill frontmatter: %w", err)
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return doc, nil
	}

	allowed := make(map[string]bool, len(keep))
	for _, k := range keep {
		allowed[k] = true
	}

	// Each key owns the lines from its own up to the next key's, so multi-line
	// values travel with the key they belong to.
	lines := strings.Split(front, "\n")
	pairs := root.Content[0].Content
	var kept []string
	for i := 0; i < len(pairs); i += 2 {
		start := pairs[i].Line - 1
		stop := len(lines)
		if i+2 < len(pairs) {
			stop = pairs[i+2].Line - 1
		}
		if start < 0 || stop > len(lines) || start > stop {
			return "", fmt.Errorf("skill frontmatter: key %q at unexpected line %d", pairs[i].Value, pairs[i].Line)
		}
		if allowed[pairs[i].Value] {
			kept = append(kept, lines[start:stop]...)
		}
	}

	return fence + strings.Join(kept, "\n") + "\n" + fence + body, nil
}
