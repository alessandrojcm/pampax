package unit

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alessandrojcm/pampax-go/internal/db"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

func TestSchemaPragmasMatchContract(t *testing.T) {
	database := setupSchemaDB(t)

	var pageSize int
	if err := database.QueryRow("PRAGMA page_size;").Scan(&pageSize); err != nil {
		t.Fatalf("query page_size pragma: %v", err)
	}
	if pageSize != 4096 {
		t.Fatalf("page_size = %d, want 4096", pageSize)
	}

	var journalMode string
	if err := database.QueryRow("PRAGMA journal_mode;").Scan(&journalMode); err != nil {
		t.Fatalf("query journal_mode pragma: %v", err)
	}
	if journalMode != "delete" {
		t.Fatalf("journal_mode = %q, want %q", journalMode, "delete")
	}

	var encoding string
	if err := database.QueryRow("PRAGMA encoding;").Scan(&encoding); err != nil {
		t.Fatalf("query encoding pragma: %v", err)
	}
	if encoding != "UTF-8" {
		t.Fatalf("encoding = %q, want %q", encoding, "UTF-8")
	}
}

func TestSchemaCreatesAllTablesAndIndices(t *testing.T) {
	database := setupSchemaDB(t)

	expectedTables := []string{"code_chunks", "intention_cache", "query_patterns"}
	for _, tableName := range expectedTables {
		assertSQLiteObjectExists(t, database, "table", tableName)
	}

	expectedIndexes := []string{
		"idx_file_path",
		"idx_symbol",
		"idx_lang",
		"idx_provider",
		"idx_chunk_type",
		"idx_pampa_tags",
		"idx_pampa_intent",
		"idx_lang_provider",
		"idx_query_normalized",
		"idx_target_sha",
		"idx_usage_count",
		"idx_pattern_frequency",
	}

	for _, indexName := range expectedIndexes {
		assertSQLiteObjectExists(t, database, "index", indexName)
	}
}

func TestValidateChunkJSONFieldsWarnsAndSkipsInvalidValues(t *testing.T) {
	originalLogger := log.Logger
	var logBuffer bytes.Buffer
	log.Logger = zerolog.New(&logBuffer)
	t.Cleanup(func() {
		log.Logger = originalLogger
	})

	validArray := `["auth","login"]`
	invalidArray := `{"not":"array"}`
	invalidObject := `not-json`

	validated := db.ValidateChunkJSONFields(&validArray, &invalidArray, &invalidObject)

	if validated.PampaTags == nil || *validated.PampaTags != validArray {
		t.Fatalf("expected valid pampa_tags to be preserved")
	}
	if validated.VariablesUsed != nil {
		t.Fatalf("expected invalid variables_used to be skipped")
	}
	if validated.ContextInfo != nil {
		t.Fatalf("expected invalid context_info to be skipped")
	}

	logs := logBuffer.String()
	if !strings.Contains(logs, "invalid JSON field, skipping") {
		t.Fatalf("expected warning log for invalid fields, got: %s", logs)
	}
	if !strings.Contains(logs, `"field":"variables_used"`) {
		t.Fatalf("expected variables_used warning, got: %s", logs)
	}
	if !strings.Contains(logs, `"field":"context_info"`) {
		t.Fatalf("expected context_info warning, got: %s", logs)
	}
}

func TestEmbeddingBLOBSerializationUsesJSONBytes(t *testing.T) {
	database := setupSchemaDB(t)

	embedding := []float64{0.029445774853229523, -0.0034673467744141817, 0.007123}
	embeddingBytes, err := json.Marshal(embedding)
	if err != nil {
		t.Fatalf("marshal embedding: %v", err)
	}
	if bytes.Contains(embeddingBytes, []byte(", ")) {
		t.Fatalf("embedding JSON must not contain whitespace: %s", string(embeddingBytes))
	}

	if _, err := database.Exec(`
		INSERT INTO code_chunks (
			id, file_path, symbol, sha, lang, embedding, embedding_provider, embedding_dimensions
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "chunk-embedding", "src/search.go", "", "abc123", "go", embeddingBytes, "OpenAI", len(embedding)); err != nil {
		t.Fatalf("insert code_chunks row: %v", err)
	}

	var stored []byte
	if err := database.QueryRow(`SELECT embedding FROM code_chunks WHERE id = ?`, "chunk-embedding").Scan(&stored); err != nil {
		t.Fatalf("query embedding blob: %v", err)
	}

	if !bytes.Equal(stored, embeddingBytes) {
		t.Fatalf("stored embedding bytes differ\n got: %s\nwant: %s", string(stored), string(embeddingBytes))
	}
}

func TestSymbolStoredAsEmptyStringNotNull(t *testing.T) {
	database := setupSchemaDB(t)

	if _, err := database.Exec(`
		INSERT INTO code_chunks (id, file_path, symbol, sha, lang)
		VALUES (?, ?, ?, ?, ?)
	`, "chunk-symbol", "src/no_symbol.go", "", "deadbeef", "go"); err != nil {
		t.Fatalf("insert code_chunks row: %v", err)
	}

	var symbol string
	var symbolIsNull int
	if err := database.QueryRow(`
		SELECT symbol, symbol IS NULL
		FROM code_chunks
		WHERE id = ?
	`, "chunk-symbol").Scan(&symbol, &symbolIsNull); err != nil {
		t.Fatalf("query symbol row: %v", err)
	}

	if symbol != "" {
		t.Fatalf("symbol = %q, want empty string", symbol)
	}
	if symbolIsNull != 0 {
		t.Fatalf("symbol should not be NULL")
	}
}

func setupSchemaDB(t *testing.T) *sql.DB {
	t.Helper()

	schemaPath := filepath.Join("..", "..", "sql", "schema.sql")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema file: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "pampa.db")
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})

	if _, err := database.Exec(string(schemaBytes)); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	return database
}

func assertSQLiteObjectExists(t *testing.T, database *sql.DB, objectType string, objectName string) {
	t.Helper()

	var count int
	if err := database.QueryRow(`
		SELECT COUNT(1)
		FROM sqlite_master
		WHERE type = ? AND name = ?
	`, objectType, objectName).Scan(&count); err != nil {
		t.Fatalf("query sqlite_master for %s %s: %v", objectType, objectName, err)
	}

	if count != 1 {
		t.Fatalf("expected %s %s to exist", objectType, objectName)
	}
}
