package tui

import (
	"io"

	"github.com/allbin/yt/internal/youtrack"
)

// mockAPI implements youtrack.API for testing.
type mockAPI struct {
	issue    *youtrack.Issue
	issues   []youtrack.Issue
	comments []youtrack.Comment
	states   []youtrack.StateBundleElement
	board    *youtrack.Agile
	issueErr error
	boardErr error
	stateSet string
}

func (m *mockAPI) GetIssue(string) (*youtrack.Issue, error)         { return m.issue, m.issueErr }
func (m *mockAPI) ListIssues(string, int) ([]youtrack.Issue, error) { return m.issues, nil }
func (m *mockAPI) ListBoards() ([]youtrack.Agile, error)            { return nil, nil }
func (m *mockAPI) GetBoardByName(string) (*youtrack.Agile, error)   { return m.board, m.boardErr }
func (m *mockAPI) GetBoardForView(string) (*youtrack.Agile, error)  { return m.board, m.boardErr }
func (m *mockAPI) ListProjects() ([]youtrack.Project, error)        { return nil, nil }
func (m *mockAPI) ResolveUser(string) (string, error)               { return "", nil }
func (m *mockAPI) UpdateIssue(string, string) error                          { return nil }
func (m *mockAPI) UpdateIssueFields(string, map[string]string) error         { return nil }
func (m *mockAPI) ListComments(string) ([]youtrack.Comment, error)  { return m.comments, nil }
func (m *mockAPI) CreateIssue(string, string, string, []string) (*youtrack.Issue, error) {
	return nil, nil
}
func (m *mockAPI) AddComment(string, string) (*youtrack.Comment, error) { return nil, nil }
func (m *mockAPI) GetIssueStates(string) ([]youtrack.StateBundleElement, error) {
	return m.states, nil
}
func (m *mockAPI) SetIssueState(_ string, state string) error {
	m.stateSet = state
	return nil
}
func (m *mockAPI) GetSprintBoard(string, string) (*youtrack.SprintBoard, error) {
	return nil, nil
}
func (m *mockAPI) ListAttachments(string) ([]youtrack.Attachment, error) { return nil, nil }
func (m *mockAPI) DownloadAttachment(string, io.Writer) error           { return nil }
func (m *mockAPI) GetFieldValues(string, string) ([]youtrack.BundleValue, error) {
	return nil, nil
}
func (m *mockAPI) GetProjectFieldValues(string, string) ([]youtrack.BundleValue, error) {
	return nil, nil
}
func (m *mockAPI) ListProjectFields(string) ([]youtrack.ProjectField, error) { return nil, nil }
func (m *mockAPI) ListFieldNames(string) ([]string, error)                   { return nil, nil }
