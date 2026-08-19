package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// interactive reports whether a full-screen TUI can be drawn for this command.
//
// A TUI needs a terminal to draw on, so commands that offer one fall back to
// their non-interactive output when there isn't one -- an agent, a pipe, or a
// redirect gets data instead of "could not open a new TTY". Passing --json is
// the explicit way to ask for the same fallback.
func interactive(cmd *cobra.Command) bool {
	if jsonOutput {
		return false
	}
	f, ok := cmd.OutOrStdout().(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}
