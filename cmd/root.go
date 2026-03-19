package cmd

import (
	"fmt"
	"os"

	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var jsonOutput bool

var rootCmd = &cobra.Command{
	Use:   "yt",
	Short: "YouTrack CLI",
	Long: `Command-line interface for JetBrains YouTrack.

Fetch issues, list and filter them, and output as human-readable text or JSON.
Requires YOUTRACK_URL and YOUTRACK_TOKEN environment variables.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Root exposes the root command for doc generators and tooling.
func Root() *cobra.Command { return rootCmd }

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "yt:", err)
		os.Exit(1)
	}
}

func init() {
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

func newClient() (*youtrack.Client, error) {
	u := viper.GetString("URL")
	token := viper.GetString("TOKEN")
	if u == "" {
		return nil, fmt.Errorf("YOUTRACK_URL not set")
	}
	if token == "" {
		return nil, fmt.Errorf("YOUTRACK_TOKEN not set")
	}
	return youtrack.NewClient(u, token), nil
}
