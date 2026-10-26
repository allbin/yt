package youtrack

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	// sprintsSettingsFields is what Kind and Membership need.
	sprintsSettingsFields = "sprintsSettings(disableSprints,isExplicit,explicitQuery,sprintSyncField(name))"

	agileFields = "id,name,projects(shortName),currentSprint(id,name)," +
		"sprints(id,name,start,finish)," + sprintsSettingsFields + ",columnSettings(field(name))"
	agileBoardView = "id,name,projects(shortName)," +
		"currentSprint(id,name),sprints(id,name)," + sprintsSettingsFields + "," +
		"columnSettings(field(name),columns(presentation,ordinal,fieldValues(name,isResolved)))," +
		"swimlaneSettings(enabled,field(name),values(name))"
	sprintBoardFields = "columns(agileColumn(presentation),cells(row(id,name,$type," +
		"issue(idReadable,summary)),issues(idReadable,summary,description,resolved,created,updated," +
		"tags(name),customFields(name,value(name,text,presentation,login,fullName)))))"
)

// boardIDPattern matches a board's database id, such as 124-66.
var boardIDPattern = regexp.MustCompile(`^\d+-\d+$`)

// ListBoards returns the boards /api/agiles lists: those shared with the user.
// It leaves out a board private to its owner, which the user may still be able
// to open by id.
func (c *Client) ListBoards() ([]Agile, error) {
	return c.listBoards(agileFields)
}

// GetBoard resolves a board by name, id or URL (see resolveBoard), with its
// sprint list populated for sprint lookup.
func (c *Client) GetBoard(ref string) (*Agile, error) {
	return c.resolveBoard(ref, agileFields)
}

// GetBoardForView resolves a board by name, id or URL (see resolveBoard), with
// expanded column/swimlane settings.
func (c *Client) GetBoardForView(ref string) (*Agile, error) {
	return c.resolveBoard(ref, agileBoardView)
}

// resolveBoard finds a board from a URL (https://<host>/agiles/124-66/...), an
// id (124-66) or a case-insensitive name. A URL or id is fetched directly, so
// it reaches boards the list leaves out. An id that matches no board is tried
// as a name too, for a board named like one.
func (c *Client) resolveBoard(ref, fields string) (*Agile, error) {
	if id, ok, err := boardIDFromURL(ref); ok {
		if err != nil {
			return nil, err
		}
		return c.getBoard(id, fields)
	}
	if !boardIDPattern.MatchString(ref) {
		return c.findBoardByName(ref, fields)
	}
	board, err := c.getBoard(ref, fields)
	if !isNotFound(err) {
		return board, err
	}
	if board, nameErr := c.findBoardByName(ref, fields); nameErr == nil {
		return board, nil
	}
	return nil, fmt.Errorf("board %q not found", ref)
}

// boardIDFromURL extracts the board id from a board URL. ok reports whether
// ref is a URL; err is set for one that names no board.
func boardIDFromURL(ref string) (id string, ok bool, err error) {
	u, perr := url.Parse(ref)
	if perr != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", false, nil
	}
	segs := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, s := range segs {
		if s == "agiles" && i+1 < len(segs) && boardIDPattern.MatchString(segs[i+1]) {
			return segs[i+1], true, nil
		}
	}
	return "", true, fmt.Errorf("no board id in %q; expected a URL like https://<host>/agiles/124-66", ref)
}

func (c *Client) getBoard(id, fields string) (*Agile, error) {
	data, err := c.get("/api/agiles/"+url.PathEscape(id), url.Values{"fields": {fields}})
	if err != nil {
		return nil, fmt.Errorf("get board %s: %w", id, err)
	}

	var board Agile
	if err := json.Unmarshal(data, &board); err != nil {
		return nil, fmt.Errorf("parse board: %w", err)
	}

	return &board, nil
}

func (c *Client) findBoardByName(name, fields string) (*Agile, error) {
	boards, err := c.listBoards(fields)
	if err != nil {
		return nil, err
	}

	for _, b := range boards {
		if strings.EqualFold(b.Name, name) {
			return &b, nil
		}
	}

	return nil, fmt.Errorf("board %q not found among the boards shared with you; "+
		"a board private to its owner is not listed, so pass its id or URL (https://<host>/agiles/<id>) instead", name)
}

func (c *Client) listBoards(fields string) ([]Agile, error) {
	data, err := c.get("/api/agiles", url.Values{"fields": {fields}, "$top": {"-1"}})
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}

	var boards []Agile
	if err := json.Unmarshal(data, &boards); err != nil {
		return nil, fmt.Errorf("parse boards: %w", err)
	}

	return boards, nil
}

// ListBoardIssues lists the issues on a board that also match filter, a
// YouTrack query (may be empty). sprint is the sprint shown, nil on a
// sprintless board. The board is searched by its Board attribute, but search
// knows only boards shared with the user and rejects the attribute of one
// private to its owner; the board's issues are then named by id, read from
// its sprint.
func (c *Client) ListBoardIssues(board *Agile, sprint *Sprint, filter string) ([]Issue, error) {
	sprintName := ""
	if sprint != nil {
		sprintName = sprint.Name
	}
	issues, err := c.ListIssues(joinQuery(board.SearchQuery(sprintName), filter), 0)
	if !isBadRequest(err) {
		return issues, err
	}

	if sprint == nil {
		sprint = board.CurrentSprint
	}
	if sprint == nil || sprint.ID == "" {
		return nil, err
	}
	ids, idErr := c.ListSprintIssues(board.ID, sprint.ID)
	if idErr != nil {
		return nil, idErr
	}
	if len(ids) == 0 {
		return []Issue{}, nil
	}
	return c.ListIssues(joinQuery("issue id: "+strings.Join(ids, ", "), filter), 0)
}

func joinQuery(parts ...string) string {
	return strings.TrimSpace(strings.Join(parts, " "))
}

func isNotFound(err error) bool {
	return hasStatus(err, http.StatusNotFound)
}

func isBadRequest(err error) bool {
	return hasStatus(err, http.StatusBadRequest)
}

func hasStatus(err error, code int) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == code
}

func (c *Client) GetSprintBoard(boardID, sprintID string) (*SprintBoard, error) {
	path := fmt.Sprintf("/api/agiles/%s/sprints/%s/board",
		url.PathEscape(boardID), url.PathEscape(sprintID))
	params := url.Values{"fields": {sprintBoardFields}}

	data, err := c.get(path, params)
	if err != nil {
		return nil, fmt.Errorf("get sprint board: %w", err)
	}

	var board SprintBoard
	if err := json.Unmarshal(data, &board); err != nil {
		return nil, fmt.Errorf("parse sprint board: %w", err)
	}

	return &board, nil
}
