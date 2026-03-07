package reaction

import (
	"context"
	"fmt"

	gh "github.com/google/go-github/v84/github"
	"github.com/pvinchon/agent/internal/github/comment"
	"github.com/pvinchon/agent/internal/github/pullrequest"
)

// ReactionContent represents a GitHub reaction content type.
type Content string

const (
	ThumbsUp   Content = "+1"
	ThumbsDown Content = "-1"
	Laugh      Content = "laugh"
	Confused   Content = "confused"
	Heart      Content = "heart"
	Hooray     Content = "hooray"
	Rocket     Content = "rocket"
	Eyes       Content = "eyes"
)

type Reaction struct {
	id int64
}

// Client provides reaction-related operations against the GitHub API.
type Client struct {
	github *gh.Client
}

// NewClient creates a reaction Client from a configured *gh.Client.
func NewClient(ghClient *gh.Client) *Client {
	return &Client{github: ghClient}
}

// Add creates a reaction on a pull request review comment and returns the
// reaction ID. Common reaction values: "+1", "-1", "eyes", "rocket", etc.
func (c *Client) Add(ctx context.Context, pr pullrequest.PullRequest, comment comment.Comment, reaction Content) (Reaction, error) {
	r, _, err := c.github.Reactions.CreatePullRequestCommentReaction(ctx, pr.Owner, pr.Repo, comment.ID, string(reaction))
	if err != nil {
		return Reaction{}, fmt.Errorf("add reaction: %w", err)
	}
	return Reaction{id: r.GetID()}, nil
}

// Remove deletes a reaction from a pull request review comment.
func (c *Client) Remove(ctx context.Context, pr pullrequest.PullRequest, comment comment.Comment, reaction Reaction) error {
	_, err := c.github.Reactions.DeletePullRequestCommentReaction(ctx, pr.Owner, pr.Repo, comment.ID, reaction.id)
	if err != nil {
		return fmt.Errorf("remove reaction: %w", err)
	}
	return nil
}
