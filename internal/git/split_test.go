package git

import (
	"testing"
)

const testDiff = `diff --git a/internal/foo/foo.go b/internal/foo/foo.go
index abc..def 100644
--- a/internal/foo/foo.go
+++ b/internal/foo/foo.go
@@ -1,3 +1,4 @@
 package foo
+
+// added line
diff --git a/internal/foo/foo_test.go b/internal/foo/foo_test.go
index 111..222 100644
--- a/internal/foo/foo_test.go
+++ b/internal/foo/foo_test.go
@@ -1,3 +1,4 @@
 package foo
+// test change
diff --git a/main.go b/main.go
index 333..444 100644
--- a/main.go
+++ b/main.go
@@ -1 +1,2 @@
 package main
+// root change`

func TestSplitByFile(t *testing.T) {
	got := SplitByFile(testDiff)

	if len(got) != 3 {
		t.Fatalf("got %d files, want 3", len(got))
	}

	files := []string{"internal/foo/foo.go", "internal/foo/foo_test.go", "main.go"}
	for _, f := range files {
		if _, ok := got[f]; !ok {
			t.Errorf("missing file %q in result", f)
		}
	}

	if !containsLine(got["internal/foo/foo.go"], "diff --git a/internal/foo/foo.go b/internal/foo/foo.go") {
		t.Error("file diff should start with its diff --git header")
	}
	if !containsLine(got["internal/foo/foo.go"], "+// added line") {
		t.Error("file diff should contain its added line")
	}
	if containsLine(got["internal/foo/foo.go"], "diff --git a/main.go") {
		t.Error("file diff should not contain another file's diff header")
	}
}

func TestSplitByFile_empty(t *testing.T) {
	got := SplitByFile("")
	if len(got) != 0 {
		t.Errorf("got %d files for empty diff, want 0", len(got))
	}
}

func TestSplitByFile_singleFile(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1 +1 @@
-old
+new`
	got := SplitByFile(diff)
	if len(got) != 1 {
		t.Fatalf("got %d files, want 1", len(got))
	}
	if _, ok := got["foo.go"]; !ok {
		t.Error("missing file foo.go")
	}
}

func TestSplitByFolder(t *testing.T) {
	got := SplitByFolder(testDiff)

	// Expect two folders: "internal/foo" and "."
	if len(got) != 2 {
		t.Fatalf("got %d folders, want 2: %v", len(got), keys(got))
	}
	if _, ok := got["internal/foo"]; !ok {
		t.Error("missing folder internal/foo")
	}
	if _, ok := got["."]; !ok {
		t.Error("missing root folder .")
	}

	// The "internal/foo" folder diff should contain both files
	fooFolderDiff := got["internal/foo"]
	if !containsLine(fooFolderDiff, "diff --git a/internal/foo/foo.go b/internal/foo/foo.go") {
		t.Error("folder diff should contain foo.go diff header")
	}
	if !containsLine(fooFolderDiff, "diff --git a/internal/foo/foo_test.go b/internal/foo/foo_test.go") {
		t.Error("folder diff should contain foo_test.go diff header")
	}
}

func TestSplitByFolder_empty(t *testing.T) {
	got := SplitByFolder("")
	if len(got) != 0 {
		t.Errorf("got %d folders for empty diff, want 0", len(got))
	}
}

func containsLine(s, line string) bool {
	for _, l := range splitLines(s) {
		if l == line {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func keys(m map[string]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
