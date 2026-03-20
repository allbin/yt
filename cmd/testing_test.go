package cmd

import (
	"bytes"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

func setupTest(t *testing.T, api youtrack.API) func(args ...string) (string, error) {
	t.Helper()
	orig := apiFactory
	apiFactory = func() (youtrack.API, error) { return api, nil }
	t.Cleanup(func() { apiFactory = orig })

	return func(args ...string) (string, error) {
		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetErr(buf)
		rootCmd.SetArgs(args)
		// Reset persistent flags to defaults before each invocation.
		jsonOutput = false
		t.Cleanup(func() {
			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
			jsonOutput = false
		})
		err := rootCmd.Execute()
		return buf.String(), err
	}
}
