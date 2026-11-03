package youtrack

import "io"

// API defines the YouTrack API surface. Used by CLI commands and future TUI.
type API interface {
	GetIssue(id string) (*Issue, error)
	ListIssues(query string, limit int) ([]Issue, error)
	ListIssuesCompact(query string, limit int) ([]IssueCompact, error)
	ListBoards() ([]Agile, error)
	GetBoard(ref string) (*Agile, error)
	GetBoardForView(ref string) (*Agile, error)
	ListProjects() ([]Project, error)
	CurrentUser() (*User, error)
	ResolveUser(query string) (string, error)
	UpdateIssueFields(id string, core map[string]string, custom []FieldUpdate) error
	// ListComments, AddComment, ListAttachments and UploadAttachments take
	// an issue ID (PROJ-12) or an article ID (PROJ-A-12).
	ListComments(id string) ([]Comment, error)
	AddComment(id, text string) (*Comment, error)
	UpdateComment(id, commentID, text string) (*Comment, error)
	DeleteComment(id, commentID string) error
	CreateIssue(n NewIssue) (*Issue, error)
	ListTags(query string) ([]Tag, error)
	CreateTag(name string) (*Tag, error)
	// AddTag and RemoveTag take an issue or article ID.
	AddTag(id, tagID string) error
	RemoveTag(id, tagID string) error
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
	ListBoardIssues(board *Agile, sprint *Sprint, filter string) ([]Issue, error)
	ListFieldActivities(f FieldActivityFilter) ([]FieldActivity, error)
	ListAttachments(id string) ([]Attachment, error)
	DownloadAttachment(url string, w io.Writer) error
	UploadAttachments(id string, files []UploadFile) ([]Attachment, error)
	ListLinkTypes() ([]LinkType, error)
	CreateLink(sourceID, phrase, targetID string) error
	RemoveLink(sourceID, linkID, targetID string) error
	GetArticle(id string) (*Article, error)
	ListArticles(query string, limit int) ([]Article, error)
	ListChildArticles(id string, limit int) ([]Article, error)
	ArticlesMentioning(issueID string) ([]ArticleRef, error)
	CreateArticle(n NewArticle) (*Article, error)
	UpdateArticle(id string, u ArticleUpdate) error
	DeleteArticle(id string) error
	History(id string) ([]HistoryEntry, error)
}
