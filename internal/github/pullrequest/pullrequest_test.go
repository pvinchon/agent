package pullrequest

import "testing"

func TestParseURL(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name    string
		url     string
		want    PullRequest
		wantErr bool
	}{
		{
			name: "standard URL",
			url:  "https://github.com/owner/repo/pull/42",
			want: PullRequest{Owner: "owner", Repo: "repo", Number: 42},
		},
		{
			name: "URL with trailing path segments",
			url:  "https://github.com/owner/repo/pull/7/files",
			want: PullRequest{Owner: "owner", Repo: "repo", Number: 7},
		},
		{
			name: "HTTP scheme",
			url:  "http://github.com/my-org/my-repo/pull/100",
			want: PullRequest{Owner: "my-org", Repo: "my-repo", Number: 100},
		},
		{
			name: "GitHub Enterprise",
			url:  "https://github.example.com/corp/project/pull/1",
			want: PullRequest{Owner: "corp", Repo: "project", Number: 1},
		},
		{
			name:    "missing pull number",
			url:     "https://github.com/owner/repo/pull/",
			wantErr: true,
		},
		{
			name:    "not a PR URL",
			url:     "https://github.com/owner/repo/issues/5",
			wantErr: true,
		},
		{
			name:    "empty string",
			url:     "",
			wantErr: true,
		},
		{
			name:    "random text",
			url:     "not a url at all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.ParseURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
