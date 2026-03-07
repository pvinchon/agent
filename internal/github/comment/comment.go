package comment

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	gh "github.com/google/go-github/v84/github"
	"github.com/pvinchon/agent/internal/github/pullrequest"
)

// Comment represents a single review comment on a pull request.
type Comment struct {
	ID     int64
	Author string
	Body   string
	Path   string
	Line   int

	// unexported; used internally for threading and sorting.
	inReplyTo int64
	createdAt time.Time
}

// Client provides comment-related operations against the GitHub API.
type Client struct {
	github *gh.Client
}

// BodyContains reports whether the comment body contains the given substring.
func (c Comment) BodyContains(s string) bool {
	return strings.Contains(c.Body, s)
}

// NewClient creates a comment Client from a configured *gh.Client.
func NewClient(ghClient *gh.Client) *Client {
	return &Client{github: ghClient}
}

func fromGitHub(c *gh.PullRequestComment) Comment {
	cm := Comment{
		ID:        c.GetID(),
		Author:    c.GetUser().GetLogin(),
		Body:      c.GetBody(),
		Path:      c.GetPath(),
		createdAt: c.GetCreatedAt().Time,
		inReplyTo: c.GetInReplyTo(),
	}
	if c.Line != nil {
		cm.Line = *c.Line
	} else if c.OriginalLine != nil {
		cm.Line = *c.OriginalLine
	}
	return cm
}

// List fetches all review comments for a pull request, handling pagination
// automatically.
func (c *Client) List(ctx context.Context, pr pullrequest.PullRequest) ([]Comment, error) {
	opts := &gh.PullRequestListCommentsOptions{
		ListOptions: gh.ListOptions{PerPage: 100},
	}
	var all []Comment
	for {
		comments, resp, err := c.github.PullRequests.ListComments(ctx, pr.Owner, pr.Repo, pr.Number, opts)
		if err != nil {
			return nil, fmt.Errorf("list review comments: %w", err)
		}
		for _, cm := range comments {
			all = append(all, fromGitHub(cm))
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

// Thread returns all comments that belong to the same review thread as
// target, sorted by creation time. It resolves deep reply chains
// (A -> B -> C) to find the root ancestor.
func Thread(all []Comment, target Comment) []Comment {
	byID := make(map[int64]Comment)
	for _, cm := range all {
		byID[cm.ID] = cm
	}

	cache := make(map[int64]int64)
	var root func(int64) int64
	root = func(id int64) int64 {
		if r, ok := cache[id]; ok {
			return r
		}
		cm, ok := byID[id]
		if !ok || cm.inReplyTo == 0 {
			cache[id] = id
			return id
		}
		r := root(cm.inReplyTo)
		cache[id] = r
		return r
	}

	rootID := root(target.ID)

	var thread []Comment
	for _, cm := range all {
		if root(cm.ID) == rootID {
			thread = append(thread, cm)
		}
	}

	sort.Slice(thread, func(i, j int) bool {
		return thread[i].createdAt.Before(thread[j].createdAt)
	})

	return thread
}

// Format returns a human-readable representation of a list of comments,
// starting with the file path and line number of the first comment,
// followed by each comment with its author.
func Format(comments []Comment) string {
	if len(comments) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s#L%d\n", comments[0].Path, comments[0].Line)
	for _, cm := range comments {
		fmt.Fprintf(&b, "  @%s: %s\n", cm.Author, cm.Body)
	}
	return b.String()
}
