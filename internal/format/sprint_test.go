package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func TestSprintDates(t *testing.T) {
	start := int64(1748779200000)
	finish := int64(1749988800000)

	if got := sprintDates(youtrack.Sprint{}); got != "" {
		t.Errorf("no dates: got %q, want empty", got)
	}

	both := sprintDates(youtrack.Sprint{Start: &start, Finish: &finish})
	if !strings.Contains(both, " → ") {
		t.Errorf("both dates: %q missing range separator", both)
	}
	if want := millisDate(start) + " → " + millisDate(finish); both != want {
		t.Errorf("both dates: got %q, want %q", both, want)
	}

	if got := sprintDates(youtrack.Sprint{Start: &start}); got != millisDate(start)+" →" {
		t.Errorf("start only: got %q", got)
	}
	if got := sprintDates(youtrack.Sprint{Finish: &finish}); got != "→ "+millisDate(finish) {
		t.Errorf("finish only: got %q", got)
	}
}

func TestSprintListEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := SprintList(&buf, &youtrack.Agile{Name: "Empty"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No sprints") {
		t.Errorf("expected empty message, got %q", buf.String())
	}
}
