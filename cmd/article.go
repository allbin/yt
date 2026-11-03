package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var articleCmd = &cobra.Command{
	Use:   "article <id>",
	Short: "Show or manage knowledge base articles",
	Long: `Fetch a knowledge base article by its readable ID (e.g. PROJ-A-12) and display
its summary, project, parent and sub-articles, tags, the issues its content
mentions, its content, and its attachments.

Use subcommands to search, create, update, comment on, and delete articles.
Attachments on an article go through "yt attachment" with the article ID.`,
	Example: `  # show an article
  yt article PROJ-A-12

  # show an article as JSON
  yt article PROJ-A-12 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runArticle,
}

func init() {
	rootCmd.AddCommand(articleCmd)
}

func runArticle(cmd *cobra.Command, args []string) error {
	id, err := articleIDArg(args[0])
	if err != nil {
		return err
	}
	client, err := apiFactory()
	if err != nil {
		return err
	}
	article, err := client.GetArticle(id)
	if err != nil {
		return err
	}
	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, article)
	}
	return format.Article(w, article)
}

// articleIDArg refuses an ID that is not an article's, so an issue ID given
// by mistake errors here instead of as an article the server cannot find.
// The server finds articles by the upper-case ID only.
func articleIDArg(id string) (string, error) {
	if !youtrack.IsArticleID(id) {
		return "", fmt.Errorf("%q is not an article ID (articles look like PROJ-A-12)", id)
	}
	return strings.ToUpper(id), nil
}
