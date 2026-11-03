package youtrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// HistoryEntry is one change in an issue's or article's activity history.
type HistoryEntry struct {
	Time      string `json:"time"`
	Timestamp int64  `json:"timestamp"`
	Author    *User  `json:"author"`
	// Kind is what changed: created, summary, description, content, field,
	// link, tag, attachment, attachment-rename, comment, comment-edit,
	// comment-attachment, resolved, sprint, project, visibility, commit, or
	// pull-request.
	Kind string `json:"kind"`
	// Field names the custom field, link type, or board; for a comment kind,
	// the comment's ID.
	Field   string `json:"field,omitempty"`
	Removed string `json:"removed,omitempty"`
	Added   string `json:"added,omitempty"`
}

// historyKinds maps each activity category read for a history to its kind.
// An unknown category is accepted by the server and returns nothing, so each
// one here was seen to return activities.
var issueHistoryKinds = map[string]string{
	"IssueCreatedCategory":       "created",
	"SummaryCategory":            "summary",
	"DescriptionCategory":        "description",
	"CustomFieldCategory":        "field",
	"LinksCategory":              "link",
	"TagsCategory":               "tag",
	"AttachmentsCategory":        "attachment",
	"AttachmentRenameCategory":   "attachment-rename",
	"CommentsCategory":           "comment",
	"CommentTextCategory":        "comment-edit",
	"CommentAttachmentsCategory": "comment-attachment",
	"IssueResolvedCategory":      "resolved",
	"SprintCategory":             "sprint",
	"ProjectCategory":            "project",
	"IssueVisibilityCategory":    "visibility",
	"VcsChangeCategory":          "commit",
	"PullRequestChangeCategory":  "pull-request",
}

var articleHistoryKinds = map[string]string{
	"ArticleCreatedCategory":            "created",
	"ArticleSummaryCategory":            "summary",
	"ArticleDescriptionCategory":        "content",
	"ArticleCommentsCategory":           "comment",
	"ArticleTagsCategory":               "tag",
	"ArticleAttachmentsCategory":        "attachment",
	"ArticleCommentAttachmentsCategory": "comment-attachment",
}

const historyFields = "id,timestamp,author(login,fullName),category(id),field(name),target(id)," +
	"added(id,idReadable,name,login,fullName,presentation,text,version)," +
	"removed(id,idReadable,name,login,fullName,presentation,text,version)"

type historyActivity struct {
	Timestamp int64 `json:"timestamp"`
	Author    *User `json:"author"`
	Category  struct {
		ID string `json:"id"`
	} `json:"category"`
	Field  *ActivityField `json:"field"`
	Target struct {
		ID string `json:"id"`
	} `json:"target"`
	Added   json.RawMessage `json:"added"`
	Removed json.RawMessage `json:"removed"`
}

// History returns every recorded change to the issue or article id, oldest
// first.
func (c *Client) History(id string) ([]HistoryEntry, error) {
	kinds := issueHistoryKinds
	if IsArticleID(id) {
		kinds = articleHistoryKinds
	}
	cats := make([]string, 0, len(kinds))
	for c := range kinds {
		cats = append(cats, c)
	}
	params := url.Values{
		"categories": {strings.Join(cats, ",")},
		"fields":     {"afterCursor,hasAfter,activities(" + historyFields + ")"},
	}

	acts, err := allActivities[historyActivity](c, entityPath(id)+"/activitiesPage", params)
	if err != nil {
		return nil, fmt.Errorf("history of %s: %w", id, err)
	}
	out := make([]HistoryEntry, 0, len(acts))
	for _, a := range acts {
		out = append(out, historyEntry(a, kinds[a.Category.ID]))
	}
	return out, nil
}

func historyEntry(a historyActivity, kind string) HistoryEntry {
	e := HistoryEntry{
		Time:      time.UnixMilli(a.Timestamp).Format(time.RFC3339),
		Timestamp: a.Timestamp,
		Author:    a.Author,
		Kind:      kind,
		Added:     activityValue(a.Added),
		Removed:   activityValue(a.Removed),
	}
	switch kind {
	case "field", "link", "sprint":
		if a.Field != nil {
			e.Field = a.Field.Name
		}
	case "comment":
		e.Field = commentID(a.Added, a.Removed)
	case "comment-edit", "comment-attachment":
		e.Field = a.Target.ID
	case "resolved":
		// The values are resolution timestamps: one added means resolved, none
		// means reopened.
		e.Removed = ""
		e.Added = "resolved"
		if isNull(a.Added) {
			e.Added = "reopened"
		}
	case "created":
		e.Added, e.Removed = "", ""
	}
	return e
}

// activityValue renders an activity's added or removed value: text as is, an
// entity by its display key, a list joined with commas.
func activityValue(raw json.RawMessage) string {
	if isNull(raw) {
		return ""
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return ""
	}
	return renderValue(v)
}

func renderValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case bool:
		return fmt.Sprint(x)
	case []any:
		parts := make([]string, 0, len(x))
		for _, item := range x {
			if s := renderValue(item); s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	case map[string]any:
		for _, key := range []string{"idReadable", "name", "fullName", "presentation", "text", "version", "login"} {
			if s, ok := x[key].(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// commentID is the ID of the comment an added or removed comment activity
// carries.
func commentID(added, removed json.RawMessage) string {
	for _, raw := range []json.RawMessage{added, removed} {
		var items []struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &items) == nil && len(items) > 0 {
			return items[0].ID
		}
	}
	return ""
}

func isNull(raw json.RawMessage) bool {
	s := strings.TrimSpace(string(raw))
	return s == "" || s == "null" || s == "[]"
}
