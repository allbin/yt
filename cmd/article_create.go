package cmd

import (
	"fmt"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var (
	articleCreateProject string
	articleCreateSummary string
	articleCreateContent string
	articleCreateParent  string
)

var articleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a knowledge base article",
	Long: `Create a knowledge base article in a project, optionally as a sub-article of
another. With --parent the project defaults to the parent's.

Content is markdown. Issue IDs written in it (PROJ-123) become mentions: the
issue shows the article under "Articles" in "yt issue", and the article lists
the issue under "Mentions".

--content accepts "@path" to read a file or "-" to read stdin.`,
	Example: `  # a top-level article
  yt article create -p PROJ -s "Onboarding"

  # meeting notes under a notes page, content from a file
  yt article create --parent PROJ-A-3 -s "Planning 2026-11-03" -c @notes.md

  # content from stdin
  cat notes.md | yt article create -p PROJ -s "Retro" -c - --json`,
	Args: cobra.NoArgs,
	RunE: runArticleCreate,
}

func init() {
	articleCmd.AddCommand(articleCreateCmd)
	articleCreateCmd.Flags().StringVarP(&articleCreateProject, "project", "p", "", "project short name (default: the parent's)")
	articleCreateCmd.Flags().StringVarP(&articleCreateSummary, "summary", "s", "", "article title (required)")
	articleCreateCmd.Flags().StringVarP(&articleCreateContent, "content", "c", "", "markdown content (@file or - for stdin)")
	articleCreateCmd.Flags().StringVar(&articleCreateParent, "parent", "", "create as a sub-article of this article")
	_ = articleCreateCmd.MarkFlagRequired("summary")
}

func runArticleCreate(cmd *cobra.Command, _ []string) error {
	content, err := readTextArg(articleCreateContent, cmd.InOrStdin())
	if err != nil {
		return err
	}

	client, err := apiFactory()
	if err != nil {
		return err
	}

	n := youtrack.NewArticle{Project: articleCreateProject, Summary: articleCreateSummary, Content: content}
	if articleCreateParent != "" {
		parentID, err := articleIDArg(articleCreateParent)
		if err != nil {
			return err
		}
		parent, err := client.GetArticle(parentID)
		if err != nil {
			return err
		}
		n.ParentID = parent.ID
		if n.Project == "" && parent.Project != nil {
			n.Project = parent.Project.ShortName
		}
	}
	if n.Project == "" {
		return fmt.Errorf("--project or --parent is required")
	}

	article, err := client.CreateArticle(n)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, article)
	}
	_, err = fmt.Fprintf(w, "Created %s %s\n", article.IDReadable, article.Summary)
	return err
}
