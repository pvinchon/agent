package comment

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

func TestList(t *testing.T) {
	mux, client := newTestClient(t)
	pr := pullrequest.PullRequest{Owner: "owner", Repo: "repo", Number: 1}

	comments := []*gh.PullRequestComment{
		{
			ID:        gh.Ptr[int64](100),
			Body:      gh.Ptr("first comment"),
			Path:      gh.Ptr("main.go"),
			Line:      gh.Ptr(10),
			User:      &gh.User{Login: gh.Ptr("alice")},
			CreatedAt: &gh.Timestamp{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			ID:        gh.Ptr[int64](200),
			Body:      gh.Ptr("reply to first"),
			Path:      gh.Ptr("main.go"),
			InReplyTo: gh.Ptr[int64](100),
			User:      &gh.User{Login: gh.Ptr("bob")},
			CreatedAt: &gh.Timestamp{Time: time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)},
		},
	}

	mux.HandleFunc("/repos/owner/repo/pulls/1/comments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", got)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	})

	ctx := context.Background()
	got, err := client.List(ctx, pr)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d comments, want 2", len(got))
	}
	if got[0].ID != 100 || got[0].Author != "alice" || got[0].Body != "first comment" {
		t.Errorf("comment[0] = %+v", got[0])
	}
	if got[0].Path != "main.go" || got[0].Line != 10 {
		t.Errorf("comment[0] path/line = %s#%d, want main.go#10", got[0].Path, got[0].Line)
	}
	// Verify threading info is captured internally.
	thread := Thread(got, got[0])
	if len(thread) != 2 {
		t.Errorf("Thread(comment[0]) has %d comments, want 2", len(thread))
	}
}

func TestList_APIError(t *testing.T) {
	mux, client := newTestClient(t)
	pr := pullrequest.PullRequest{Owner: "owner", Repo: "repo", Number: 1}

	mux.HandleFunc("/repos/owner/repo/pulls/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	ctx := context.Background()
	_, err := client.List(ctx, pr)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status code: %v", err)
	}
}

func TestList_Pagination(t *testing.T) {
	mux, client := newTestClient(t)
	pr := pullrequest.PullRequest{Owner: "owner", Repo: "repo", Number: 1}

	page1 := make([]*gh.PullRequestComment, 100)
	for i := range page1 {
		page1[i] = &gh.PullRequestComment{
			ID:   gh.Ptr[int64](int64(i + 1)),
			Body: gh.Ptr("comment"),
			User: &gh.User{Login: gh.Ptr("user")},
		}
	}
	page2 := []*gh.PullRequestComment{
		{
			ID:   gh.Ptr[int64](101),
			Body: gh.Ptr("last"),
			User: &gh.User{Login: gh.Ptr("user")},
		},
	}

	mux.HandleFunc("/repos/owner/repo/pulls/1/comments", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		if page == "2" {
			json.NewEncoder(w).Encode(page2)
		} else {
			w.Header().Set("Link", `<http://test?page=2>; rel="next"`)
			json.NewEncoder(w).Encode(page1)
		}
	})

	ctx := context.Background()
	got, err := client.List(ctx, pr)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 101 {
		t.Fatalf("got %d comments, want 101", len(got))
	}
	if got[100].Body != "last" {
		t.Errorf("last comment body = %q, want 'last'", got[100].Body)
	}
}

func TestThread(t *testing.T) {
	now := time.Now()
	all := []Comment{
		{ID: 1, Author: "alice", Body: "thread 1 root", Path: "a.go", Line: 10, createdAt: now},
		{ID: 2, Author: "bob", Body: "thread 2 root", Path: "b.go", Line: 20, createdAt: now},
		{ID: 3, Author: "carol", Body: "reply to thread 1", inReplyTo: 1, createdAt: now.Add(2 * time.Hour)},
		{ID: 4, Author: "dave", Body: "earlier reply to thread 1", inReplyTo: 1, createdAt: now.Add(1 * time.Hour)},
	}

	// Thread from root comment.
	thread := Thread(all, all[0])
	if len(thread) != 3 {
		t.Fatalf("Thread(root) has %d comments, want 3", len(thread))
	}
	if thread[0].ID != 1 {
		t.Errorf("thread[0] ID = %d, want 1 (root)", thread[0].ID)
	}
	// dave's reply (earlier) should come before carol's.
	if thread[1].ID != 4 {
		t.Errorf("thread[1] ID = %d, want 4 (dave)", thread[1].ID)
	}
	if thread[2].ID != 3 {
		t.Errorf("thread[2] ID = %d, want 3 (carol)", thread[2].ID)
	}

	// Thread from a reply returns the same thread.
	thread2 := Thread(all, all[2])
	if len(thread2) != 3 {
		t.Fatalf("Thread(reply) has %d comments, want 3", len(thread2))
	}
	if thread2[0].ID != 1 {
		t.Errorf("Thread(reply) root ID = %d, want 1", thread2[0].ID)
	}

	// Unrelated comment gets its own thread.
	thread3 := Thread(all, all[1])
	if len(thread3) != 1 {
		t.Fatalf("Thread(unrelated) has %d comments, want 1", len(thread3))
	}
	if thread3[0].ID != 2 {
		t.Errorf("Thread(unrelated)[0] ID = %d, want 2", thread3[0].ID)
	}
}

