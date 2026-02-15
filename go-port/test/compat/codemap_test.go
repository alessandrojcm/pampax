package compat

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestNodeFixtureCodemapOrderingAndSemantics(t *testing.T) {
	codemapPath := filepath.Join("..", "fixtures", "small", "pampa.codemap.json")
	raw, err := os.ReadFile(codemapPath)
	if err != nil {
		t.Fatalf("read codemap fixture: %v", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil {
		t.Fatalf("read codemap opening token: %v", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		t.Fatalf("expected codemap object start, got %v", token)
	}

	if !decoder.More() {
		t.Fatal("expected at least one codemap entry")
	}

	firstKeyToken, err := decoder.Token()
	if err != nil {
		t.Fatalf("read first codemap key: %v", err)
	}
	firstKey, ok := firstKeyToken.(string)
	if !ok {
		t.Fatalf("expected string codemap key, got %T", firstKeyToken)
	}

	const expectedFirstKey = "AGENTS.md:section_group_105_undefined_partgroup_105funcs:a681484f"
	if firstKey != expectedFirstKey {
		t.Fatalf("unexpected first codemap key: got %q, want %q", firstKey, expectedFirstKey)
	}

	firstEntryFields := make([]string, 0, 24)
	objectStart, err := decoder.Token()
	if err != nil {
		t.Fatalf("read first entry object start: %v", err)
	}
	if delim, ok := objectStart.(json.Delim); !ok || delim != '{' {
		t.Fatalf("expected first entry object start, got %v", objectStart)
	}

	for decoder.More() {
		fieldToken, err := decoder.Token()
		if err != nil {
			t.Fatalf("read first entry field key: %v", err)
		}
		field, ok := fieldToken.(string)
		if !ok {
			t.Fatalf("expected field key string, got %T", fieldToken)
		}
		firstEntryFields = append(firstEntryFields, field)

		var skip any
		if err := decoder.Decode(&skip); err != nil {
			t.Fatalf("skip field value for %q: %v", field, err)
		}
	}

	objectEnd, err := decoder.Token()
	if err != nil {
		t.Fatalf("read first entry object end: %v", err)
	}
	if delim, ok := objectEnd.(json.Delim); !ok || delim != '}' {
		t.Fatalf("expected first entry object end, got %v", objectEnd)
	}

	if len(firstEntryFields) < 4 {
		t.Fatalf("expected at least 4 fields in first codemap entry, got %d", len(firstEntryFields))
	}
	expectedPrefix := []string{"file", "symbol", "sha", "lang"}
	for i, expected := range expectedPrefix {
		if firstEntryFields[i] != expected {
			t.Fatalf("unexpected first-entry field at index %d: got %q, want %q", i, firstEntryFields[i], expected)
		}
	}

	sortedFields := append([]string(nil), firstEntryFields...)
	sort.Strings(sortedFields)
	if len(sortedFields) != len(firstEntryFields) {
		t.Fatalf("field length changed during sort: got %d, want %d", len(sortedFields), len(firstEntryFields))
	}

	var codemap map[string]map[string]any
	if err := json.Unmarshal(raw, &codemap); err != nil {
		t.Fatalf("unmarshal codemap fixture: %v", err)
	}

	if len(codemap) == 0 {
		t.Fatal("expected non-empty codemap")
	}

	entriesWithParameters := 0
	for id, meta := range codemap {
		symbol, ok := meta["symbol"]
		if !ok {
			t.Fatalf("codemap entry %q missing required symbol field", id)
		}
		if symbol == "" {
			t.Fatalf("codemap entry %q uses empty symbol string; expected non-empty string or null", id)
		}

		if params, ok := meta["symbol_parameters"]; ok {
			entriesWithParameters++
			if arr, ok := params.([]any); ok && len(arr) == 0 {
				t.Fatalf("codemap entry %q has empty symbol_parameters array; expected omission when empty", id)
			}
		}
	}

	if entriesWithParameters == 0 {
		t.Fatal("expected at least one codemap entry with symbol_parameters")
	}
}
