package reviewer

import (
	"testing"
)

func TestParseFrontmatter_valid(t *testing.T) {
	data := "---\nslug: go\nname: Go\ndescription: Reviews Go code\n---\n\n## Body\n"

	fm, body, err := parseFrontmatter(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Slug != "go" {
		t.Errorf("slug: got %q, want %q", fm.Slug, "go")
	}
	if fm.Name != "Go" {
		t.Errorf("name: got %q, want %q", fm.Name, "Go")
	}
	if fm.Description != "Reviews Go code" {
		t.Errorf("description: got %q, want %q", fm.Description, "Reviews Go code")
	}
	if body != "\n## Body\n" {
		t.Errorf("body: got %q, want %q", body, "\n## Body\n")
	}
}

func TestParseFrontmatter_noOpeningDelimiter(t *testing.T) {
	_, _, err := parseFrontmatter("slug: go\nname: Go\n---\n")
	if err == nil {
		t.Fatal("expected error for missing opening ---")
	}
}

func TestParseFrontmatter_noClosingDelimiter(t *testing.T) {
	_, _, err := parseFrontmatter("---\nslug: go\nname: Go\ndescription: x\n")
	if err == nil {
		t.Fatal("expected error for missing closing ---")
	}
}

func TestParseFrontmatter_missingSlug(t *testing.T) {
	_, _, err := parseFrontmatter("---\nname: Go\ndescription: Reviews Go code\n---\n")
	if err == nil {
		t.Fatal("expected error for missing slug")
	}
}

func TestParseFrontmatter_missingName(t *testing.T) {
	_, _, err := parseFrontmatter("---\nslug: go\ndescription: Reviews Go code\n---\n")
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestParseFrontmatter_missingDescription(t *testing.T) {
	_, _, err := parseFrontmatter("---\nslug: go\nname: Go\n---\n")
	if err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestParseFrontmatter_slugWithComma(t *testing.T) {
	_, _, err := parseFrontmatter("---\nslug: go,rust\nname: Go\ndescription: x\n---\n")
	if err == nil {
		t.Fatal("expected error for slug containing comma")
	}
}

func TestParseFrontmatter_bodyAfterSecondDelimiter(t *testing.T) {
	// A --- inside the body should not be treated as another delimiter.
	data := "---\nslug: go\nname: Go\ndescription: x\n---\n\nsome text\n---\nmore text\n"
	_, body, err := parseFrontmatter(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "\nsome text\n---\nmore text\n" {
		t.Errorf("body: got %q", body)
	}
}
