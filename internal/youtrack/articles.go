package youtrack

import (
	"cmp"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Article is a knowledge base article. Content is markdown, like an issue's
// description.
type Article struct {
	ID              string        `json:"id,omitempty"`
	IDReadable      string        `json:"idReadable"`
	Summary         string        `json:"summary"`
	Content         string        `json:"content,omitempty"`
	Created         int64         `json:"created,omitempty"`
	Updated         int64         `json:"updated,omitempty"`
	Reporter        *User         `json:"reporter,omitempty"`
	UpdatedBy       *User         `json:"updatedBy,omitempty"`
	Project         *Project      `json:"project,omitempty"`
	Parent          *ArticleRef   `json:"parentArticle"`
	Children        []ArticleRef  `json:"childArticles,omitempty"`
	HasChildren     bool          `json:"hasChildren"`
	Ordinal         int           `json:"ordinal"`
	Tags            []Tag         `json:"tags,omitempty"`
	Attachments     []Attachment  `json:"attachments,omitempty"`
	MentionedIssues []LinkedIssue `json:"mentionedIssues,omitempty"`
}

// ArticleRef is the minimal reference to an article: a parent, a child, or
// an article that mentions an issue.
type ArticleRef struct {
	ID         string `json:"id,omitempty"`
	IDReadable string `json:"idReadable"`
	Summary    string `json:"summary"`
	Ordinal    int    `json:"ordinal,omitempty"` // place among its siblings; set on children only
}

// articleIDPattern matches a readable article ID: the project short name, -A-,
// a number (VIU-A-14). A short name cannot hold a hyphen, so an issue ID
// (VIU-14) never matches.
var articleIDPattern = regexp.MustCompile(`(?i)^[a-z0-9_]+-a-[0-9]+$`)

// IsArticleID reports whether id is a readable article ID.
func IsArticleID(id string) bool { return articleIDPattern.MatchString(id) }

// entityPath is the API path of the issue or article id names. Comments and
// attachments live under either, so the ID alone picks the endpoint.
func entityPath(id string) string {
	if IsArticleID(id) {
		return "/api/articles/" + url.PathEscape(id)
	}
	return "/api/issues/" + url.PathEscape(id)
}

const articleRefFields = "id,idReadable,summary"

const articleListFields = "id,idReadable,summary,created,updated,hasChildren,ordinal," +
	"reporter(login,fullName),project(id,shortName,name)," +
	"parentArticle(" + articleRefFields + ")"

const articleDetailFields = articleListFields + ",content," +
	"updatedBy(login,fullName),tags(id,name)," +
	"childArticles(" + articleRefFields + ",ordinal)," +
	"attachments(" + attachmentFields + ")," +
	"mentionedIssues(id,idReadable,summary)"

func (c *Client) GetArticle(id string) (*Article, error) {
	data, err := c.get("/api/articles/"+url.PathEscape(id), url.Values{"fields": {articleDetailFields}})
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", id, err)
	}
	var a Article
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parse %s: %w", id, err)
	}
	// The server returns children unordered; show them as the sidebar does.
	slices.SortStableFunc(a.Children, func(x, y ArticleRef) int { return cmp.Compare(x.Ordinal, y.Ordinal) })
	return &a, nil
}

// ListArticles searches articles with YouTrack query syntax (project: X,
// free text, mentions: PROJ-123). An empty query lists every article.
func (c *Client) ListArticles(query string, limit int) ([]Article, error) {
	params := url.Values{"fields": {articleListFields}}
	if query != "" {
		params.Set("query", query)
	}
	if limit > 0 {
		params.Set("$top", strconv.Itoa(limit))
	}
	data, err := c.get("/api/articles", params)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}
	var articles []Article
	if err := json.Unmarshal(data, &articles); err != nil {
		return nil, fmt.Errorf("parse articles: %w", err)
	}
	return articles, nil
}

// ListChildArticles lists the direct sub-articles of id.
func (c *Client) ListChildArticles(id string, limit int) ([]Article, error) {
	params := url.Values{"fields": {articleListFields}}
	if limit > 0 {
		params.Set("$top", strconv.Itoa(limit))
	}
	data, err := c.get("/api/articles/"+url.PathEscape(id)+"/childArticles", params)
	if err != nil {
		return nil, fmt.Errorf("list sub-articles of %s: %w", id, err)
	}
	var articles []Article
	if err := json.Unmarshal(data, &articles); err != nil {
		return nil, fmt.Errorf("parse sub-articles: %w", err)
	}
	return articles, nil
}

// mentionPageSize is how many search hits ArticlesMentioning reads per request.
var mentionPageSize = 100

// ArticlesMentioning lists the articles whose content mentions issueID, a
// readable ID. The search can match loosely, so every page of hits is read
// and each is checked against the article's own list of mentioned issues.
func (c *Client) ArticlesMentioning(issueID string) ([]ArticleRef, error) {
	params := url.Values{
		"fields": {articleRefFields + ",mentionedIssues(idReadable)"},
		"query":  {"mentions: " + issueID},
		"$top":   {strconv.Itoa(mentionPageSize)},
	}
	var refs []ArticleRef
	for skip := 0; ; skip += mentionPageSize {
		params.Set("$skip", strconv.Itoa(skip))
		data, err := c.get("/api/articles", params)
		if err != nil {
			return nil, fmt.Errorf("articles mentioning %s: %w", issueID, err)
		}
		var page []Article
		if err := json.Unmarshal(data, &page); err != nil {
			return nil, fmt.Errorf("parse articles: %w", err)
		}
		for _, a := range page {
			if slices.ContainsFunc(a.MentionedIssues, func(i LinkedIssue) bool { return strings.EqualFold(i.IDReadable, issueID) }) {
				refs = append(refs, ArticleRef{ID: a.ID, IDReadable: a.IDReadable, Summary: a.Summary})
			}
		}
		if len(page) < mentionPageSize {
			return refs, nil
		}
	}
}

// NewArticle is what CreateArticle sends. ParentID is the parent's database
// id (109-147), not its readable ID.
type NewArticle struct {
	Project  string // short name
	Summary  string
	Content  string
	ParentID string
}

func (c *Client) CreateArticle(n NewArticle) (*Article, error) {
	type idRef struct {
		ID string `json:"id"`
	}
	body := struct {
		Project struct {
			ShortName string `json:"shortName"`
		} `json:"project"`
		Summary string `json:"summary"`
		Content string `json:"content,omitempty"`
		Parent  *idRef `json:"parentArticle,omitempty"`
	}{Summary: n.Summary, Content: n.Content}
	body.Project.ShortName = n.Project
	if n.ParentID != "" {
		body.Parent = &idRef{ID: n.ParentID}
	}

	data, err := c.postJSON("/api/articles?fields="+url.QueryEscape(articleDetailFields), body)
	if err != nil {
		return nil, fmt.Errorf("create article: %w", err)
	}
	var a Article
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parse created article: %w", err)
	}
	return &a, nil
}

// UpdateArticle writes the given core fields (summary, content) in one
// request.
func (c *Client) UpdateArticle(id string, fields map[string]string) error {
	if err := c.post("/api/articles/"+url.PathEscape(id), fields); err != nil {
		return fmt.Errorf("update %s: %w", id, err)
	}
	return nil
}

func (c *Client) DeleteArticle(id string) error {
	if err := c.delete("/api/articles/" + url.PathEscape(id)); err != nil {
		return fmt.Errorf("delete %s: %w", id, err)
	}
	return nil
}
