package state

import (
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

type BoardLayout struct {
	Minimized []string `yaml:"minimized,omitempty"`
}

type AppState struct {
	Boards map[string]BoardLayout `yaml:"boards,omitempty"`
	path   string
}

func Load() *AppState {
	p := filePath()
	s := &AppState{
		Boards: map[string]BoardLayout{},
		path:   p,
	}

	data, err := os.ReadFile(p)
	if err != nil {
		return s
	}
	_ = yaml.Unmarshal(data, s)
	if s.Boards == nil {
		s.Boards = map[string]BoardLayout{}
	}
	return s
}

func (s *AppState) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *AppState) BoardMinimized(boardID string) []string {
	if bl, ok := s.Boards[boardID]; ok {
		return bl.Minimized
	}
	return nil
}

func (s *AppState) SetBoardMinimized(boardID string, names []string) {
	bl := s.Boards[boardID]
	bl.Minimized = names
	s.Boards[boardID] = bl
}

func filePath() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, "yt", "state.yaml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "yt", "state.yaml")
}
