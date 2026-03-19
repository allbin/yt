package youtrack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListComments(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/api/issues/TEST-1/comments" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewEncoder(w).Encode([]Comment{
			{
				ID:      "4-1",
				Text:    "First comment",
				Author:  &User{Login: "alice", FullName: "Alice"},
				Created: 1700000000000,
			},
			{
				ID:      "4-2",
				Text:    "Second comment",
				Author:  &User{Login: "bob", FullName: "Bob"},
				Created: 1700000060000,
			},
		}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	comments, err := client.ListComments("TEST-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 2 {
		t.Fatalf("got %d comments, want 2", len(comments))
	}
	if comments[0].ID != "4-1" {
		t.Errorf("comments[0].ID = %q, want 4-1", comments[0].ID)
	}
	if comments[0].Author.Login != "alice" {
		t.Errorf("comments[0].Author.Login = %q, want alice", comments[0].Author.Login)
	}
}

func TestListCommentsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode([]Comment{}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	comments, err := client.ListComments("TEST-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 0 {
		t.Errorf("got %d comments, want 0", len(comments))
	}
}

func TestAddComment(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var req struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatal(err)
		}
		if req.Text != "Hello" {
			t.Errorf("text = %q, want Hello", req.Text)
		}

		if err := json.NewEncoder(w).Encode(Comment{
			ID:      "4-99",
			Text:    "Hello",
			Author:  &User{Login: "me", FullName: "Me"},
			Created: 1700000000000,
		}); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	comment, err := client.AddComment("TEST-1", "Hello")
	if err != nil {
		t.Fatal(err)
	}
	if comment.ID != "4-99" {
		t.Errorf("ID = %q, want 4-99", comment.ID)
	}
	if comment.Text != "Hello" {
		t.Errorf("Text = %q, want Hello", comment.Text)
	}
}
