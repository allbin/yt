package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/spf13/cobra"
)

var (
	articleUpdateSummary string
	articleUpdateContent string
	articleUpdateAppend  string
)

var articleUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Change an article's title or content",
	Long: `Change a knowledge base article's title, replace its content, or append to it.
Title and content are saved in one request.

--content replaces the whole content. --append adds text after the current
content, separated by a blank line: it reads the article and writes it back,
so an edit someone saves between the two is lost. For notes several people
write at once, give each their own sub-article.

--content and --append accept "@path" to read a file or "-" to read stdin.`,
	Example: `  # rename
  yt article update PROJ-A-12 -s "Planning 2026-11-03"

  # replace the content from a file
  yt article update PROJ-A-12 -c @notes.md

  # append a section
  yt article update PROJ-A-12 --append "## Decisions
- ship PROJ-123 this sprint"`,
	Args: cobra.ExactArgs(1),
	RunE: runArticleUpdate,
}

func init() {
	articleCmd.AddCommand(articleUpdateCmd)
	articleUpdateCmd.Flags().StringVarP(&articleUpdateSummary, "summary", "s", "", "new title")
	articleUpdateCmd.Flags().StringVarP(&articleUpdateContent, "content", "c", "", "replace content (@file or - for stdin)")
	articleUpdateCmd.Flags().StringVar(&articleUpdateAppend, "append", "", "append to content (@file or - for stdin)")
	articleUpdateCmd.MarkFlagsMutuallyExclusive("content", "append")
	articleUpdateCmd.MarkFlagsOneRequired("summary", "content", "append")
}

func runArticleUpdate(cmd *cobra.Command, args []string) error {
	id, err := articleIDArg(args[0])
	if err != nil {
		return err
	}

	fields := map[string]string{}
	if cmd.Flags().Changed("summary") {
		if strings.TrimSpace(articleUpdateSummary) == "" {
			return fmt.Errorf("--summary cannot be empty")
		}
		fields["summary"] = articleUpdateSummary
	}
	if cmd.Flags().Changed("content") {
		content, err := readTextArg(articleUpdateContent, cmd.InOrStdin())
		if err != nil {
			return err
		}
		fields["content"] = content
	}

	client, err := apiFactory()
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("append") {
		addition, err := readTextArg(articleUpdateAppend, cmd.InOrStdin())
		if err != nil {
			return err
		}
		current, err := client.GetArticle(id)
		if err != nil {
			return err
		}
		fields["content"] = appendContent(current.Content, addition)
	}

	if err := client.UpdateArticle(id, fields); err != nil {
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
	_, err = fmt.Fprintf(w, "Updated %s %s\n", article.IDReadable, article.Summary)
	return err
}

// appendContent joins addition to content with one blank line between them.
func appendContent(content, addition string) string {
	content = strings.TrimRight(content, "\n")
	addition = strings.Trim(addition, "\n")
	if content == "" {
		return addition
	}
	return content + "\n\n" + addition
}
