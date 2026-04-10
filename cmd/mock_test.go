package cmd

import (
	"io"

	"github.com/allbin/yt/internal/youtrack"
)

type mockAPI struct {
	issue     *youtrack.Issue
	issues    []youtrack.Issue
	comments  []youtrack.Comment
	states    []youtrack.StateBundleElement
	board     *youtrack.Agile
	boards    []youtrack.Agile
	projects  []youtrack.Project
	issueErr  error
	boardErr  error
	stateSet      string
	command       string
	updateErr     error
	fieldValues   []youtrack.BundleValue
	fieldNames    []string
	projectFields []youtrack.ProjectField
}

func (m *mockAPI) GetIssue(string) (*youtrack.Issue, error)         { return m.issue, m.issueErr }
func (m *mockAPI) ListIssues(string, int) ([]youtrack.Issue, error) { return m.issues, nil }
func (m *mockAPI) ListBoards() ([]youtrack.Agile, error)            { return m.boards, nil }
func (m *mockAPI) GetBoardByName(string) (*youtrack.Agile, error)   { return m.board, m.boardErr }
func (m *mockAPI) GetBoardForView(string) (*youtrack.Agile, error)  { return m.board, m.boardErr }
func (m *mockAPI) ListProjects() ([]youtrack.Project, error)        { return m.projects, nil }
func (m *mockAPI) ResolveUser(q string) (string, error)             { return q, nil }
func (m *mockAPI) UpdateIssue(_ string, cmd string) error {
	m.command = cmd
	return m.updateErr
}
func (m *mockAPI) ListComments(string) ([]youtrack.Comment, error)  { return m.comments, nil }
func (m *mockAPI) AddComment(string, string) (*youtrack.Comment, error) {
	return &youtrack.Comment{ID: "mock-comment-1", Text: "mock"}, nil
}
func (m *mockAPI) CreateIssue(_, summary, _ string, tags []string) (*youtrack.Issue, error) {
	issue := &youtrack.Issue{IDReadable: "PROJ-999", Summary: summary}
	for _, t := range tags {
		issue.Tags = append(issue.Tags, youtrack.Tag{Name: t})
	}
	return issue, nil
}
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
	return m.fieldValues, nil
}
func (m *mockAPI) GetProjectFieldValues(string, string) ([]youtrack.BundleValue, error) {
	return m.fieldValues, nil
}
func (m *mockAPI) ListProjectFields(string) ([]youtrack.ProjectField, error) {
	return m.projectFields, nil
}
func (m *mockAPI) ListFieldNames(string) ([]string, error) { return m.fieldNames, nil }
