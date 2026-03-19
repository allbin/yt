package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var completionShell string

var installCompletionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Install shell completions",
	Long: `Install shell completion script for tab-completion of yt commands and flags.

Auto-detects shell from $SHELL. Supported: bash, zsh, fish.
Use --shell to override detection.`,
	Example: `  yt install completion
  yt install completion --shell fish`,
	RunE: runInstallCompletion,
}

func init() {
	installCmd.AddCommand(installCompletionCmd)
	installCompletionCmd.Flags().StringVar(&completionShell, "shell", "", "override shell detection (bash|zsh|fish)")
}

func runInstallCompletion(cmd *cobra.Command, args []string) error {
	shell := completionShell
	if shell == "" {
		shell = detectShell()
	}
	if shell == "" {
		return fmt.Errorf("could not detect shell, use --shell")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var path string
	var gen func(*os.File) error

	switch shell {
	case "fish":
		path = filepath.Join(home, ".config", "fish", "completions", "yt.fish")
		gen = func(f *os.File) error { return rootCmd.GenFishCompletion(f, true) }
	case "bash":
		dataDir := os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = filepath.Join(home, ".local", "share")
		}
		path = filepath.Join(dataDir, "bash-completion", "completions", "yt")
		gen = func(f *os.File) error { return rootCmd.GenBashCompletionV2(f, true) }
	case "zsh":
		path = filepath.Join(home, ".local", "share", "zsh", "site-functions", "_yt")
		gen = func(f *os.File) error { return rootCmd.GenZshCompletion(f) }
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	if err := gen(f); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	fmt.Printf("installed %s completions to %s\n", shell, path)
	return nil
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	base := filepath.Base(shell)
	switch base {
	case "bash", "zsh", "fish":
		return base
	}
	return ""
}
