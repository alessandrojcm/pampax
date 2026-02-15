package codemap

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

type ChunkMetadata struct {
	File              string
	Symbol            *string
	SHA               string
	Lang              string
	ChunkType         string
	Provider          string
	Dimensions        int
	HasPampaTags      bool
	HasIntent         bool
	HasDocumentation  bool
	VariableCount     int
	Synonyms          []string
	PathWeight        float64
	LastUsedAt        string
	SuccessRate       float64
	Encrypted         bool
	SymbolSignature   string
	SymbolParameters  []string
	SymbolReturn      string
	SymbolCalls       []string
	SymbolCallTargets []string
	SymbolCallers     []string
	SymbolNeighbors   []string
}

func NormalizeChunkMetadata(input ChunkMetadata) ChunkMetadata {
	out := input
	out.File = normalizePathForStorage(out.File)
	out.Symbol = normalizeSymbol(out.Symbol)
	out.Synonyms = sanitizeStringArray(out.Synonyms)
	out.SymbolCalls = sanitizeStringArray(out.SymbolCalls)
	out.SymbolCallTargets = sanitizeStringArray(out.SymbolCallTargets)
	out.SymbolCallers = sanitizeStringArray(out.SymbolCallers)
	out.SymbolNeighbors = sanitizeStringArray(out.SymbolNeighbors)

	params := sanitizeStringArray(out.SymbolParameters)
	if len(params) > 0 {
		out.SymbolParameters = params
	} else {
		out.SymbolParameters = nil
	}

	out.SymbolSignature = strings.TrimSpace(out.SymbolSignature)
	out.SymbolReturn = strings.TrimSpace(out.SymbolReturn)

	if out.VariableCount < 0 {
		out.VariableCount = 0
	}

	if out.PathWeight < 0 {
		out.PathWeight = 0
	}

	if out.PathWeight == 0 {
		out.PathWeight = 1
	}

	if out.SuccessRate < 0 {
		out.SuccessRate = 0
	}

	if out.SuccessRate > 1 {
		out.SuccessRate = 1
	}

	return out
}

func (m ChunkMetadata) MarshalJSON() ([]byte, error) {
	normalized := NormalizeChunkMetadata(m)

	payload := map[string]any{
		"file":                normalized.File,
		"symbol":              normalized.Symbol,
		"sha":                 normalized.SHA,
		"lang":                normalized.Lang,
		"hasPampaTags":        normalized.HasPampaTags,
		"hasIntent":           normalized.HasIntent,
		"hasDocumentation":    normalized.HasDocumentation,
		"variableCount":       normalized.VariableCount,
		"synonyms":            normalized.Synonyms,
		"path_weight":         normalized.PathWeight,
		"success_rate":        normalized.SuccessRate,
		"encrypted":           normalized.Encrypted,
		"symbol_calls":        normalized.SymbolCalls,
		"symbol_call_targets": normalized.SymbolCallTargets,
		"symbol_callers":      normalized.SymbolCallers,
		"symbol_neighbors":    normalized.SymbolNeighbors,
	}

	if normalized.ChunkType != "" {
		payload["chunkType"] = normalized.ChunkType
	}

	if normalized.Provider != "" {
		payload["provider"] = normalized.Provider
	}

	if normalized.Dimensions > 0 {
		payload["dimensions"] = normalized.Dimensions
	}

	if normalized.LastUsedAt != "" {
		payload["last_used_at"] = normalized.LastUsedAt
	}

	if normalized.SymbolSignature != "" {
		payload["symbol_signature"] = normalized.SymbolSignature
	}

	if len(normalized.SymbolParameters) > 0 {
		payload["symbol_parameters"] = normalized.SymbolParameters
	}

	if normalized.SymbolReturn != "" {
		payload["symbol_return"] = normalized.SymbolReturn
	}

	return json.Marshal(payload)
}

func normalizePathForStorage(path string) string {
	normalized := strings.ReplaceAll(path, "\\", "/")
	normalized = filepath.ToSlash(normalized)
	normalized = strings.TrimPrefix(normalized, "./")
	return normalized
}

func normalizeSymbol(symbol *string) *string {
	if symbol == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*symbol)
	if trimmed == "" {
		return nil
	}

	normalized := trimmed
	return &normalized
}

func sanitizeStringArray(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))

	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}

		if _, exists := seen[trimmed]; exists {
			continue
		}

		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}

	if len(out) == 0 {
		return []string{}
	}

	return out
}
