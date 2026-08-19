package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/skill"
	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillContent string

var installSkillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Install the yt skill for Claude Code and Codex",
	Long: `Install the YouTrack CLI skill for every supported agent found on this machine.

Supported agents:

  Claude Code   ~/.claude/skills/yt/SKILL.md
  Codex         ~/.codex/skills/yt/SKILL.md

An agent counts as present when its executable is on $PATH or its config
directory holds more than the yt skill itself. Agents that are absent are
skipped, and every agent is reported as installed (nothing was there),
updated (the file existed with older content), unchanged (already current),
or skipped.

The skill frontmatter is adapted per agent: Codex only reads name and
description, so Claude-specific keys are stripped from its copy.

A legacy Claude Code command at ~/.claude/commands/yt.md is removed if present.`,
	Example: `  yt install skill`,
	RunE:    runInstallSkill,
}

func init() {
	installCmd.AddCommand(installSkillCmd)
}

func runInstallSkill(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	results := skill.Install(skillContent, home)
	if err := format.SkillInstall(cmd.OutOrStdout(), results); err != nil {
		return err
	}

	var failed []string
	var written int
	for _, r := range results {
		switch {
		case r.Status == skill.StatusFailed:
			failed = append(failed, r.Name)
		case r.Status.Written():
			written++
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("could not install for %s", strings.Join(failed, ", "))
	}
	if written == 0 {
		return fmt.Errorf("no supported agent found on this machine")
	}
	return nil
}
