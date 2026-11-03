package cmd

import (
	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var (
	articleListProject string
	articleListQuery   string
	articleListParent  string
	articleListLimit   int
)

var articleListCmd = &cobra.Command{
	Use:   "list",
	Short: "Search knowledge base articles",
	Long: `List knowledge base articles, without their content. --project and --query
combine into one YouTrack search; the query takes free text and attributes
such as "mentions: PROJ-123". --parent lists the direct sub-articles of an
article instead.

Text output is a tree: each article under its parent when the parent is in
the result, in the order the knowledge base shows them. --json gives each
article's id, summary, project, parent, hasChildren, and timestamps; read one
with "yt article <id>" for its content.`,
	Example: `  # every article in a project
  yt article list -p PROJ

  # full-text search
  yt article list -q "retro"

  # articles mentioning an issue
  yt article list -q "mentions: PROJ-123"

  # sub-articles of a page, e.g. a meeting notes folder
  yt article list --parent PROJ-A-3 --json`,
	Args: cobra.NoArgs,
	RunE: runArticleList,
}

func init() {
	articleCmd.AddCommand(articleListCmd)
	articleListCmd.Flags().StringVarP(&articleListProject, "project", "p", "", "filter by project")
	articleListCmd.Flags().StringVarP(&articleListQuery, "query", "q", "", "search text or YouTrack query")
	articleListCmd.Flags().StringVar(&articleListParent, "parent", "", "list the sub-articles of this article")
	articleListCmd.Flags().IntVarP(&articleListLimit, "limit", "n", 100, "max results")
	articleListCmd.MarkFlagsMutuallyExclusive("parent", "project")
	articleListCmd.MarkFlagsMutuallyExclusive("parent", "query")
}

func runArticleList(cmd *cobra.Command, _ []string) error {
	client, err := apiFactory()
	if err != nil {
		return err
	}

	var articles []youtrack.Article
	if articleListParent != "" {
		parent, err := articleIDArg(articleListParent)
		if err != nil {
			return err
		}
		articles, err = client.ListChildArticles(parent)
		if err != nil {
			return err
		}
	} else {
		query := youtrack.BuildQuery(articleListProject, "", "", articleListQuery)
		articles, err = client.ListArticles(query, articleListLimit)
		if err != nil {
			return err
		}
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, articles)
	}
	return format.ArticleList(w, articles)
}
