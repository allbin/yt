package cmd

import (
	"slices"
	"strings"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

var (
	historyFull  bool
	historyKinds []string
)

const historyKindHelp = "only these kinds of change (repeatable or comma-separated)"

var issueHistoryCmd = &cobra.Command{
	Use:   "history [id]",
	Short: "Show every change to an issue",
	Long: `List every recorded change to an issue, oldest first: creation, summary and
description edits, custom fields, links, tags, attachments, comments and their
edits, sprints, project moves, visibility, resolution, and commits.

Each change has a kind: created, summary, description, field, link, tag,
attachment, attachment-rename, comment, comment-edit, comment-attachment,
resolved, sprint, project, visibility, commit, pull-request. --kind keeps only
the listed kinds.

Text output cuts long text to its first line; --full prints it whole, with a
description's text before and after. --json always carries the full text.

If no ID is given, attempts to detect it from the current git branch name.
For State changes across many issues, use "yt transitions".`,
	Example: `  # everything that happened to an issue
  yt issue history PROJ-123

  # only field changes and comments
  yt issue history PROJ-123 --kind field,comment

  # full text of description edits
  yt issue history PROJ-123 --kind description --full`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, client, err := issueClient(args)
		if err != nil {
			return err
		}
		if id == "" {
			return cmd.Help()
		}
		return runHistory(cmd, client, id)
	},
}

var articleHistoryCmd = &cobra.Command{
	Use:   "history <id>",
	Short: "Show every change to an article",
	Long: `List every recorded change to a knowledge base article, oldest first: creation,
title and content edits, comments, tags, and attachments. Moving an article
is not recorded by YouTrack.

Each change has a kind: created, summary, content, comment, tag, attachment,
comment-attachment. --kind keeps only the listed kinds.

Text output cuts long text to its first line; --full prints it whole, with
the content before and after each edit. --json always carries the full text.`,
	Example: `  yt article history PROJ-A-12

  # who changed the content, and how
  yt article history PROJ-A-12 --kind content --full`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := articleIDArg(args[0])
		if err != nil {
			return err
		}
		client, err := apiFactory()
		if err != nil {
			return err
		}
		return runHistory(cmd, client, id)
	},
}

func init() {
	for _, c := range []*cobra.Command{issueHistoryCmd, articleHistoryCmd} {
		c.Flags().BoolVar(&historyFull, "full", false, "print long text whole")
		c.Flags().StringSliceVar(&historyKinds, "kind", nil, historyKindHelp)
	}
	issueCmd.AddCommand(issueHistoryCmd)
	articleCmd.AddCommand(articleHistoryCmd)
}

func runHistory(cmd *cobra.Command, client youtrack.API, id string) error {
	entries, err := client.History(id)
	if err != nil {
		return err
	}
	if len(historyKinds) > 0 {
		kinds := make([]string, len(historyKinds))
		for i, k := range historyKinds {
			kinds[i] = strings.ToLower(strings.TrimSpace(k))
		}
		entries = slices.DeleteFunc(entries, func(e youtrack.HistoryEntry) bool { return !slices.Contains(kinds, e.Kind) })
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		if entries == nil {
			entries = []youtrack.HistoryEntry{}
		}
		return format.JSON(w, entries)
	}
	return format.History(w, entries, historyFull)
}
