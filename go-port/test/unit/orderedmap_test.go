package unit

import (
	"strings"
	"testing"

	"github.com/alessandrojcm/pampax-go/internal/codemap"
)

func TestOrderedMapPreservesTopLevelInsertionOrder(t *testing.T) {
	mapData := codemap.NewOrderedMap()
	mapData.Set("z-chunk", codemap.ChunkMetadata{File: "src/z.js", SHA: "sha-z", Lang: "javascript"})
	mapData.Set("a-chunk", codemap.ChunkMetadata{File: "src/a.js", SHA: "sha-a", Lang: "javascript"})

	payload, err := codemap.MarshalCodemap(mapData)
	if err != nil {
		t.Fatalf("MarshalCodemap returned error: %v", err)
	}

	serialized := string(payload)
	indexZ := strings.Index(serialized, `"z-chunk"`)
	indexA := strings.Index(serialized, `"a-chunk"`)

	if indexZ == -1 || indexA == -1 {
		t.Fatalf("expected top-level keys in output, got: %s", serialized)
	}

	if indexZ > indexA {
		t.Fatalf("expected insertion order to be preserved, got: %s", serialized)
	}
}

func TestChunkMetadataFieldsAreAlphabeticallySortedInJSON(t *testing.T) {
	symbol := "handler"
	mapData := codemap.NewOrderedMap()
	mapData.Set("chunk", codemap.ChunkMetadata{
		ChunkType:         "function",
		Dimensions:        1536,
		Encrypted:         false,
		File:              "src/utils/logger.js",
		HasDocumentation:  false,
		HasIntent:         false,
		HasPampaTags:      true,
		Lang:              "javascript",
		PathWeight:        1,
		Provider:          "OpenAI",
		SHA:               "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SuccessRate:       0,
		Symbol:            &symbol,
		SymbolCalls:       []string{"extract"},
		SymbolCallTargets: nil,
		SymbolCallers:     nil,
		SymbolNeighbors:   nil,
		SymbolSignature:   "handler()",
		Synonyms:          nil,
		VariableCount:     0,
	})

	payload, err := codemap.MarshalCodemap(mapData)
	if err != nil {
		t.Fatalf("MarshalCodemap returned error: %v", err)
	}

	serialized := string(payload)
	orderedKeys := []string{
		`"chunkType"`,
		`"dimensions"`,
		`"encrypted"`,
		`"file"`,
		`"hasDocumentation"`,
		`"hasIntent"`,
		`"hasPampaTags"`,
		`"lang"`,
		`"path_weight"`,
		`"provider"`,
		`"sha"`,
		`"success_rate"`,
		`"symbol"`,
		`"symbol_call_targets"`,
		`"symbol_callers"`,
		`"symbol_calls"`,
		`"symbol_neighbors"`,
		`"symbol_signature"`,
		`"synonyms"`,
		`"variableCount"`,
	}

	lastIndex := -1
	for _, key := range orderedKeys {
		idx := strings.Index(serialized, key)
		if idx == -1 {
			t.Fatalf("expected key %s in serialized metadata: %s", key, serialized)
		}
		if idx < lastIndex {
			t.Fatalf("expected keys in alphabetical order, got: %s", serialized)
		}
		lastIndex = idx
	}
}

func TestCodemapFormattingIsTwoSpaceUnixNewlineWithFinalNewline(t *testing.T) {
	mapData := codemap.NewOrderedMap()
	mapData.Set("chunk", codemap.ChunkMetadata{File: "src/main.js", SHA: "sha", Lang: "javascript"})

	payload, err := codemap.MarshalCodemap(mapData)
	if err != nil {
		t.Fatalf("MarshalCodemap returned error: %v", err)
	}

	serialized := string(payload)

	if strings.Contains(serialized, "\r\n") {
		t.Fatalf("expected Unix newlines only, got: %q", serialized)
	}

	if !strings.HasSuffix(serialized, "\n") {
		t.Fatalf("expected final newline, got: %q", serialized)
	}

	if !strings.Contains(serialized, "\n  \"chunk\": {") {
		t.Fatalf("expected 2-space indentation, got: %q", serialized)
	}
}

func TestSymbolNullAndSymbolParametersOmittedWhenEmpty(t *testing.T) {
	mapData := codemap.NewOrderedMap()
	mapData.Set("chunk", codemap.ChunkMetadata{
		File:             "src/utils/logger.js",
		SHA:              "sha",
		Lang:             "javascript",
		Symbol:           nil,
		SymbolParameters: []string{},
	})

	payload, err := codemap.MarshalCodemap(mapData)
	if err != nil {
		t.Fatalf("MarshalCodemap returned error: %v", err)
	}

	serialized := string(payload)

	if !strings.Contains(serialized, `"symbol": null`) {
		t.Fatalf("expected symbol to be null, got: %s", serialized)
	}

	if strings.Contains(serialized, `"symbol_parameters"`) {
		t.Fatalf("expected symbol_parameters to be omitted when empty, got: %s", serialized)
	}
}

func TestPathsAreNormalizedToForwardSlashes(t *testing.T) {
	mapData := codemap.NewOrderedMap()
	mapData.Set("chunk", codemap.ChunkMetadata{
		File: "src\\utils\\logger.js",
		SHA:  "sha",
		Lang: "javascript",
	})

	payload, err := codemap.MarshalCodemap(mapData)
	if err != nil {
		t.Fatalf("MarshalCodemap returned error: %v", err)
	}

	serialized := string(payload)
	if !strings.Contains(serialized, `"file": "src/utils/logger.js"`) {
		t.Fatalf("expected normalized forward-slash path, got: %s", serialized)
	}
}

func TestRequiredArraysAlwaysPresent(t *testing.T) {
	mapData := codemap.NewOrderedMap()
	mapData.Set("chunk", codemap.ChunkMetadata{File: "src/main.js", SHA: "sha", Lang: "javascript"})

	payload, err := codemap.MarshalCodemap(mapData)
	if err != nil {
		t.Fatalf("MarshalCodemap returned error: %v", err)
	}

	serialized := string(payload)
	required := []string{
		`"synonyms": []`,
		`"symbol_calls": []`,
		`"symbol_call_targets": []`,
		`"symbol_callers": []`,
		`"symbol_neighbors": []`,
	}

	for _, key := range required {
		if !strings.Contains(serialized, key) {
			t.Fatalf("expected required empty array field %s, got: %s", key, serialized)
		}
	}
}
