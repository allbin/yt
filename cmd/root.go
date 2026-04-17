package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/git"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	jsonOutput bool
	apiFactory func() (youtrack.API, error)
)

var rootCmd = &cobra.Command{
	Use:   "yt",
	Short: "YouTrack CLI",
	Long: `Command-line interface for JetBrains YouTrack.

Fetch issues, list and filter them, and output as human-readable text or JSON.

Configuration is read from environment variables or ~/.config/yt/config.yaml:

  Environment variables:
    YOUTRACK_URL     Base URL of the YouTrack instance
    YOUTRACK_TOKEN   Permanent token for authentication

  Config file (~/.config/yt/config.yaml):
    url: https://youtrack.example.com
    token: perm:abc123...

Environment variables take precedence over the config file.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Root exposes the root command for doc generators and tooling.
func Root() *cobra.Command { return rootCmd }

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, formatError(err))
		os.Exit(1)
	}
}

var errLabel = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("error")

func formatError(err error) string {
	var apiErr *youtrack.APIError
	if errors.As(err, &apiErr) {
		full := err.Error()
		context := strings.TrimSuffix(full, ": "+apiErr.Error())
		msg := apiErr.Message()
		if context != full && context != "" {
			return fmt.Sprintf("%s %s — %s", errLabel, context, msg)
		}
		return fmt.Sprintf("%s %s", errLabel, msg)
	}
	return fmt.Sprintf("%s %s", errLabel, err)
}

func init() {
	apiFactory = newClient
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output raw JSON")
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/yt")
	viper.SetEnvPrefix("YOUTRACK")
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}

func newClient() (youtrack.API, error) {
	u := viper.GetString("URL")
	token := viper.GetString("TOKEN")
	if u == "" {
		return nil, fmt.Errorf("url not configured (set YOUTRACK_URL or url in ~/.config/yt/config.yaml)")
	}
	if token == "" {
		return nil, fmt.Errorf("token not configured (set YOUTRACK_TOKEN or token in ~/.config/yt/config.yaml)")
	}
	return youtrack.NewClient(u, token), nil
}

// issueIDFromArgs resolves an issue ID from command args or the current git branch.
// Returns "" if no ID could be determined.
func issueIDFromArgs(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	branch, err := git.CurrentBranch()
	if err != nil {
		return ""
	}
	return git.IssueFromBranch(branch)
}

// resolveAssignee converts a user-provided assignee value (full name, partial,
// or login) to a YouTrack login for use in queries. Special values like "me"
// pass through unchanged.
func resolveAssignee(client youtrack.API, assignee string) (string, error) {
	if assignee == "" {
		return "", nil
	}
	return client.ResolveUser(assignee)
}
