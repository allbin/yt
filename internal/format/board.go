package format

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/allbin/yt/internal/youtrack"
)

func BoardList(w io.Writer, boards []youtrack.Agile) error {
	if len(boards) == 0 {
		_, err := fmt.Fprintln(w, "No boards found.")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	ew := &errWriter{w: tw}
	ew.println("NAME\tPROJECTS\tCURRENT SPRINT")
	for _, b := range boards {
		sprint := ""
		if b.CurrentSprint != nil {
			sprint = b.CurrentSprint.Name
		}
		ew.printf("%s\t%s\t%s\n", b.Name, formatProjects(b.Projects), sprint)
	}
	if ew.err != nil {
		return ew.err
	}
	return tw.Flush()
}

func SprintIssues(w io.Writer, board, sprint string, issues []youtrack.Issue) error {
	ew := &errWriter{w: w}
	ew.printf("# %s — %s (%d issues)\n\n", board, sprint, len(issues))

	if len(issues) == 0 {
		ew.println("No issues.")
		return ew.err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	tew := &errWriter{w: tw}
	tew.println("ID\tSTATE\tPRIORITY\tASSIGNEE\tSUMMARY")
	for _, issue := range issues {
		tew.printf("%s\t%s\t%s\t%s\t%s\n",
			issue.IDReadable,
			issue.Field("State"),
			issue.Field("Priority"),
			issue.Field("Assignee"),
			issue.Summary,
		)
	}
	if tew.err != nil {
		return tew.err
	}
	return tw.Flush()
}

func formatProjects(projects []youtrack.Project) string {
	const max = 3
	names := make([]string, 0, max+1)
	for i, p := range projects {
		if i >= max {
			names = append(names, fmt.Sprintf("+%d", len(projects)-max))
			break
		}
		names = append(names, p.ShortName)
	}
	return strings.Join(names, ", ")
}
