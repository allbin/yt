package cmd

import (
	"bytes"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/pflag"
)

func setupTest(t *testing.T, api youtrack.API) func(args ...string) (string, error) {
	t.Helper()
	orig := apiFactory
	apiFactory = func() (youtrack.API, error) { return api, nil }
	t.Cleanup(func() { apiFactory = orig })

	resetFlags := func() {
		jsonOutput = false
		createProject = ""
		createSummary = ""
		createDescription = ""
		createSubsystem = ""
		createTags = nil
		createFields = nil
		updateState = ""
		updateAssignee = ""
		updatePriority = ""
		updateType = ""
		updateSubsystem = ""
		updateTags = nil
		updateRemoveTags = nil
		updateFields = nil
		// Reset cobra's Changed state on all flags.
		rootCmd.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
		rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
		for _, c := range rootCmd.Commands() {
			c.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
			for _, sc := range c.Commands() {
				sc.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
			}
		}
	}

	return func(args ...string) (string, error) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs(args)
		resetFlags()
		t.Cleanup(func() {
			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
			resetFlags()
		})
		err := rootCmd.Execute()
		return buf.String(), err
	}
}
