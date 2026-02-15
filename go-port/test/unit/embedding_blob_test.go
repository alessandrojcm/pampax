package unit

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestEmbeddingBlobSerializationNoWhitespace(t *testing.T) {
	embedding := []float64{0.029445774853229523, -0.0034673467744141817, 0.007123, 1.234567890123456}
	payload, err := json.Marshal(embedding)
	if err != nil {
		t.Fatalf("marshal embedding: %v", err)
	}

	if bytes.Contains(payload, []byte(", ")) {
		t.Fatalf("embedding JSON must be compact: %s", string(payload))
	}

	var roundTrip []float64
	if err := json.Unmarshal(payload, &roundTrip); err != nil {
		t.Fatalf("unmarshal embedding: %v", err)
	}
	if len(roundTrip) != len(embedding) {
		t.Fatalf("roundtrip embedding length mismatch: got %d, want %d", len(roundTrip), len(embedding))
	}
}

func TestEmbeddingBlobDimensionValidationAgainstColumn(t *testing.T) {
	database := setupSchemaDB(t)

	embedding := []float64{0.1, 0.2, 0.3, 0.4}
	embeddingBytes, err := json.Marshal(embedding)
	if err != nil {
		t.Fatalf("marshal embedding: %v", err)
	}

	if _, err := database.Exec(`
		INSERT INTO code_chunks (
			id, file_path, symbol, sha, lang, embedding, embedding_provider, embedding_dimensions
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "chunk-dims", "src/search.go", "", "abcd1234", "go", embeddingBytes, "OpenAI", len(embedding)); err != nil {
		t.Fatalf("insert chunk row: %v", err)
	}

	var storedEmbedding []byte
	var storedDims int
	if err := database.QueryRow(`
		SELECT embedding, embedding_dimensions
		FROM code_chunks
		WHERE id = ?
	`, "chunk-dims").Scan(&storedEmbedding, &storedDims); err != nil {
		t.Fatalf("query chunk row: %v", err)
	}

	var parsed []float64
	if err := json.Unmarshal(storedEmbedding, &parsed); err != nil {
		t.Fatalf("unmarshal stored embedding: %v", err)
	}
	if len(parsed) != storedDims {
		t.Fatalf("embedding dimension mismatch: len=%d dimensions=%d", len(parsed), storedDims)
	}
}
