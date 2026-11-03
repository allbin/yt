package cmd

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/allbin/yt/internal/git"
	"github.com/allbin/yt/internal/version"
	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	jsonOutput bool
	apiFactory func() (youtrack.API, error)
)

var rootCmd = &cobra.Command{
	Use:     "yt",
	Short:   "YouTrack CLI",
	Version: version.Version(),
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

// issueClient returns the issue ID from the command args or, without one,
// from the current git branch, with the client to fetch it. An empty ID
// with a nil error means neither names an issue, and the caller shows help.
//
// Any branch that starts with word-number looks like an issue ID
// (session-996d05d2 reads as SESSION-996), so a branch's ID counts only when
// its prefix is a project's short name.
// currentBranch is the git branch lookup, replaced in tests.
var currentBranch = git.CurrentBranch

func issueClient(args []string) (string, youtrack.API, error) {
	id, branch := "", ""
	if len(args) > 0 {
		id = args[0]
	} else if b, err := currentBranch(); err == nil {
		branch = b
		id = git.IssueFromBranch(b)
	}
	if id == "" {
		return "", nil, nil
	}

	client, err := apiFactory()
	if err != nil {
		return "", nil, err
	}
	if branch == "" {
		return id, client, nil
	}
	projects, err := client.ListProjects()
	if err != nil {
		return "", nil, err
	}
	prefix, _, _ := strings.Cut(id, "-")
	if !slices.ContainsFunc(projects, func(p youtrack.Project) bool { return strings.EqualFold(p.ShortName, prefix) }) {
		return "", nil, fmt.Errorf("no issue ID given, and branch %q names none (%s is not a project); pass the issue ID", branch, prefix)
	}
	return id, client, nil
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
