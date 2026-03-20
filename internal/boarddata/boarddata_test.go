package boarddata

import (
	"testing"

	"github.com/allbin/yt/internal/youtrack"
)

type mockAPI struct {
	board       *youtrack.Agile
	issues      []youtrack.Issue
	sprintBoard *youtrack.SprintBoard
	boardErr    error
	issuesErr   error
	sbErr       error
	stateSet    string
}

func (m *mockAPI) GetIssue(string) (*youtrack.Issue, error)              { return nil, nil }
func (m *mockAPI) ListIssues(string, int) ([]youtrack.Issue, error)      { return m.issues, m.issuesErr }
func (m *mockAPI) ListBoards() ([]youtrack.Agile, error)                 { return nil, nil }
func (m *mockAPI) GetBoardByName(string) (*youtrack.Agile, error)        { return m.board, m.boardErr }
func (m *mockAPI) GetBoardForView(string) (*youtrack.Agile, error)       { return m.board, m.boardErr }
func (m *mockAPI) ListProjects() ([]youtrack.Project, error)             { return nil, nil }
func (m *mockAPI) ResolveUser(string) (string, error)                    { return "", nil }
func (m *mockAPI) UpdateIssue(string, string) error                      { return nil }
func (m *mockAPI) ListComments(string) ([]youtrack.Comment, error)       { return nil, nil }
func (m *mockAPI) AddComment(string, string) (*youtrack.Comment, error)  { return nil, nil }
func (m *mockAPI) CreateIssue(string, string, string) (*youtrack.Issue, error) { return nil, nil }
func (m *mockAPI) GetIssueStates(string) ([]youtrack.StateBundleElement, error) { return nil, nil }
func (m *mockAPI) SetIssueState(_ string, state string) error { m.stateSet = state; return nil }
func (m *mockAPI) GetSprintBoard(string, string) (*youtrack.SprintBoard, error) {
	return m.sprintBoard, m.sbErr
}

func TestLoadQueryPath(t *testing.T) {
	api := &mockAPI{
		board:  &youtrack.Agile{ID: "b1", Name: "Board", CurrentSprint: &youtrack.Sprint{ID: "s1", Name: "Sprint 1"}},
		issues: []youtrack.Issue{{IDReadable: "T-1"}},
	}
	f := New(api)

	r, err := f.Load("Board", "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if r.Board == nil || r.Board.ID != "b1" {
		t.Error("expected board")
	}
	if len(r.Issues) != 1 {
		t.Errorf("issues = %d, want 1", len(r.Issues))
	}
	if r.SprintBoard != nil {
		t.Error("expected no sprint board on query path")
	}
}

func TestLoadSprintBoardPath(t *testing.T) {
	sb := &youtrack.SprintBoard{
		Columns: []youtrack.SprintBoardColumn{
			{Cells: []youtrack.BoardCell{{Issues: []youtrack.Issue{{IDReadable: "T-1"}, {IDReadable: "T-2"}}}}},
		},
	}
	api := &mockAPI{
		board: &youtrack.Agile{
			ID:               "b1",
			Name:             "Board",
			CurrentSprint:    &youtrack.Sprint{ID: "s1", Name: "Sprint 1"},
			SwimlaneSettings: &youtrack.AgileSwimlaneSetting{Enabled: true},
		},
		sprintBoard: sb,
	}
	f := New(api)

	r, err := f.Load("Board", "")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if r.SprintBoard == nil {
		t.Fatal("expected sprint board")
	}
	if len(r.Issues) != 2 {
		t.Errorf("flattened issues = %d, want 2", len(r.Issues))
	}
}

func TestRefresh(t *testing.T) {
	board := &youtrack.Agile{ID: "b1", Name: "Board", CurrentSprint: &youtrack.Sprint{ID: "s1", Name: "Sprint 1"}}
	api := &mockAPI{issues: []youtrack.Issue{{IDReadable: "T-1"}}}
	f := New(api)

	r, err := f.Refresh(board, "")
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if len(r.Issues) != 1 {
		t.Errorf("issues = %d, want 1", len(r.Issues))
	}
}

func TestSetStateAndRefresh(t *testing.T) {
	board := &youtrack.Agile{ID: "b1", Name: "Board", CurrentSprint: &youtrack.Sprint{ID: "s1", Name: "Sprint 1"}}
	api := &mockAPI{issues: []youtrack.Issue{{IDReadable: "T-1"}}}
	f := New(api)

	r, err := f.SetStateAndRefresh("T-1", "Done", board, "")
	if err != nil {
		t.Fatalf("SetStateAndRefresh: %v", err)
	}
	if api.stateSet != "Done" {
		t.Errorf("state not set: %q", api.stateSet)
	}
	if len(r.Issues) != 1 {
		t.Errorf("issues = %d, want 1", len(r.Issues))
	}
}

func TestResolveSprintByName(t *testing.T) {
	sb := &youtrack.SprintBoard{
		Columns: []youtrack.SprintBoardColumn{
			{Cells: []youtrack.BoardCell{{Issues: []youtrack.Issue{{IDReadable: "T-1"}}}}},
		},
	}
	api := &mockAPI{
		board: &youtrack.Agile{
			ID:               "b1",
			Name:             "Board",
			Sprints:          []youtrack.Sprint{{ID: "s1", Name: "Sprint 1"}, {ID: "s2", Name: "Sprint 2"}},
			SwimlaneSettings: &youtrack.AgileSwimlaneSetting{Enabled: true},
		},
		sprintBoard: sb,
	}
	f := New(api)

	r, err := f.Load("Board", "Sprint 2")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if r.SprintBoard == nil {
		t.Error("expected sprint board with named sprint")
	}
}
