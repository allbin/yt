package youtrack

// IssueView is a pre-resolved snapshot of an Issue's display fields.
type IssueView struct {
	ID          string
	Summary     string
	Description string
	State       string
	Priority    string
	Assignee    string
	Type        string
	Subsystem   string
	Tags        string
	IsResolved  bool
}

func (i *Issue) View() IssueView {
	return IssueView{
		ID:          i.IDReadable,
		Summary:     i.Summary,
		Description: i.Desc(),
		State:       i.Field("State"),
		Priority:    i.Field("Priority"),
		Assignee:    i.Field("Assignee"),
		Type:        i.Field("Type"),
		Subsystem:   i.Field("Subsystem"),
		Tags:        i.TagNames(),
		IsResolved:  i.Resolved != nil,
	}
}

// CommentView is a pre-resolved snapshot of a Comment's display fields.
type CommentView struct {
	Author  string
	Created int64
	Text    string
}

func (c *Comment) View() CommentView {
	author := "Unknown"
	if c.Author != nil {
		if c.Author.FullName != "" {
			author = c.Author.FullName
		} else {
			author = c.Author.Login
		}
	}
	return CommentView{Author: author, Created: c.Created, Text: c.Text}
}
