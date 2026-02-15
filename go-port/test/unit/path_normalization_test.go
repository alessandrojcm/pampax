package unit

import (
	"testing"

	"github.com/alessandrojcm/pampax-go/internal/codemap"
)

func TestNormalizeChunkMetadataConvertsWindowsPathToForwardSlashes(t *testing.T) {
	meta := codemap.NormalizeChunkMetadata(codemap.ChunkMetadata{
		File: "src\\services\\auth\\login.ts",
		SHA:  "abc",
		Lang: "typescript",
	})

	if meta.File != "src/services/auth/login.ts" {
		t.Fatalf("normalized file path = %q, want %q", meta.File, "src/services/auth/login.ts")
	}
}

func TestNormalizeChunkMetadataRemovesDotSlashPrefix(t *testing.T) {
	meta := codemap.NormalizeChunkMetadata(codemap.ChunkMetadata{
		File: "./src/cli/main.go",
		SHA:  "abc",
		Lang: "go",
	})

	if meta.File != "src/cli/main.go" {
		t.Fatalf("normalized file path = %q, want %q", meta.File, "src/cli/main.go")
	}
}
