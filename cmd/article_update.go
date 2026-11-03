package cmd

import (
	"fmt"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var (
	articleUpdateSummary    string
	articleUpdateContent    string
	articleUpdateAppend     string
	articleUpdateParent     string
	articleUpdateTags       []string
	articleUpdateRemoveTags []string
	articleUpdateCreateTag  bool
)

var articleUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Change an article's title, content, place, or tags",
	Long: `Change a knowledge base article's title, replace or append to its content,
move it under another article, or add and remove tags. Title, content and
parent are saved in one request; tags follow, and are read back after.

--content replaces the whole content. --append adds text after the current
content, separated by a blank line: it reads the article and writes it back,
so an edit someone saves between the two is lost. For notes several people
write at once, give each their own sub-article.

--parent moves the article, with its sub-articles, under another article;
--parent none moves it to the top level of its project.

--tag and --remove-tag match tag names as "yt issue update" does: an unknown
name errors with the closest tags, and --create-tag creates it instead.

--content and --append accept "@path" to read a file or "-" to read stdin.`,
	Example: `  # rename
  yt article update PROJ-A-12 -s "Planning 2026-11-03"

  # replace the content from a file
  yt article update PROJ-A-12 -c @notes.md

  # append a section
  yt article update PROJ-A-12 --append "## Decisions
- ship PROJ-123 this sprint"

  # move under a notes page, and tag it
  yt article update PROJ-A-12 --parent PROJ-A-3 --tag meeting

  # move to the top level
  yt article update PROJ-A-12 --parent none`,
	Args: cobra.ExactArgs(1),
	RunE: runArticleUpdate,
}

func init() {
	articleCmd.AddCommand(articleUpdateCmd)
	f := articleUpdateCmd.Flags()
	f.StringVarP(&articleUpdateSummary, "summary", "s", "", "new title")
	f.StringVarP(&articleUpdateContent, "content", "c", "", "replace content (@file or - for stdin)")
	f.StringVar(&articleUpdateAppend, "append", "", "append to content (@file or - for stdin)")
	f.StringVar(&articleUpdateParent, "parent", "", `move under this article ("none": to the top level)`)
	f.StringArrayVar(&articleUpdateTags, "tag", nil, "add tag (repeatable)")
	f.StringArrayVar(&articleUpdateRemoveTags, "remove-tag", nil, "remove tag (repeatable)")
	f.BoolVar(&articleUpdateCreateTag, "create-tag", false, "create --tag names that do not exist yet")
	articleUpdateCmd.MarkFlagsMutuallyExclusive("content", "append")
	articleUpdateCmd.MarkFlagsOneRequired("summary", "content", "append", "parent", "tag", "remove-tag")
}

func runArticleUpdate(cmd *cobra.Command, args []string) error {
	id, err := articleIDArg(args[0])
	if err != nil {
		return err
	}

	var u youtrack.ArticleUpdate
	if cmd.Flags().Changed("summary") {
		if strings.TrimSpace(articleUpdateSummary) == "" {
			return fmt.Errorf("--summary cannot be empty")
		}
		u.Summary = &articleUpdateSummary
	}
	if cmd.Flags().Changed("content") {
		content, err := readTextArg(articleUpdateContent, cmd.InOrStdin())
		if err != nil {
			return err
		}
		u.Content = &content
	}

	client, err := apiFactory()
	if err != nil {
		return err
	}

	// Everything is resolved before the first write, so a bad parent or tag
	// changes nothing.
	if cmd.Flags().Changed("append") {
		addition, err := readTextArg(articleUpdateAppend, cmd.InOrStdin())
		if err != nil {
			return err
		}
		current, err := client.GetArticle(id)
		if err != nil {
			return err
		}
		content := appendContent(current.Content, addition)
		u.Content = &content
	}
	if cmd.Flags().Changed("parent") {
		if u.ParentID, err = moveTarget(client, id, articleUpdateParent); err != nil {
			return err
		}
		u.Move = true
	}
	tags, err := planTags(client, articleUpdateTags, articleUpdateCreateTag, articleUpdateRemoveTags)
	if err != nil {
		return err
	}

	if u.Summary != nil || u.Content != nil || u.Move {
		if err := client.UpdateArticle(id, u); err != nil {
			return err
		}
	}
	added, err := tags.apply(client, id)
	if err != nil {
		return err
	}

	article, err := client.GetArticle(id)
	if err != nil {
		return err
	}
	if err := verifyTags(id, article.Tags, added, tags.remove); err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, article)
	}
	_, err = fmt.Fprintf(w, "Updated %s %s\n", article.IDReadable, article.Summary)
	return err
}

// moveTarget resolves a --parent value to the database id of the new parent,
// or "" for the top level.
func moveTarget(client youtrack.API, id, parent string) (string, error) {
	if strings.EqualFold(parent, "none") {
		return "", nil
	}
	parentID, err := articleIDArg(parent)
	if err != nil {
		return "", err
	}
	if parentID == id {
		return "", fmt.Errorf("%s cannot be its own parent", id)
	}
	p, err := client.GetArticle(parentID)
	if err != nil {
		return "", err
	}
	return p.ID, nil
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
