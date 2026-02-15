package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreMatcherPrecedenceDefaultThenGitThenPampa(t *testing.T) {
	root := t.TempDir()
	mustWriteIgnoreFile(t, filepath.Join(root, ".gitignore"), "!data.json\n")
	mustWriteIgnoreFile(t, filepath.Join(root, ".pampignore"), "data.json\n")

	matcher, err := NewIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("new ignore matcher: %v", err)
	}

	decision := matcher.DecisionFor("data.json", false)
	if !decision.Excluded {
		t.Fatalf("expected data.json to be excluded by pampignore precedence, got %#v", decision)
	}

	if decision.Source != RuleSourcePampIgnore {
		t.Fatalf("expected pampignore source, got %s", decision.Source)
	}
}

func TestIgnoreMatcherDefaultPatternsMatchRootFiles(t *testing.T) {
	root := t.TempDir()

	matcher, err := NewIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("new ignore matcher: %v", err)
	}

	if !matcher.ShouldSkipFile("config.json") {
		t.Fatalf("expected default pattern **/*.json to match root file")
	}

	if !matcher.ShouldSkipFile("script.sh") {
		t.Fatalf("expected default pattern **/*.sh to match root file")
	}
}

func TestIgnoreMatcherNestedGitignoreWithNegation(t *testing.T) {
	root := t.TempDir()
	mustWriteIgnoreFile(t, filepath.Join(root, ".gitignore"), "src/generated.ts\nsrc/nested/ignored/**\n")
	mustWriteIgnoreFile(t, filepath.Join(root, "src", "nested", ".gitignore"), "ignored/**\n!ignored/reinclude.js\n")

	matcher, err := NewIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("new ignore matcher: %v", err)
	}

	cases := []struct {
		path        string
		wantExclude bool
	}{
		{path: "src/generated.ts", wantExclude: true},
		{path: "src/nested/ignored/a.js", wantExclude: true},
		{path: "src/nested/ignored/reinclude.js", wantExclude: false},
	}

	for _, tc := range cases {
		decision := matcher.DecisionFor(tc.path, false)
		if decision.Excluded != tc.wantExclude {
			t.Fatalf("unexpected decision for %s: got excluded=%v, want=%v, decision=%#v", tc.path, decision.Excluded, tc.wantExclude, decision)
		}
	}
}

func TestIgnoreMatcherAnchoredAndDirectoryOnlyPatterns(t *testing.T) {
	root := t.TempDir()
	mustWriteIgnoreFile(t, filepath.Join(root, ".gitignore"), "/rootonly/\nlogs/\n")

	matcher, err := NewIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("new ignore matcher: %v", err)
	}

	if !matcher.ShouldSkipDir("rootonly") {
		t.Fatalf("expected /rootonly/ anchored pattern to skip root directory")
	}

	if matcher.ShouldSkipDir("src/rootonly") {
		t.Fatalf("did not expect /rootonly/ anchored pattern to skip src/rootonly")
	}

	if !matcher.ShouldSkipDir("src/logs") {
		t.Fatalf("expected logs/ directory-only pattern to skip nested logs dir")
	}

	if !matcher.ShouldSkipFile("src/logs/app.ts") {
		t.Fatalf("expected logs/ directory-only pattern to skip descendants")
	}
}

func TestWalkUsesIgnoreMatcher(t *testing.T) {
	root := t.TempDir()
	mustWriteIgnoreFile(t, filepath.Join(root, ".gitignore"), "src/ignored/*\n!src/ignored/keep.ts\n")
	mustWriteFile(t, filepath.Join(root, "src", "ignored", "drop.ts"))
	mustWriteFile(t, filepath.Join(root, "src", "ignored", "keep.ts"))
	mustWriteFile(t, filepath.Join(root, "src", "ok.ts"))

	matcher, err := NewIgnoreMatcher(root)
	if err != nil {
		t.Fatalf("new ignore matcher: %v", err)
	}

	result, err := Walk(WalkOptions{Root: root, SupportedExts: DefaultSupportedExtensions(), Matcher: matcher})
	if err != nil {
		t.Fatalf("walk with matcher failed: %v", err)
	}

	if len(result.Paths) != 2 {
		t.Fatalf("unexpected path count: got %d (%#v)", len(result.Paths), result.Paths)
	}

	if result.Paths[0] != "src/ignored/keep.ts" || result.Paths[1] != "src/ok.ts" {
		t.Fatalf("unexpected paths: %#v", result.Paths)
	}
}

func mustWriteIgnoreFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
