package prompt

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Prompt holds the text content of a prompt loaded from a path.
type Prompt struct{ content string }

// New constructs a Prompt from a string. Used in tests to avoid filesystem/network access.
func New(content string) Prompt { return Prompt{content: content} }

// String returns the prompt content.
func (p Prompt) String() string { return p.content }

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Load reads a prompt from path, which may be a local file path or an HTTP/HTTPS URL.
// Whitespace is trimmed from path before dispatch.
func Load(path string) (Prompt, error) {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return loadURL(path)
	}
	return loadFile(path)
}

func loadURL(url string) (Prompt, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return Prompt{}, fmt.Errorf("load %q: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Prompt{}, fmt.Errorf("load %q: status %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Prompt{}, fmt.Errorf("load %q: %w", url, err)
	}
	return Prompt{content: string(data)}, nil
}

func loadFile(path string) (Prompt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Prompt{}, fmt.Errorf("load %q: %w", path, err)
	}
	return Prompt{content: string(data)}, nil
}
