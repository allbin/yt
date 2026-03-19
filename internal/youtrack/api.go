package youtrack

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
	ListComments(issueID string) ([]Comment, error)
	AddComment(issueID, text string) (*Comment, error)
	CreateIssue(project, summary, description string) (*Issue, error)
	GetIssueStates(issueID string) ([]StateBundleElement, error)
	SetIssueState(issueID, stateName string) error
	GetSprintBoard(boardID, sprintID string) (*SprintBoard, error)
}
