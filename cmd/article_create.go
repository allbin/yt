package cmd

import (
	"fmt"
	"slices"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var (
	articleCreateProject string
	articleCreateSummary string
	articleCreateContent string
	articleCreateParent  string
	articleCreateTags    []string
	articleCreateNewTag  bool
)

var articleCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a knowledge base article",
	Long: `Create a knowledge base article in a project, optionally as a sub-article of
another. With --parent the project defaults to the parent's.

Content is markdown. Issue IDs written in it (PROJ-123) become mentions: the
issue shows the article under "Articles" in "yt issue", and the article lists
the issue under "Mentions".

--tag matches tag names as "yt issue create" does; an unknown name errors
with the closest tags unless --create-tag is given. Existing tags go with the
create request, so a bad value creates nothing.

--content accepts "@path" to read a file or "-" to read stdin.`,
	Example: `  # a top-level article
  yt article create -p PROJ -s "Onboarding"

  # meeting notes under a notes page, content from a file
  yt article create --parent PROJ-A-3 -s "Planning 2026-11-03" -c @notes.md

  # tagged
  yt article create -p PROJ -s "Retro 2026-11-03" --tag meeting --tag retro

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
	articleCreateCmd.Flags().StringArrayVarP(&articleCreateTags, "tag", "t", nil, "tag (repeatable)")
	articleCreateCmd.Flags().BoolVar(&articleCreateNewTag, "create-tag", false, "create --tag names that do not exist yet")
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

	tags, err := lookupAddTags(client, articleCreateTags, articleCreateNewTag)
	if err != nil {
		return err
	}
	n.TagIDs = tagIDs(tags.found)

	article, err := client.CreateArticle(n)
	if err != nil {
		return err
	}

	// The article exists from here on. New tags wait for it, so a rejected
	// create leaves none behind; a failed tag step still prints the article.
	stepErr := finishArticleTags(client, article, tags)

	w := cmd.OutOrStdout()
	if jsonOutput {
		err = format.JSON(w, article)
	} else {
		_, err = fmt.Fprintf(w, "Created %s %s\n", article.IDReadable, article.Summary)
	}
	if err != nil {
		return err
	}
	if stepErr != nil {
		return fmt.Errorf("%s was created, but tagging failed; finish it with `yt article update %s`, do not create it again: %w",
			article.IDReadable, article.IDReadable, stepErr)
	}
	return nil
}

// finishArticleTags creates and applies the --create-tag tags of a new
// article, then reads its tags back. article is refreshed in place.
func finishArticleTags(client youtrack.API, article *youtrack.Article, tags tagLookup) error {
	if len(tags.found) == 0 && len(tags.missing) == 0 {
		return nil
	}
	// The found tags went with the create request; only new ones remain.
	created, err := tagPlan{add: tagLookup{missing: tags.missing}}.apply(client, article.IDReadable)
	if err != nil {
		return err
	}
	fresh, err := client.GetArticle(article.IDReadable)
	if err != nil {
		return err
	}
	*article = *fresh
	return verifyTags(article.IDReadable, article.Tags, append(slices.Clone(tags.found), created...), nil)
}
