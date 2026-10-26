package cmd

import (
	"errors"
	"io"
	"strings"

	"github.com/allbin/yt/internal/youtrack"
)

type mockAPI struct {
	issue          *youtrack.Issue
	issues         []youtrack.Issue
	comments       []youtrack.Comment
	states         []youtrack.StateBundleElement
	board          *youtrack.Agile
	boards         []youtrack.Agile
	projects       []youtrack.Project
	issueErr       error
	boardErr       error
	stateSet       string
	updatedFields  map[string]string
	updateErr      error
	fieldValues    []youtrack.BundleValue
	fieldNames     []string
	projectFields  []youtrack.ProjectField
	linkTypes      []youtrack.LinkType
	createdLinks   []string
	removedLinks   []string
	linkErr        error
	currentUser    *youtrack.User
	currentUserErr error
	sprintIssues   map[string][]string // "agileID|sprintID" -> issue IDs
	addedToSprint  []string            // "agileID|sprintID|issueID"
	removedSprint  []string            // "agileID|sprintID|issueID"
	issueBoards    []youtrack.BoardMembership
	sprintErr      error

	issueFields   []youtrack.ProjectField
	setFields     []youtrack.FieldUpdate
	addedBundle   []string // "field|value"
	uploaded      []string // file names
	createdFields []youtrack.FieldUpdate
	createdTagIDs []string
	createCalled  bool

	tags         []youtrack.Tag // the instance's tags
	createdTags  []string
	tagged       []string // "issueID|tagID"
	untagged     []string // "issueID|tagID"
	tagErr       error
	createTagErr error // CreateTag alone fails

	createdDescription string
	addedComment       string

	activities     []youtrack.FieldActivity
	activityFilter youtrack.FieldActivityFilter
	boardIssues    []string
}

func (m *mockAPI) CurrentUser() (*youtrack.User, error) {
	return m.currentUser, m.currentUserErr
}

func (m *mockAPI) GetIssue(string) (*youtrack.Issue, error)         { return m.issue, m.issueErr }
func (m *mockAPI) ListIssues(string, int) ([]youtrack.Issue, error) { return m.issues, nil }
func (m *mockAPI) ListIssuesCompact(string, int) ([]youtrack.IssueCompact, error) {
	return youtrack.Compacts(m.issues), nil
}
func (m *mockAPI) ListBoards() ([]youtrack.Agile, error)           { return m.boards, nil }
func (m *mockAPI) GetBoard(string) (*youtrack.Agile, error)        { return m.board, m.boardErr }
func (m *mockAPI) GetBoardForView(string) (*youtrack.Agile, error) { return m.board, m.boardErr }
func (m *mockAPI) ListProjects() ([]youtrack.Project, error)       { return m.projects, nil }
func (m *mockAPI) ResolveUser(q string) (string, error)            { return q, nil }
func (m *mockAPI) UpdateIssueFields(_ string, core map[string]string, custom []youtrack.FieldUpdate) error {
	m.updatedFields = core
	m.setFields = custom
	return m.updateErr
}
func (m *mockAPI) ListComments(string) ([]youtrack.Comment, error) { return m.comments, nil }
func (m *mockAPI) AddComment(_ string, text string) (*youtrack.Comment, error) {
	m.addedComment = text
	return &youtrack.Comment{ID: "mock-comment-1", Text: "mock"}, nil
}
func (m *mockAPI) CreateIssue(n youtrack.NewIssue) (*youtrack.Issue, error) {
	m.createCalled = true
	m.createdDescription = n.Description
	m.createdFields = n.Fields
	m.createdTagIDs = n.TagIDs
	return &youtrack.Issue{IDReadable: "PROJ-999", Summary: n.Summary}, nil
}

// ListTags matches the server: a case-insensitive substring query.
func (m *mockAPI) ListTags(query string) ([]youtrack.Tag, error) {
	var out []youtrack.Tag
	for _, t := range m.tags {
		if strings.Contains(strings.ToLower(t.Name), strings.ToLower(query)) {
			out = append(out, t)
		}
	}
	return out, m.tagErr
}
func (m *mockAPI) CreateTag(name string) (*youtrack.Tag, error) {
	if err := errors.Join(m.tagErr, m.createTagErr); err != nil {
		return nil, err
	}
	tag := youtrack.Tag{ID: "tag-" + name, Name: name}
	m.tags = append(m.tags, tag)
	m.createdTags = append(m.createdTags, name)
	return &tag, nil
}
func (m *mockAPI) TagIssue(issueID, tagID string) error {
	m.tagged = append(m.tagged, issueID+"|"+tagID)
	return m.tagErr
}
func (m *mockAPI) UntagIssue(issueID, tagID string) error {
	m.untagged = append(m.untagged, issueID+"|"+tagID)
	return m.tagErr
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
func (m *mockAPI) ListSprintIssues(agileID, sprintID string) ([]string, error) {
	return m.sprintIssues[agileID+"|"+sprintID], m.sprintErr
}
func (m *mockAPI) AddIssueToSprint(agileID, sprintID, issueID string) error {
	m.addedToSprint = append(m.addedToSprint, agileID+"|"+sprintID+"|"+issueID)
	return m.sprintErr
}
func (m *mockAPI) RemoveIssueFromSprint(agileID, sprintID, issueID string) error {
	m.removedSprint = append(m.removedSprint, agileID+"|"+sprintID+"|"+issueID)
	return m.sprintErr
}
func (m *mockAPI) IssueBoards(string) ([]youtrack.BoardMembership, error) {
	return m.issueBoards, m.sprintErr
}
func (m *mockAPI) ListBoardIssues(*youtrack.Agile, *youtrack.Sprint, string) ([]youtrack.Issue, error) {
	return m.issues, nil
}

func (m *mockAPI) BoardIssues(*youtrack.Agile) ([]string, error) {
	return m.boardIssues, m.sprintErr
}
func (m *mockAPI) ListFieldActivities(f youtrack.FieldActivityFilter) ([]youtrack.FieldActivity, error) {
	m.activityFilter = f
	return m.activities, nil
}
func (m *mockAPI) ListAttachments(string) ([]youtrack.Attachment, error) { return nil, nil }
func (m *mockAPI) DownloadAttachment(string, io.Writer) error            { return nil }
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
func (m *mockAPI) ListLinkTypes() ([]youtrack.LinkType, error) {
	return m.linkTypes, m.linkErr
}
func (m *mockAPI) CreateLink(source, phrase, target string) error {
	m.createdLinks = append(m.createdLinks, source+"|"+phrase+"|"+target)
	return m.linkErr
}
func (m *mockAPI) RemoveLink(source, linkID, target string) error {
	m.removedLinks = append(m.removedLinks, source+"|"+linkID+"|"+target)
	return m.linkErr
}
func (m *mockAPI) ListIssueFields(string) ([]youtrack.ProjectField, error) {
	return m.issueFields, nil
}
func (m *mockAPI) AddBundleValue(f youtrack.ProjectField, name string) error {
	m.addedBundle = append(m.addedBundle, f.Name+"|"+name)
	return m.updateErr
}
func (m *mockAPI) UploadAttachments(_ string, files []youtrack.UploadFile) ([]youtrack.Attachment, error) {
	var out []youtrack.Attachment
	for _, f := range files {
		m.uploaded = append(m.uploaded, f.Name)
		out = append(out, youtrack.Attachment{Name: f.Name})
	}
	return out, nil
}
