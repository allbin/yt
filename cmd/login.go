package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
	"golang.org/x/term"
)

var (
	loginURL   string
	loginToken string
)

// loginAPIFactory builds a client from explicit credentials, bypassing config.
// Overridable in tests.
var loginAPIFactory = func(url, token string) youtrack.API {
	return youtrack.NewClient(url, token)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and save YouTrack credentials",
	Long: `Authenticate against a YouTrack instance and save the credentials to
~/.config/yt/config.yaml.

Prompts for the base URL and a permanent token unless --url and --token are
given. The token is validated by fetching the current user before anything is
written, so a bad token fails fast and never reaches the config file.

Create a permanent token in YouTrack under:
  Profile -> Account Security -> New token...

The token input is hidden when reading from a terminal. Environment variables
(YOUTRACK_URL, YOUTRACK_TOKEN) still take precedence over the saved config at
runtime.`,
	Example: `  # interactive: prompts for URL and token
  yt login

  # non-interactive
  yt login --url https://youtrack.example.com --token perm:abc123...`,
	Args: cobra.NoArgs,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&loginURL, "url", "", "YouTrack base URL (prompted if omitted)")
	loginCmd.Flags().StringVar(&loginToken, "token", "", "permanent token (prompted if omitted)")
}

func runLogin(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()

	rawURL := loginURL
	if rawURL == "" {
		var err error
		rawURL, err = promptLine(out, "YouTrack URL", viper.GetString("URL"))
		if err != nil {
			return err
		}
	}
	u := normalizeURL(rawURL)
	if u == "" {
		return fmt.Errorf("url is required")
	}

	token := loginToken
	if token == "" {
		var err error
		token, err = promptToken(out)
		if err != nil {
			return err
		}
	}
	if token == "" {
		return fmt.Errorf("token is required")
	}

	user, err := loginAPIFactory(u, token).CurrentUser()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	path, err := writeConfig(u, token)
	if err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	name := user.FullName
	if name == "" {
		name = user.Login
	}
	_, err = fmt.Fprintf(out, "Logged in as %s (%s)\nSaved %s\n", name, user.Login, path)
	return err
}

// normalizeURL trims whitespace and trailing slashes and prepends https:// when
// no scheme is present.
func normalizeURL(raw string) string {
	s := strings.TrimRight(strings.TrimSpace(raw), "/")
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}
	return s
}

// promptLine reads a single line from stdin, showing def as the default.
func promptLine(out io.Writer, label, def string) (string, error) {
	prompt := label + ": "
	if def != "" {
		prompt = fmt.Sprintf("%s [%s]: ", label, def)
	}
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return "", err
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

// promptToken reads a permanent token, hiding input when stdin is a terminal.
func promptToken(out io.Writer) (string, error) {
	if _, err := fmt.Fprint(out, "Permanent token: "); err != nil {
		return "", err
	}
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		if _, perr := fmt.Fprintln(out); perr != nil && err == nil {
			err = perr
		}
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// writeConfig writes url and token to ~/.config/yt/config.yaml (0600),
// preserving any other keys already present.
func writeConfig(url, token string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "yt")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "config.yaml")

	cfg := map[string]any{}
	if existing, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(existing, &cfg)
	}
	cfg["url"] = url
	cfg["token"] = token

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}
