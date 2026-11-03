package cmd

import (
	"bytes"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func setupTest(t *testing.T, api youtrack.API) func(args ...string) (string, error) {
	t.Helper()
	orig := apiFactory
	apiFactory = func() (youtrack.API, error) { return api, nil }
	t.Cleanup(func() { apiFactory = orig })

	resetFlags := func() {
		jsonOutput = false
		listCompact = false
		attachmentUploadName = ""
		attachmentOutput = ""
		createProject = ""
		createSummary = ""
		createDescription = ""
		createSubsystem = ""
		createType = ""
		createPriority = ""
		createAssignee = ""
		createLinks = nil
		createTags = nil
		createCreateTag = false
		createFields = nil
		createBoard = ""
		createSprint = ""
		createLike = ""
		createParent = ""
		updateState = ""
		updateAssignee = ""
		updatePriority = ""
		updateType = ""
		updateSubsystem = ""
		updateTags = nil
		updateRemoveTags = nil
		updateCreateTag = false
		updateFields = nil
		updateAddFields = nil
		updateRemoveFields = nil
		updateBoard = ""
		updateSprint = ""
		boardSprint = ""
		boardViewSprint = ""
		boardAddSprint = ""
		boardRemoveSprint = ""
		transitionsUser = "me"
		transitionsSince = "7d"
		transitionsUntil = ""
		transitionsField = "State"
		transitionsBoard = ""
		transitionsProject = ""
		transitionsQuery = ""
		loginURL = ""
		loginToken = ""
		articleListProject = ""
		articleListQuery = ""
		articleListParent = ""
		articleListLimit = 100
		articleCreateProject = ""
		articleCreateSummary = ""
		articleCreateContent = ""
		articleCreateParent = ""
		articleUpdateSummary = ""
		articleUpdateContent = ""
		articleUpdateAppend = ""
		articleUpdateParent = ""
		articleUpdateTags = nil
		articleUpdateRemoveTags = nil
		articleUpdateCreateTag = false
		articleCreateTags = nil
		articleCreateNewTag = false
		commentMessage = ""
		commentEditMessage = ""
		commentDeleteYes = false
		historyFull = false
		historyKinds = nil
		articleDeleteYes = false
		// Reset cobra's Changed state on all flags.
		rootCmd.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
		rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
		var reset func(c *cobra.Command)
		reset = func(c *cobra.Command) {
			c.Flags().VisitAll(func(f *pflag.Flag) { f.Changed = false })
			for _, sc := range c.Commands() {
				reset(sc)
			}
		}
		reset(rootCmd)
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
