package compat

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNodeFixtureDatabaseContract(t *testing.T) {
	dbPath := filepath.Join("..", "fixtures", "small", ".pampa", "pampa.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("fixture database missing at %s: %v", dbPath, err)
	}

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open fixture database: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	assertPragmaEqualsInt(t, database, "page_size", 4096)
	assertPragmaEqualsString(t, database, "journal_mode", "delete")
	assertPragmaEqualsString(t, database, "encoding", "UTF-8")
	assertPragmaEqualsInt(t, database, "foreign_keys", 0)

	var chunkCount int
	if err := database.QueryRow(`SELECT COUNT(1) FROM code_chunks`).Scan(&chunkCount); err != nil {
		t.Fatalf("count code_chunks rows: %v", err)
	}
	if chunkCount == 0 {
		t.Fatal("expected non-empty code_chunks table")
	}

	var nullSymbols int
	if err := database.QueryRow(`SELECT COUNT(1) FROM code_chunks WHERE symbol IS NULL`).Scan(&nullSymbols); err != nil {
		t.Fatalf("count NULL symbols: %v", err)
	}
	if nullSymbols != 0 {
		t.Fatalf("expected no NULL symbols, got %d", nullSymbols)
	}

	var embeddingJSON []byte
	var dimensions int
	if err := database.QueryRow(`
		SELECT embedding, embedding_dimensions
		FROM code_chunks
		WHERE embedding IS NOT NULL
		LIMIT 1
	`).Scan(&embeddingJSON, &dimensions); err != nil {
		t.Fatalf("select sample embedding row: %v", err)
	}

	if strings.Contains(string(embeddingJSON), ", ") {
		t.Fatalf("embedding JSON contains unexpected whitespace: %s", string(embeddingJSON))
	}

	var vector []float64
	if err := json.Unmarshal(embeddingJSON, &vector); err != nil {
		t.Fatalf("unmarshal embedding JSON: %v", err)
	}
	if len(vector) != dimensions {
		t.Fatalf("embedding length mismatch: got %d, want %d", len(vector), dimensions)
	}
}

func assertPragmaEqualsInt(t *testing.T, database *sql.DB, pragma string, expected int) {
	t.Helper()

	var value int
	if err := database.QueryRow("PRAGMA " + pragma + ";").Scan(&value); err != nil {
		t.Fatalf("query pragma %s: %v", pragma, err)
	}
	if value != expected {
		t.Fatalf("PRAGMA %s = %d, want %d", pragma, value, expected)
	}
}

func assertPragmaEqualsString(t *testing.T, database *sql.DB, pragma string, expected string) {
	t.Helper()

	var value string
	if err := database.QueryRow("PRAGMA " + pragma + ";").Scan(&value); err != nil {
		t.Fatalf("query pragma %s: %v", pragma, err)
	}
	if value != expected {
		t.Fatalf("PRAGMA %s = %q, want %q", pragma, value, expected)
	}
}
