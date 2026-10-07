package youtrack

import "io"

// API defines the YouTrack API surface. Used by CLI commands and future TUI.
type API interface {
	GetIssue(id string) (*Issue, error)
	ListIssues(query string, limit int) ([]Issue, error)
	ListBoards() ([]Agile, error)
	GetBoardByName(name string) (*Agile, error)
	GetBoardForView(name string) (*Agile, error)
	ListProjects() ([]Project, error)
	CurrentUser() (*User, error)
	ResolveUser(query string) (string, error)
	UpdateIssue(id string, command string) error
	UpdateIssueFields(id string, core map[string]string, custom []FieldUpdate) error
	ListComments(issueID string) ([]Comment, error)
	AddComment(issueID, text string) (*Comment, error)
	CreateIssue(project, summary, description string, fields []FieldUpdate) (*Issue, error)
	GetIssueStates(issueID string) ([]StateBundleElement, error)
	SetIssueState(issueID, stateName string) error
	GetFieldValues(issueID, fieldName string) ([]BundleValue, error)
	GetProjectFieldValues(projectID, fieldName string) ([]BundleValue, error)
	ListProjectFields(projectID string) ([]ProjectField, error)
	ListFieldNames(issueID string) ([]string, error)
	ListIssueFields(issueID string) ([]ProjectField, error)
	AddBundleValue(field ProjectField, name string) error
	GetSprintBoard(boardID, sprintID string) (*SprintBoard, error)
	ListSprintIssues(agileID, sprintID string) ([]string, error)
	AddIssueToSprint(agileID, sprintID, idReadable string) error
	RemoveIssueFromSprint(agileID, sprintID, issueID string) error
	IssueBoards(issueID string) ([]BoardMembership, error)
	BoardIssues(board *Agile) ([]string, error)
	ListFieldActivities(f FieldActivityFilter) ([]FieldActivity, error)
	ListAttachments(issueID string) ([]Attachment, error)
	DownloadAttachment(url string, w io.Writer) error
	UploadAttachments(issueID string, files []UploadFile) ([]Attachment, error)
	ListLinkTypes() ([]LinkType, error)
	CreateLink(sourceID, phrase, targetID string) error
	RemoveLink(sourceID, linkID, targetID string) error
}
