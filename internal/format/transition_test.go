package format

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/allbin/yt/internal/youtrack"
)

func TestTransitionList(t *testing.T) {
	transitions := []youtrack.Transition{
		{
			Issue:     "AX-1",
			Summary:   "First card",
			Author:    &youtrack.User{Login: "alice", FullName: "Alice"},
			Timestamp: 1748779200000,
			Field:     "State",
			From:      "Planned",
			To:        "In Progress",
		},
		{
			Issue:     "AX-2",
			Summary:   "Second card",
			Author:    &youtrack.User{Login: "bob"},
			Timestamp: 1748779260000,
			Field:     "State",
			To:        "Submitted",
		},
	}

	var buf bytes.Buffer
	if err := TransitionList(&buf, transitions); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	// Times render in the local zone, so derive the expected text the same way.
	when := time.UnixMilli(1748779200000).Format("2006-01-02 15:04")
	for _, want := range []string{"AX-1", "First card", "Planned", "In Progress", "Alice", "AX-2", "bob", "—", when} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestTransitionListEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := TransitionList(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No changes found") {
		t.Errorf("output = %q, want empty message", buf.String())
	}
}

func TestTransitionListNilAuthor(t *testing.T) {
	var buf bytes.Buffer
	if err := TransitionList(&buf, []youtrack.Transition{{Issue: "AX-1", To: "Done"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Unknown") {
		t.Errorf("output = %q, want Unknown author", buf.String())
	}
}
