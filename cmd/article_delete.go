package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var articleDeleteYes bool

var articleDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a knowledge base article",
	Long: `Delete a knowledge base article permanently. It cannot be undone, so --yes is
required. An article with sub-articles is refused: delete or move them
first, so a whole tree never goes in one call.`,
	Example: `  yt article delete PROJ-A-12 --yes`,
	Args:    cobra.ExactArgs(1),
	RunE:    runArticleDelete,
}

func init() {
	articleCmd.AddCommand(articleDeleteCmd)
	articleDeleteCmd.Flags().BoolVar(&articleDeleteYes, "yes", false, "confirm the permanent delete")
}

func runArticleDelete(cmd *cobra.Command, args []string) error {
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
	if article.HasChildren || len(article.Children) > 0 {
		return fmt.Errorf("%s %q has %d sub-articles; delete or move them first", id, article.Summary, len(article.Children))
	}
	if !articleDeleteYes {
		return fmt.Errorf("deleting %s %q cannot be undone; rerun with --yes", id, article.Summary)
	}

	if err := client.DeleteArticle(id); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s %s\n", id, article.Summary)
	return err
}
