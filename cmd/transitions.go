package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/allbin/yt/internal/format"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
)

// allUsers is the --user value that turns the author filter off.
const allUsers = "all"

var transitionsCmd = &cobra.Command{
	Use:   "transitions",
	Short: "List state changes from issue activity history",
	Long: `List changes of an issue field (State by default) as recorded in YouTrack's
activity history: who changed which issue, when, and from which value to which.

This reads the actual history, so it shows what a user did — not who the
issue is assigned to now or when it was last updated. All pages of the
history are fetched.

--user takes "me" (default), a login, or a full name; "all" shows every user.
--since and --until take a date (2026-09-30), an RFC 3339 timestamp, or a
duration back from now (36h, 7d, 2w). A date-only --until includes that whole
day. --since defaults to 7d, --until to now.

--board reads the changes against an agile board: the field becomes the one
the board's columns are bound to (State, Stage, ...), each change gets the
board column it left and entered, and only issues currently on one of the
board's sprints are kept.`,
	Example: `  # my state changes over the last week
  yt transitions

  # same, as JSON for a standup summary
  yt transitions --json

  # a teammate's changes in a date range
  yt transitions -u alice --since 2026-09-28 --until 2026-10-04

  # everyone's moves on a board, with column names
  yt transitions -u all --board AllTix --since 14d

  # changes to another field, limited to one project
  yt transitions --field Priority -p AX`,
	Args: cobra.NoArgs,
	RunE: runTransitions,
}

var (
	transitionsUser    string
	transitionsSince   string
	transitionsUntil   string
	transitionsField   string
	transitionsBoard   string
	transitionsProject string
	transitionsQuery   string
)

func init() {
	rootCmd.AddCommand(transitionsCmd)
	transitionsCmd.Flags().StringVarP(&transitionsUser, "user", "u", "me", `who made the change ("me", login, full name, or "all")`)
	transitionsCmd.Flags().StringVar(&transitionsSince, "since", "7d", "start of range (date, RFC 3339, or duration ago)")
	transitionsCmd.Flags().StringVar(&transitionsUntil, "until", "", "end of range (default: now)")
	transitionsCmd.Flags().StringVar(&transitionsField, "field", "State", "field whose changes to list")
	transitionsCmd.Flags().StringVar(&transitionsBoard, "board", "", "read changes against this board's columns")
	transitionsCmd.Flags().StringVarP(&transitionsProject, "project", "p", "", "only issues in this project")
	transitionsCmd.Flags().StringVarP(&transitionsQuery, "query", "q", "", "only issues matching this YouTrack query")
}

func runTransitions(cmd *cobra.Command, args []string) error {
	now := time.Now()
	since, err := parseTimeFlag(transitionsSince, now, false)
	if err != nil {
		return fmt.Errorf("--since: %w", err)
	}
	until := now
	if transitionsUntil != "" {
		if until, err = parseTimeFlag(transitionsUntil, now, true); err != nil {
			return fmt.Errorf("--until: %w", err)
		}
	}
	if !since.Before(until) {
		return fmt.Errorf("--since (%s) must be before --until (%s)",
			since.Format(time.RFC3339), until.Format(time.RFC3339))
	}

	client, err := apiFactory()
	if err != nil {
		return err
	}

	author := ""
	if !strings.EqualFold(transitionsUser, allUsers) {
		if author, err = resolveAssignee(client, transitionsUser); err != nil {
			return err
		}
	}

	field := transitionsField
	var board *youtrack.Agile
	if transitionsBoard != "" {
		if board, err = client.GetBoardForView(transitionsBoard); err != nil {
			return err
		}
		if field, err = boardField(board, cmd.Flags().Changed("field")); err != nil {
			return err
		}
	}

	acts, err := client.ListFieldActivities(youtrack.FieldActivityFilter{
		Author:     author,
		Start:      since,
		End:        until,
		IssueQuery: youtrack.BuildQuery(transitionsProject, "", "", transitionsQuery),
	})
	if err != nil {
		return err
	}
	transitions := youtrack.Transitions(acts, field)

	if board != nil {
		if transitions, err = onBoard(client, board, transitions); err != nil {
			return err
		}
	}

	w := cmd.OutOrStdout()
	if jsonOutput {
		return format.JSON(w, transitions)
	}
	return format.TransitionList(w, transitions)
}

// boardField returns the field the board's columns are bound to. An explicit
// --field must agree with it, since only that field moves cards between columns.
func boardField(board *youtrack.Agile, fieldSet bool) (string, error) {
	field := board.ColumnField()
	if field == "" {
		return "", fmt.Errorf("board %q has no column field", board.Name)
	}
	if fieldSet && !strings.EqualFold(transitionsField, field) {
		return "", fmt.Errorf("board %q columns are bound to %s, not %s", board.Name, field, transitionsField)
	}
	return field, nil
}

// onBoard keeps the transitions of issues on the board's sprints and fills in
// the columns each one left and entered.
func onBoard(client youtrack.API, board *youtrack.Agile, transitions []youtrack.Transition) ([]youtrack.Transition, error) {
	ids, err := client.BoardIssues(board)
	if err != nil {
		return nil, err
	}
	member := make(map[string]bool, len(ids))
	for _, id := range ids {
		member[strings.ToUpper(id)] = true
	}

	out := []youtrack.Transition{}
	for _, t := range transitions {
		if !member[strings.ToUpper(t.Issue)] {
			continue
		}
		t.FromColumn = board.ColumnOf(t.From)
		t.ToColumn = board.ColumnOf(t.To)
		out = append(out, t)
	}
	return out, nil
}

// parseTimeFlag reads a date (2006-01-02, local midnight), an RFC 3339
// timestamp, or a duration before now (36h, 7d, 2w). With endOfDay, a date
// means the last millisecond of that day so the day is included.
func parseTimeFlag(s string, now time.Time, endOfDay bool) (time.Time, error) {
	if t, err := time.ParseInLocation(time.DateOnly, s, now.Location()); err == nil {
		if endOfDay {
			return t.AddDate(0, 0, 1).Add(-time.Millisecond), nil
		}
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if d, ok := parseAgo(s); ok {
		return now.Add(-d), nil
	}
	return time.Time{}, fmt.Errorf("invalid time %q (want 2006-01-02, RFC 3339, or a duration like 7d)", s)
}

// parseAgo reads a positive duration, extending time.ParseDuration with d
// (days) and w (weeks).
func parseAgo(s string) (time.Duration, bool) {
	if n, ok := strings.CutSuffix(s, "d"); ok {
		days, err := strconv.Atoi(n)
		return time.Duration(days) * 24 * time.Hour, err == nil && days > 0
	}
	if n, ok := strings.CutSuffix(s, "w"); ok {
		weeks, err := strconv.Atoi(n)
		return time.Duration(weeks) * 7 * 24 * time.Hour, err == nil && weeks > 0
	}
	d, err := time.ParseDuration(s)
	return d, err == nil && d > 0
}
