package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/allbin/yt/internal/youtrack"
	"go.yaml.in/yaml/v3"
)

// setupLogin wires a mock auth client and an isolated HOME, returning a runner.
func setupLogin(t *testing.T, mock *mockAPI) func(args ...string) (string, error) {
	t.Helper()
	run := setupTest(t, mock)

	orig := loginAPIFactory
	loginAPIFactory = func(string, string) youtrack.API { return mock }
	t.Cleanup(func() { loginAPIFactory = orig })

	t.Setenv("HOME", t.TempDir())
	return run
}

func TestLoginSuccess(t *testing.T) {
	mock := &mockAPI{currentUser: &youtrack.User{Login: "jdoe", FullName: "Jane Doe"}}
	run := setupLogin(t, mock)

	out, err := run("login", "--url", "https://yt.example.com/", "--token", "perm:abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Jane Doe") || !strings.Contains(out, "jdoe") {
		t.Errorf("output missing user: %s", out)
	}

	path := filepath.Join(os.Getenv("HOME"), ".config", "yt", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config not written: %v", err)
	}
	var cfg map[string]string
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("invalid yaml: %v", err)
	}
	if cfg["url"] != "https://yt.example.com" {
		t.Errorf("url = %q (trailing slash should be trimmed)", cfg["url"])
	}
	if cfg["token"] != "perm:abc123" {
		t.Errorf("token = %q", cfg["token"])
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestLoginAddsScheme(t *testing.T) {
	mock := &mockAPI{currentUser: &youtrack.User{Login: "jdoe"}}
	run := setupLogin(t, mock)

	if _, err := run("login", "--url", "yt.example.com", "--token", "perm:x"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".config", "yt", "config.yaml"))
	var cfg map[string]string
	_ = yaml.Unmarshal(data, &cfg)
	if cfg["url"] != "https://yt.example.com" {
		t.Errorf("url = %q, want https:// prepended", cfg["url"])
	}
}

func TestLoginValidationFailsNoWrite(t *testing.T) {
	mock := &mockAPI{currentUserErr: &youtrack.APIError{StatusCode: 401, Body: "Unauthorized"}}
	run := setupLogin(t, mock)

	_, err := run("login", "--url", "https://yt.example.com", "--token", "bad")
	if err == nil {
		t.Fatal("expected error for bad token")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("unexpected error: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(os.Getenv("HOME"), ".config", "yt", "config.yaml")); !os.IsNotExist(statErr) {
		t.Errorf("config should not be written on auth failure")
	}
}

func TestLoginPreservesExistingKeys(t *testing.T) {
	mock := &mockAPI{currentUser: &youtrack.User{Login: "jdoe"}}
	run := setupLogin(t, mock)

	dir := filepath.Join(os.Getenv("HOME"), ".config", "yt")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("default_board: Sprint\ntoken: old\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := run("login", "--url", "https://yt.example.com", "--token", "perm:new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	var cfg map[string]string
	_ = yaml.Unmarshal(data, &cfg)
	if cfg["default_board"] != "Sprint" {
		t.Errorf("default_board lost: %v", cfg)
	}
	if cfg["token"] != "perm:new" {
		t.Errorf("token = %q, want perm:new", cfg["token"])
	}
}
