package reviewer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// readPrompt reads prompt content from a local file or remote URL.
func readPrompt(source string) (string, error) {
	if strings.HasPrefix(source, "https://") {
		return fetchPrompt(source)
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// httpClient is a package-level variable so tests can swap it.
// Do not use t.Parallel() in tests that replace this client.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// fetchPrompt downloads prompt content from a remote URL.
func fetchPrompt(url string) (string, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
