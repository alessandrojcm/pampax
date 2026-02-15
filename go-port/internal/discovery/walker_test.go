package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

type noopMatcher struct{}

func (noopMatcher) ShouldSkipDir(string) bool  { return false }
func (noopMatcher) ShouldSkipFile(string) bool { return false }

func TestWalkReturnsDeterministicSortedOutput(t *testing.T) {
	root := t.TempDir()

	mustWriteFile(t, filepath.Join(root, "zeta", "index.ts"))
	mustWriteFile(t, filepath.Join(root, "alpha", "helper.go"))
	mustWriteFile(t, filepath.Join(root, "alpha", "nested", "view.jsx"))
	mustWriteFile(t, filepath.Join(root, "beta", "readme.md"))
	mustWriteFile(t, filepath.Join(root, "beta", "ignore.txt"))

	options := WalkOptions{
		Root:          root,
		Workers:       4,
		SupportedExts: DefaultSupportedExtensions(),
		Matcher:       noopMatcher{},
	}

	first, err := Walk(options)
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	want := []string{
		"alpha/helper.go",
		"alpha/nested/view.jsx",
		"beta/readme.md",
		"zeta/index.ts",
	}

	if !reflect.DeepEqual(first.Paths, want) {
		t.Fatalf("paths mismatch\n got: %#v\nwant: %#v", first.Paths, want)
	}

	for i := 0; i < 20; i++ {
		again, walkErr := Walk(options)
		if walkErr != nil {
			t.Fatalf("walk run %d failed: %v", i+2, walkErr)
		}

		if !reflect.DeepEqual(again.Paths, first.Paths) {
			t.Fatalf("nondeterministic output on run %d\nfirst: %#v\nagain: %#v", i+2, first.Paths, again.Paths)
		}

		if !reflect.DeepEqual(again.Warnings, first.Warnings) {
			t.Fatalf("nondeterministic warnings on run %d\nfirst: %#v\nagain: %#v", i+2, first.Warnings, again.Warnings)
		}
	}
}

func TestWalkSkipsSymlinkTraversal(t *testing.T) {
	root := t.TempDir()

	targetDir := filepath.Join(root, "real")
	mustWriteFile(t, filepath.Join(targetDir, "nested", "inside.ts"))

	symlinkPath := filepath.Join(root, "linkdir")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Skipf("symlink not supported on this platform: %v", err)
	}

	result, err := Walk(WalkOptions{Root: root, SupportedExts: DefaultSupportedExtensions()})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	if len(result.Paths) != 1 || result.Paths[0] != "real/nested/inside.ts" {
		t.Fatalf("unexpected paths: %#v", result.Paths)
	}
}

func TestWalkReportsBrokenSymlinkWarning(t *testing.T) {
	root := t.TempDir()

	brokenPath := filepath.Join(root, "broken.ts")
	if err := os.Symlink(filepath.Join(root, "missing-target.ts"), brokenPath); err != nil {
		t.Skipf("symlink not supported on this platform: %v", err)
	}

	result, err := Walk(WalkOptions{Root: root, SupportedExts: DefaultSupportedExtensions()})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	if len(result.Warnings) == 0 {
		t.Fatalf("expected at least one warning for broken symlink")
	}

	found := false
	for _, warning := range result.Warnings {
		if warning.Path == "broken.ts" && warning.Code == WarningBrokenSymlink {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected broken symlink warning, got %#v", result.Warnings)
	}
}

func TestWalkReportsPermissionDeniedWarning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission simulation is not stable on Windows")
	}

	root := t.TempDir()
	restrictedDir := filepath.Join(root, "restricted")
	if err := os.MkdirAll(restrictedDir, 0o755); err != nil {
		t.Fatalf("create restricted directory: %v", err)
	}

	mustWriteFile(t, filepath.Join(root, "public", "ok.ts"))
	if err := os.Chmod(restrictedDir, 0o000); err != nil {
		t.Fatalf("chmod restricted directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(restrictedDir, 0o755)
	})

	result, err := Walk(WalkOptions{Root: root, SupportedExts: DefaultSupportedExtensions()})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}

	found := false
	for _, warning := range result.Warnings {
		if warning.Path == "restricted" && warning.Code == WarningPermissionDenied {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected permission denied warning, got %#v", result.Warnings)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}

	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
