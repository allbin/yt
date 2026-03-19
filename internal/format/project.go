package format

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/allbin/yt/internal/youtrack"
)

func ProjectList(w io.Writer, projects []youtrack.Project) error {
	if len(projects) == 0 {
		_, err := fmt.Fprintln(w, "No projects found.")
		return err
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	ew := &errWriter{w: tw}
	ew.println("SHORT\tNAME")
	for _, p := range projects {
		ew.printf("%s\t%s\n", p.ShortName, p.Name)
	}
	if ew.err != nil {
		return ew.err
	}
	return tw.Flush()
}
