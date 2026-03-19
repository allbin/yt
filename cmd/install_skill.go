package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillContent string

var installSkillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Install Claude Code skill",
	Long: `Install the YouTrack CLI skill for Claude Code.

Writes the skill definition to ~/.claude/skills/yt/SKILL.md so that Claude
can use the yt CLI to fetch and list YouTrack issues.

If a legacy command exists at ~/.claude/commands/yt.md it is removed.`,
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

	skillDir := filepath.Join(home, ".claude", "skills", "yt")
	skillPath := filepath.Join(skillDir, "SKILL.md")

	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(skillPath, []byte(skillContent), 0o644); err != nil {
		return err
	}
	fmt.Println("installed skill to", skillPath)

	legacyPath := filepath.Join(home, ".claude", "commands", "yt.md")
	if _, err := os.Stat(legacyPath); err == nil {
		if err := os.Remove(legacyPath); err == nil {
			fmt.Println("removed legacy command", legacyPath)
		}
	}

	return nil
}
