package compat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alessandrojcm/pampax-go/internal/chunks"
)

func TestNodeFixtureChunkFilesMatchSHA(t *testing.T) {
	chunkDir := filepath.Join("..", "fixtures", "small", ".pampa", "chunks")
	files, err := filepath.Glob(filepath.Join(chunkDir, "*.gz"))
	if err != nil {
		t.Fatalf("glob chunk files: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("expected gzip chunk fixtures in %s", chunkDir)
	}

	for _, chunkPath := range files {
		chunkPath := chunkPath
		t.Run(filepath.Base(chunkPath), func(t *testing.T) {
			raw, err := os.ReadFile(chunkPath)
			if err != nil {
				t.Fatalf("read chunk file: %v", err)
			}

			content, err := chunks.Decompress(raw)
			if err != nil {
				t.Fatalf("decompress chunk: %v", err)
			}

			expectedSHA := strings.TrimSuffix(filepath.Base(chunkPath), ".gz")
			gotSHA := chunks.ComputeSHA(string(content))
			if gotSHA != expectedSHA {
				t.Fatalf("SHA mismatch for %s: got %s, want %s", chunkPath, gotSHA, expectedSHA)
			}
		})
	}
}

func TestNodeFixtureEncryptedChunkHeaderIfPresent(t *testing.T) {
	chunkDir := filepath.Join("..", "fixtures", "small", ".pampa", "chunks")
	files, err := filepath.Glob(filepath.Join(chunkDir, "*.gz.enc"))
	if err != nil {
		t.Fatalf("glob encrypted chunk files: %v", err)
	}
	if len(files) == 0 {
		t.Skip("no encrypted chunk fixtures present")
	}

	for _, chunkPath := range files {
		payload, err := os.ReadFile(chunkPath)
		if err != nil {
			t.Fatalf("read encrypted chunk: %v", err)
		}

		if len(payload) < 7 {
			t.Fatalf("encrypted chunk %s is too short", chunkPath)
		}
		if string(payload[:7]) != "PAMPAE1" {
			t.Fatalf("encrypted chunk %s missing PAMPAE1 header", chunkPath)
		}
	}
}
