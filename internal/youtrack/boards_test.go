package youtrack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// privateBoardServer serves a board list that leaves out board 124-66, as
// /api/agiles does for a board private to its owner, while answering for
// 124-66 by id.
func privateBoardServer(t *testing.T) *Client {
	t.Helper()
	listed := []Agile{{ID: "124-12", Name: "SenseView"}}
	private := Agile{ID: "124-66", Name: "Security sweep (DO-33)"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body any
		switch r.URL.Path {
		case "/api/agiles":
			body = listed
		case "/api/agiles/124-12":
			body = listed[0]
		case "/api/agiles/124-66":
			body = private
		default:
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("fields") == "" {
			t.Errorf("%s fetched without fields", r.URL.Path)
		}
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(srv.Close)
	return NewClient(srv.URL, "token")
}

func TestGetBoardByID(t *testing.T) {
	board, err := privateBoardServer(t).GetBoard("124-66")
	if err != nil {
		t.Fatal(err)
	}
	if board.Name != "Security sweep (DO-33)" {
		t.Errorf("Name = %q, want Security sweep (DO-33)", board.Name)
	}
}

func TestGetBoardByURL(t *testing.T) {
	board, err := privateBoardServer(t).GetBoardForView("https://example.youtrack.cloud/agiles/124-66/current")
	if err != nil {
		t.Fatal(err)
	}
	if board.ID != "124-66" {
		t.Errorf("ID = %q, want 124-66", board.ID)
	}
}

func TestGetBoardUnlistedNameSuggestsID(t *testing.T) {
	_, err := privateBoardServer(t).GetBoard("Security sweep (DO-33)")
	if err == nil {
		t.Fatal("expected error for a board missing from the list")
	}
	if !strings.Contains(err.Error(), "id or URL") {
		t.Errorf("error %q does not point at passing the id or URL", err)
	}
}

// TestListBoardIssuesUnsearchableBoard covers a board private to its owner:
// YouTrack's search rejects its Board attribute with a 400, so its issues
// come from its sprint by id.
func TestListBoardIssuesUnsearchableBoard(t *testing.T) {
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body any
		switch r.URL.Path {
		case "/api/agiles/124-66/sprints/125-162/issues":
			body = []Issue{{IDReadable: "DO-33"}, {IDReadable: "DO-34"}}
		case "/api/issues":
			q := r.URL.Query().Get("query")
			queries = append(queries, q)
			if strings.Contains(q, "Board ") {
				w.WriteHeader(http.StatusBadRequest)
				body = map[string]string{"error": "bad_request", "error_description": "Can't parse search query"}
				break
			}
			body = []Issue{{IDReadable: "DO-33"}}
		default:
			http.NotFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	board := &Agile{
		ID: "124-66", Name: "Security sweep (DO-33)",
		SprintsSettings: &SprintsSettings{DisableSprints: true},
		CurrentSprint:   &Sprint{ID: "125-162", Name: "First sprint"},
	}
	issues, err := NewClient(srv.URL, "token").ListBoardIssues(board, nil, "State: {Submitted}")
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].IDReadable != "DO-33" {
		t.Errorf("issues = %+v, want DO-33", issues)
	}
	if want := "issue id: DO-33, DO-34 State: {Submitted}"; queries[len(queries)-1] != want {
		t.Errorf("query = %q, want %q", queries[len(queries)-1], want)
	}
}