func TestThread_DeepReplyChain(t *testing.T) {
	now := time.Now()
	all := []Comment{
		{ID: 1, Author: "alice", Body: "root", Path: "a.go", Line: 1, createdAt: now},
		{ID: 2, Author: "bob", Body: "reply to root", inReplyTo: 1, createdAt: now.Add(1 * time.Hour)},
		{ID: 3, Author: "carol", Body: "reply to reply", inReplyTo: 2, createdAt: now.Add(2 * time.Hour)},
	}

	thread := Thread(all, all[2]) // start from deepest reply
	if len(thread) != 3 {
		t.Fatalf("got %d comments, want 3", len(thread))
	}
	if thread[0].ID != 1 || thread[2].ID != 3 {
		t.Errorf("thread = [%d, %d, %d], want [1, 2, 3]", thread[0].ID, thread[1].ID, thread[2].ID)
	}
}

func TestThread_Empty(t *testing.T) {
	thread := Thread(nil, Comment{ID: 99})
	if len(thread) != 0 {
		t.Fatalf("got %d comments for nil input, want 0", len(thread))
	}
}

func TestFormat(t *testing.T) {
	comments := []Comment{
		{Author: "alice", Body: "This needs refactoring", Path: "internal/handler.go", Line: 42},
		{Author: "bob", Body: "I agree, let me fix it"},
		{Author: "alice", Body: "Thanks!"},
	}

	got := Format(comments)
	want := "internal/handler.go#L42\n" +
		"  @alice: This needs refactoring\n" +
		"  @bob: I agree, let me fix it\n" +
		"  @alice: Thanks!\n"
	if got != want {
		t.Errorf("Format:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormat_SingleComment(t *testing.T) {
	comments := []Comment{
		{Author: "user", Body: "LGTM", Path: "main.go", Line: 1},
	}

	got := Format(comments)
	want := "main.go#L1\n  @user: LGTM\n"
	if got != want {
		t.Errorf("Format:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormat_Empty(t *testing.T) {
	if got := Format(nil); got != "" {
		t.Errorf("Format(nil) = %q, want empty", got)
	}
}

func TestBodyContains(t *testing.T) {
	cm := Comment{Body: "please review @pvinchon-agent"}

	if !cm.BodyContains("@pvinchon-agent") {
		t.Error("expected BodyContains to match @pvinchon-agent")
	}
	if !cm.BodyContains("review") {
		t.Error("expected BodyContains to match partial substring")
	}
	if cm.BodyContains("missing") {
		t.Error("expected BodyContains to not match absent substring")
	}
	if cm.BodyContains("") {
		// strings.Contains returns true for empty string; just document behavior.
	} else {
		t.Error("expected BodyContains('') to return true (per strings.Contains)")
	}
}

func TestFromGitHub_OriginalLineFallback(t *testing.T) {
	origLine := 42
	c := &gh.PullRequestComment{
		ID:           gh.Ptr[int64](1),
		Body:         gh.Ptr("comment"),
		Path:         gh.Ptr("main.go"),
		OriginalLine: &origLine,
		User:         &gh.User{Login: gh.Ptr("alice")},
	}
	cm := fromGitHub(c)
	if cm.Line != 42 {
		t.Errorf("Line = %d, want 42 (OriginalLine fallback)", cm.Line)
	}
}

func TestFromGitHub_BothLinesNil(t *testing.T) {
	c := &gh.PullRequestComment{
		ID:   gh.Ptr[int64](1),
		Body: gh.Ptr("comment"),
		Path: gh.Ptr("main.go"),
		User: &gh.User{Login: gh.Ptr("alice")},
	}
	cm := fromGitHub(c)
	if cm.Line != 0 {
		t.Errorf("Line = %d, want 0 when both Line and OriginalLine are nil", cm.Line)
	}
}

func TestThread_SingleComment(t *testing.T) {
	cm := Comment{ID: 1, Author: "alice", Body: "solo", Path: "f.go", Line: 1, createdAt: time.Now()}
	thread := Thread([]Comment{cm}, cm)
	if len(thread) != 1 {
		t.Fatalf("got %d comments, want 1", len(thread))
	}
	if thread[0].ID != 1 {
		t.Errorf("thread[0].ID = %d, want 1", thread[0].ID)
	}
}

func TestThread_TargetNotInAll(t *testing.T) {
	all := []Comment{
		{ID: 1, Author: "alice", Body: "root", Path: "f.go", Line: 1, createdAt: time.Now()},
	}
	target := Comment{ID: 99, Author: "ghost", Body: "missing"}
	thread := Thread(all, target)
	if len(thread) != 0 {
		t.Fatalf("got %d comments, want 0 for target not in all", len(thread))
	}
}
