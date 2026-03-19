package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/git"
	"github.com/spf13/cobra"
)

var branchNoSlug bool

var branchCmd = &cobra.Command{
	Use:   "branch <id>",
	Short: "Create git branch from issue",
	Long: `Create and switch to a new git branch named after a YouTrack issue.

Branch name format: <id>-<slugified-summary> (lowercase).
Use --no-slug for just the issue ID.`,
	Example: `  # creates branch like "proj-123-fix-login-bug"
  yt branch PROJ-123

  # creates branch "proj-123"
  yt branch PROJ-123 --no-slug`,
	Args: cobra.ExactArgs(1),
	RunE: runBranch,
}

func init() {
	rootCmd.AddCommand(branchCmd)
	branchCmd.Flags().BoolVar(&branchNoSlug, "no-slug", false, "omit summary slug from branch name")
}

func runBranch(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	issue, err := client.GetIssue(args[0])
	if err != nil {
		return err
	}

	name := strings.ToLower(issue.IDReadable)
	if !branchNoSlug {
		slug := git.Slugify(issue.Summary)
		if slug != "" {
			name += "-" + slug
		}
	}

	if err := git.Checkout(name); err != nil {
		return fmt.Errorf("git checkout -b %s: %w", name, err)
	}

	fmt.Printf("switched to new branch %s\n", name)
	return nil
}
