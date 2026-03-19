package prompt

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	p := New("hello")
	if p.String() != "hello" {
		t.Errorf("got %q, want %q", p.String(), "hello")
	}
}

func TestZeroValue(t *testing.T) {
	var p Prompt
	if p.String() != "" {
		t.Errorf("zero-value Prompt.String() should return empty string, got %q", p.String())
	}
}

func TestLoad_localFile(t *testing.T) {
	f, err := os.CreateTemp("", "prompt-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("hello from file")
	f.Close()

	p, err := Load(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.String() != "hello from file" {
		t.Errorf("got %q, want %q", p.String(), "hello from file")
	}
}

func TestLoad_localFileWhitespace(t *testing.T) {
	f, err := os.CreateTemp("", "prompt-*.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("content")
	f.Close()

	p, err := Load("  " + f.Name() + "  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.String() != "content" {
		t.Errorf("got %q, want %q", p.String(), "content")
	}
}

func TestLoad_fileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "/nonexistent/path.md") {
		t.Errorf("error should include path, got: %v", err)
	}
}

func TestLoad_url(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from url"))
	}))
	defer srv.Close()

	p, err := Load(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.String() != "hello from url" {
		t.Errorf("got %q, want %q", p.String(), "hello from url")
	}
}

func TestLoad_urlNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := Load(srv.URL)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should mention status code, got: %v", err)
	}
}
