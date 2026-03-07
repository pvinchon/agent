package github

import (
	gh "github.com/google/go-github/v84/github"
	"github.com/pvinchon/agent/internal/github/comment"
	"github.com/pvinchon/agent/internal/github/pullrequest"
	"github.com/pvinchon/agent/internal/github/reaction"
)

// Client is the top-level GitHub API client. Use NewClient to create one.
type Client struct {
	Comment     *comment.Client
	Reaction    *reaction.Client
	PullRequest *pullrequest.Client
}

// NewClient creates a Client from an already-configured *gh.Client.
func NewClient(ghClient *gh.Client) *Client {
	return &Client{
		Comment:     comment.NewClient(ghClient),
		Reaction:    reaction.NewClient(ghClient),
		PullRequest: &pullrequest.Client{},
	}
}
