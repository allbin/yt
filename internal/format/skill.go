package format

import (
	"io"

	"github.com/allbin/yt/internal/skill"
)

// SkillInstall renders one line per target: status, agent, and either the
// installed path or the reason nothing was written.
func SkillInstall(w io.Writer, results []skill.Result) error {
	statusWidth, nameWidth := 0, 0
	for _, r := range results {
		if len(r.Status) > statusWidth {
			statusWidth = len(r.Status)
		}
		if len(r.Name) > nameWidth {
			nameWidth = len(r.Name)
		}
	}

	ew := &errWriter{w: w}
	for _, r := range results {
		detail := r.Path
		if !r.Status.Written() {
			detail = r.Reason
		}
		ew.printf("%-*s  %-*s  %s\n", statusWidth, r.Status, nameWidth, r.Name, detail)
		for _, rm := range r.Removed {
			ew.printf("%-*s  %-*s  removed legacy %s\n", statusWidth, "", nameWidth, "", rm)
		}
	}
	return ew.err
}
