package pullrequest

import (
	"fmt"
	"regexp"
	"strconv"
)

// PullRequest identifies a GitHub pull request.
type PullRequest struct {
	Owner  string
	Repo   string
	Number int
}

// Client provides pull-request-related operations.
type Client struct{}

var pattern = regexp.MustCompile(`^https?://[^/]+/([^/]+)/([^/]+)/pull/(\d+)`)

// ParseURL extracts owner, repo, and PR number from a GitHub pull request URL
// such as https://github.com/owner/repo/pull/42.
func (c *Client) ParseURL(rawURL string) (PullRequest, error) {
	m := pattern.FindStringSubmatch(rawURL)
	if m == nil {
		return PullRequest{}, fmt.Errorf("invalid pull request URL: %s", rawURL)
	}
	n, err := strconv.Atoi(m[3])
	if err != nil {
		return PullRequest{}, fmt.Errorf("invalid pull request number: %w", err)
	}
	return PullRequest{Owner: m[1], Repo: m[2], Number: n}, nil
}
