package format

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/allbin/yt/internal/youtrack"
)

func CommentList(w io.Writer, comments []youtrack.Comment) error {
	if len(comments) == 0 {
		_, err := fmt.Fprintln(w, StyleDim.Render("No comments."))
		return err
	}

	ew := &errWriter{w: w}
	for i, c := range comments {
		cv := c.View()
		ts := time.Unix(cv.Created/1000, 0).Format("2006-01-02 15:04")
		ew.printf("%s  %s\n", StyleBold.Render(cv.Author), StyleDim.Render(ts))
		ew.println(strings.TrimSpace(c.Text))
		if i < len(comments)-1 {
			ew.println()
		}
	}
	return ew.err
}
