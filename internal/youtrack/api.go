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
	ResolveUser(query string) (string, error)
	UpdateIssue(id string, command string) error
	UpdateIssueFields(id string, fields map[string]string) error
	ListComments(issueID string) ([]Comment, error)
	AddComment(issueID, text string) (*Comment, error)
	CreateIssue(project, summary, description string, tags []string) (*Issue, error)
	GetIssueStates(issueID string) ([]StateBundleElement, error)
	SetIssueState(issueID, stateName string) error
	GetFieldValues(issueID, fieldName string) ([]BundleValue, error)
	GetProjectFieldValues(projectID, fieldName string) ([]BundleValue, error)
	ListProjectFields(projectID string) ([]ProjectField, error)
	ListFieldNames(issueID string) ([]string, error)
	GetSprintBoard(boardID, sprintID string) (*SprintBoard, error)
	ListAttachments(issueID string) ([]Attachment, error)
	DownloadAttachment(url string, w io.Writer) error
	ListLinkTypes() ([]LinkType, error)
	CreateLink(sourceID, phrase, targetID string) error
	RemoveLink(sourceID, linkID, targetID string) error
}
