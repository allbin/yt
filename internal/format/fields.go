package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
)

// ProjectFields formats a list of project custom fields for display.
func ProjectFields(w io.Writer, fields []youtrack.ProjectField) error {
	if len(fields) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No custom fields found."))
		return err
	}

	ew := &errWriter{w: w}
	for i, f := range fields {
		if i > 0 {
			ew.println()
		}
		typeStr := ""
		if f.Type != "" {
			typeStr = " " + StyleDim.Render("("+f.Type+")")
		}
		ew.printf("%s%s\n", StyleBold.Render(f.Name), typeStr)
		if len(f.Values) > 0 {
			names := make([]string, len(f.Values))
			for j, v := range f.Values {
				names[j] = v.Name
			}
			ew.printf("  %s\n", strings.Join(names, ", "))
		}
	}
	return ew.err
}
