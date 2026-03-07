package reaction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	gh "github.com/google/go-github/v84/github"
	"github.com/pvinchon/agent/internal/github/comment"
	"github.com/pvinchon/agent/internal/github/pullrequest"
)

func newTestClient(t *testing.T) (*http.ServeMux, *Client) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	ghClient := gh.NewClient(srv.Client()).WithAuthToken("test-token")
	u, _ := url.Parse(srv.URL + "/")
	ghClient.BaseURL = u
	return mux, NewClient(ghClient)
}

func TestAdd(t *testing.T) {
	mux, client := newTestClient(t)
	pr := pullrequest.PullRequest{Owner: "owner", Repo: "repo", Number: 1}

	mux.HandleFunc("/repos/owner/repo/pulls/comments/100/reactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["content"] != "eyes" {
			t.Errorf("reaction content = %v, want eyes", body["content"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 999, "content": "eyes"})
	})

	ctx := context.Background()
	r, err := client.Add(ctx, pr, comment.Comment{ID: 100}, Eyes)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if r.id != 999 {
		t.Errorf("reaction.id = %d, want 999", r.id)
	}
}

func TestAdd_Error(t *testing.T) {
	mux, client := newTestClient(t)
	pr := pullrequest.PullRequest{Owner: "owner", Repo: "repo", Number: 1}

	mux.HandleFunc("/repos/owner/repo/pulls/comments/100/reactions", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
	})

	ctx := context.Background()
	_, err := client.Add(ctx, pr, comment.Comment{ID: 100}, "invalid")
	if err == nil {
		t.Fatal("expected error for 422 response")
	}
}

func TestRemove(t *testing.T) {
	mux, client := newTestClient(t)
	pr := pullrequest.PullRequest{Owner: "owner", Repo: "repo", Number: 1}

	mux.HandleFunc("/repos/owner/repo/pulls/comments/100/reactions/999", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	if err := client.Remove(ctx, pr, comment.Comment{ID: 100}, Reaction{id: 999}); err != nil {
		t.Fatalf("Remove: %v", err)
	}
}

func TestRemove_Error(t *testing.T) {
	mux, client := newTestClient(t)
	pr := pullrequest.PullRequest{Owner: "owner", Repo: "repo", Number: 1}

	mux.HandleFunc("/repos/owner/repo/pulls/comments/100/reactions/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	ctx := context.Background()
	if err := client.Remove(ctx, pr, comment.Comment{ID: 100}, Reaction{id: 999}); err == nil {
		t.Fatal("expected error for 404 response")
	}
}
