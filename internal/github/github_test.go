package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	gh "github.com/google/go-github/v84/github"
	"github.com/pvinchon/agent/internal/github/comment"
	"github.com/pvinchon/agent/internal/github/reaction"
)

// TestEndToEndFlow exercises the composable sub-clients together, simulating
// the full workflow: parse URL -> list comments -> find mention -> thread ->
// react -> format -> swap reaction.
func TestEndToEndFlow(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	ghClient := gh.NewClient(srv.Client()).WithAuthToken("test-token")
	ghClient.BaseURL, _ = url.Parse(srv.URL + "/")
	client := NewClient(ghClient)

	pr, err := client.PullRequest.ParseURL("https://github.com/owner/repo/pull/5")
	if err != nil {
		t.Fatalf("ParseURL: %v", err)
	}
	if pr.Owner != "owner" || pr.Repo != "repo" || pr.Number != 5 {
		t.Fatalf("parsed = %+v", pr)
	}

	apiComments := []*gh.PullRequestComment{
		{
			ID:        gh.Ptr[int64](10),
			Body:      gh.Ptr("Hey @pvinchon-agent please review this"),
			Path:      gh.Ptr("service.go"),
			Line:      gh.Ptr(15),
			User:      &gh.User{Login: gh.Ptr("alice")},
			CreatedAt: &gh.Timestamp{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			ID:        gh.Ptr[int64](20),
			Body:      gh.Ptr("I second this"),
			InReplyTo: gh.Ptr[int64](10),
			User:      &gh.User{Login: gh.Ptr("bob")},
			CreatedAt: &gh.Timestamp{Time: time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)},
		},
	}

	mux.HandleFunc("/repos/owner/repo/pulls/5/comments", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(apiComments)
	})

	var addedReaction string
	mux.HandleFunc("/repos/owner/repo/pulls/comments/10/reactions", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		content, _ := body["content"].(string)
		addedReaction = content
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": 555, "content": content})
	})

	var removedReaction bool
	mux.HandleFunc("/repos/owner/repo/pulls/comments/10/reactions/555", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			removedReaction = true
			w.WriteHeader(http.StatusNoContent)
		}
	})

	ctx := context.Background()

	// Step 1: Fetch all comments.
	comments, err := client.Comment.List(ctx, pr)
	if err != nil {
		t.Fatalf("Comment.List: %v", err)
	}

	// Step 2: Find the comment mentioning @pvinchon-agent.
	var target comment.Comment
	for _, cm := range comments {
		if cm.BodyContains("@pvinchon-agent") {
			target = cm
			break
		}
	}
	if target.ID == 0 {
		t.Fatal("no comment mentioning @pvinchon-agent")
	}
	if target.ID != 10 {
		t.Fatalf("target.ID = %d, want 10", target.ID)
	}

	// Step 3: Get the full thread.
	thread := comment.Thread(comments, target)
	if len(thread) != 2 {
		t.Fatalf("got %d comments in thread, want 2", len(thread))
	}

	// Step 4: Add eyes reaction.
	reactionResult, err := client.Reaction.Add(ctx, pr, target, reaction.Eyes)
	if err != nil {
		t.Fatalf("Reaction.Add(eyes): %v", err)
	}
	if addedReaction != "eyes" {
		t.Errorf("server received reaction = %q, want 'eyes'", addedReaction)
	}

	// Step 5: Format and verify output.
	output := comment.Format(thread)
	if !strings.Contains(output, "service.go#L15") {
		t.Errorf("output missing file/line header: %s", output)
	}
	if !strings.Contains(output, "@alice") || !strings.Contains(output, "@bob") {
		t.Errorf("output missing authors: %s", output)
	}

	// Step 6: Remove eyes reaction.
	if err := client.Reaction.Remove(ctx, pr, target, reactionResult); err != nil {
		t.Fatalf("Reaction.Remove: %v", err)
	}
	if !removedReaction {
		t.Error("expected reaction to be removed")
	}

	// Step 7: Add +1 reaction.
	_, err = client.Reaction.Add(ctx, pr, target, reaction.ThumbsUp)
	if err != nil {
		t.Fatalf("Reaction.Add(+1): %v", err)
	}
	if addedReaction != "+1" {
		t.Errorf("server received reaction = %q, want '+1'", addedReaction)
	}
}
