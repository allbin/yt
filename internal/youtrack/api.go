package youtrack

// API defines the YouTrack API surface. Used by CLI commands and future TUI.
type API interface {
	GetIssue(id string) (*Issue, error)
	ListIssues(query string, limit int) ([]Issue, error)
	ListBoards() ([]Agile, error)
	GetBoardByName(name string) (*Agile, error)
	ListProjects() ([]Project, error)
	ResolveUser(query string) (string, error)
}
