# Codemap JSON Schema Specification

**Version:** 1.0  
**Created:** 2026-01-28  
**Status:** Reference Specification for Go Port Compatibility

---

## Overview

This document specifies the exact JSON schema for `pampa.codemap.json`, which serves as a fast-access metadata index for all chunks in the project. The Go implementation MUST produce semantically identical JSON structures.

---

## File Location

**Path:** `pampa.codemap.json` (in project root)

**Purpose:**
- Fast metadata lookup without querying SQLite
- Human-readable chunk inventory
- Symbol relationship graph
- Usage tracking and analytics

---

## JSON Structure

### Top-Level Format

**Type:** JSON object (dictionary/map)

**Keys:** Chunk IDs (unique identifiers)

**Values:** Chunk metadata objects

```json
{
  "{chunk_id_1}": { /* metadata */ },
  "{chunk_id_2}": { /* metadata */ },
  ...
}
```

### Chunk ID Format

**Pattern:** `{file_path}:{symbol}:{sha_prefix}`

**Components:**
- `file_path`: Project-relative path (forward slashes)
- `symbol`: Function/class/method name (or descriptive identifier)
- `sha_prefix`: First 8 characters of SHA-1 hash

**Examples:**
```
src/service.js:indexProject:a681484f
src/cli/commands/search.js:buildScopeFiltersFromOptions:0e671987
AGENTS.md:section_group_105_undefined_partgroup_105funcs:a681484f
```

**Special Cases:**
- **No symbol:** Use generated identifier (e.g., `section_group_105_undefined_partgroup_105funcs`)
- **Grouped chunks:** Prefix with `group_N` (e.g., `group_3funcs`)
- **Statement chunks:** Use type identifier (e.g., `assignment`, `if_statement`)

---

## Chunk Metadata Schema

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | string | ✅ | Project-relative file path |
| `symbol` | string \| null | ✅ | Symbol name or null |
| `sha` | string | ✅ | Full SHA-1 hash (40 hex chars) |
| `lang` | string | ✅ | Language identifier |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `chunkType` | string | undefined | Chunk classification |
| `provider` | string | undefined | Embedding provider name |
| `dimensions` | number | undefined | Vector dimensions |
| `hasPampaTags` | boolean | false | Has semantic tags |
| `hasIntent` | boolean | false | Has intent description |
| `hasDocumentation` | boolean | false | Has doc comments |
| `variableCount` | number | 0 | Number of important variables |
| `synonyms` | string[] | [] | Alternative names |
| `path_weight` | number | 1 | Search relevance weight |
| `success_rate` | number | 0 | Usage success rate (0-1) |
| `encrypted` | boolean | false | Chunk file is encrypted |
| `symbol_signature` | string | undefined | Function signature |
| `symbol_parameters` | string[] | undefined | Parameter names |
| `symbol_return` | string | undefined | Return type/value |
| `symbol_calls` | string[] | [] | Functions called |
| `symbol_call_targets` | string[] | [] | Call target symbols |
| `symbol_callers` | string[] | [] | Symbols that call this |
| `symbol_neighbors` | string[] | [] | Related symbols |
| `last_used_at` | string | undefined | ISO 8601 timestamp |

### Field Ordering

**Observed order** (from Node.js implementation):

```json
{
  "chunkType": "function",
  "dimensions": 1536,
  "encrypted": false,
  "file": "src/service.js",
  "hasDocumentation": false,
  "hasIntent": false,
  "hasPampaTags": true,
  "lang": "javascript",
  "path_weight": 1,
  "provider": "OpenAI",
  "sha": "a681484fe8efce38ee22f53ef39c057bb3198732",
  "success_rate": 0,
  "symbol": "indexProject",
  "symbol_call_targets": [],
  "symbol_callers": [],
  "symbol_calls": ["extractSymbolMetadata"],
  "symbol_neighbors": [],
  "symbol_parameters": ["options", "provider"],
  "symbol_return": "result",
  "symbol_signature": "indexProject(options, provider) : result",
  "synonyms": [],
  "variableCount": 0
}
```

**Ordering Rules:**

⚠️ **IMPORTANT: Actual Node.js behavior differs from original specification**

1. **Top-level chunk IDs:** Node.js does NOT alphabetically sort (insertion order preserved)
2. **Field keys within chunks:** Alphabetically sorted (via `JSON.stringify()`)
3. **Go compatibility:** MUST match Node.js behavior (unsorted top-level keys)

**Rationale:**
- Node.js uses standard `JSON.stringify(obj, null, 2)` which preserves insertion order for object keys (ES2015+)
- Top-level chunk IDs are NOT explicitly sorted before serialization
- Field keys within each chunk ARE alphabetically sorted by `JSON.stringify()`

**Go Implementation Requirement:**
- Preserve insertion order for top-level chunk IDs (use ordered map or write in insertion order)
- Field keys within chunks will be alphabetically sorted by `json.Marshal()` (matches Node.js)
- Do NOT sort top-level keys alphabetically (breaks compatibility)

---

## Field Specifications

### 1. `file` (string, required)

**Format:** Project-relative path with forward slashes

**Examples:**
```json
"file": "src/service.js"
"file": "src/cli/commands/search.js"
"file": "AGENTS.md"
```

**Rules:**
- Always forward slash (`/`) separator (even on Windows)
- Relative to project root
- No leading `./` or `/`
- UTF-8 encoded (supports non-ASCII filenames)

### 2. `symbol` (string | null, required)

**Format:** Symbol name or null

**Examples:**
```json
"symbol": "indexProject"
"symbol": "ClassName::methodName"
"symbol": "section_group_105_undefined_partgroup_105funcs"
"symbol": null
```

**Rules:**
- `null` if no named symbol (NOT empty string `""`)
- May include scope (e.g., `ClassName::method`)
- Generated identifiers for markdown sections, statement groups, etc.

### 3. `sha` (string, required)

**Format:** Full SHA-1 hash (40 hex characters, lowercase)

**Example:**
```json
"sha": "a681484fe8efce38ee22f53ef39c057bb3198732"
```

**Rules:**
- Exactly 40 characters
- Lowercase hexadecimal
- Matches chunk filename (without `.gz` extension)

### 4. `lang` (string, required)

**Format:** Language identifier (lowercase)

**Examples:**
```json
"lang": "javascript"
"lang": "typescript"
"lang": "python"
"lang": "markdown"
```

**Supported Values:**
```
bash, c, csharp, cpp, css, elixir, go, haskell, html, java,
javascript, json, kotlin, lua, markdown, ocaml, php, python,
ruby, rust, scala, swift, tsx, typescript
```

### 5. `chunkType` (string, optional)

**Format:** Chunk classification

**Examples:**
```json
"chunkType": "function"
"chunkType": "class"
"chunkType": "method"
"chunkType": "statement"
```

**Common Values:**
- `function` (default)
- `class`
- `method`
- `statement`
- `group_{N}funcs` (combined chunks)

### 6. `provider` (string, optional)

**Format:** Embedding provider name

**Examples:**
```json
"provider": "OpenAI"
"provider": "local"
"provider": "Ollama"
```

**Values:**
- `OpenAI`: OpenAI API
- `local`: Transformers.js (local)
- `Ollama`: Ollama local models

### 7. `dimensions` (number, optional)

**Format:** Integer vector dimensions

**Examples:**
```json
"dimensions": 1536
"dimensions": 384
"dimensions": 3072
```

**Common Values:**
- `384`: local Transformers.js
- `1536`: OpenAI text-embedding-3-small
- `3072`: OpenAI text-embedding-3-large
- `1024`: Ollama mxbai-embed-large

### 8. Boolean Flags

**Format:** `true` or `false` (not `0`/`1` or `"true"`/`"false"`)

**Examples:**
```json
"hasPampaTags": true
"hasIntent": false
"hasDocumentation": false
"encrypted": false
```

**Defaults:**
- All boolean fields default to `false`
- Only included if `true` or explicitly set

### 9. `variableCount` (number, optional)

**Format:** Non-negative integer

**Example:**
```json
"variableCount": 0
"variableCount": 5
```

**Meaning:** Number of important variables detected in chunk

### 10. `synonyms` (string[], optional)

**Format:** Array of alternative names

**Examples:**
```json
"synonyms": []
"synonyms": ["createIndex", "buildIndex"]
```

**Rules:**
- Always an array (never null)
- May be empty `[]`
- Deduplicated
- Trimmed strings

### 11. `path_weight` (number, optional)

**Format:** Positive float (search relevance multiplier)

**Example:**
```json
"path_weight": 1
"path_weight": 1.5
```

**Default:** `1`  
**Purpose:** Boost search relevance for frequently accessed chunks

### 12. `success_rate` (number, optional)

**Format:** Float between 0 and 1

**Example:**
```json
"success_rate": 0
"success_rate": 0.75
```

**Default:** `0`  
**Purpose:** Track query success rate for ranking

### 13. `encrypted` (boolean, optional)

**Format:** Boolean

**Example:**
```json
"encrypted": false
"encrypted": true
```

**Meaning:** Chunk file uses `.gz.enc` format (encrypted)

### 14. Symbol Metadata

#### `symbol_signature` (string, optional)

**Format:** Function/method signature

**Examples:**
```json
"symbol_signature": "indexProject(options, provider) : result"
"symbol_signature": "buildScopeFiltersFromOptions(options, projectPath, sessionPack) : options"
"symbol_signature": "cosineSimilarity(a, b) : number"
```

**Pattern:** `{name}({params}) : {return}`

#### `symbol_parameters` (string[], optional)

**Format:** Array of parameter names

**Examples:**
```json
"symbol_parameters": ["options", "provider"]
"symbol_parameters": ["a", "b"]
```

**Rules:**
- Only present if function has parameters
- Empty array or undefined if no parameters
- Deduplicated

#### `symbol_return` (string, optional)

**Format:** Return type or value description

**Examples:**
```json
"symbol_return": "result"
"symbol_return": "number"
"symbol_return": "options"
```

#### `symbol_calls` (string[], optional)

**Format:** Array of function names called by this symbol

**Examples:**
```json
"symbol_calls": ["extractSymbolMetadata", "computeFastHash"]
"symbol_calls": []
```

**Purpose:** Call graph analysis

#### `symbol_call_targets` (string[], optional)

**Format:** Array of target symbol identifiers

**Examples:**
```json
"symbol_call_targets": []
```

**Note:** Currently unused but reserved for future

#### `symbol_callers` (string[], optional)

**Format:** Array of symbols that call this function

**Examples:**
```json
"symbol_callers": []
```

**Purpose:** Reverse call graph

#### `symbol_neighbors` (string[], optional)

**Format:** Array of related symbols (defined nearby)

**Examples:**
```json
"symbol_neighbors": []
```

**Purpose:** Symbol proximity relationships

### 15. `last_used_at` (string, optional)

**Format:** ISO 8601 timestamp (UTC)

**Example:**
```json
"last_used_at": "2026-01-28T14:30:00.000Z"
```

**Rules:**
- ISO 8601 format with milliseconds
- UTC timezone (suffix `Z`)
- Only present if chunk has been accessed

---

## Normalization Rules

### Reference Implementation

**Location:** `src/codemap/types.js:155-242`

### Key Normalization Behaviors

1. **Top-level sorting:**
   - ⚠️ **ACTUAL BEHAVIOR:** Chunk IDs are **NOT sorted** (insertion order preserved)
   - **Original spec claimed:** Alphabetically sorted
   - **Impact:** Go implementation must preserve insertion order for compatibility
   - **Rationale:** Node.js doesn't explicitly sort before `JSON.stringify()`

2. **Field presence:**
   - Omit optional fields if default value
   - Example: `"symbol_parameters"` omitted if empty

3. **Empty arrays:**
   - `synonyms`: Always present (may be `[]`)
   - `symbol_calls`: Always present (may be `[]`)
   - `symbol_callers`: Always present (may be `[]`)
   - `symbol_neighbors`: Always present (may be `[]`)
   - `symbol_call_targets`: Always present (may be `[]`)

4. **Null vs undefined:**
   - `symbol`: `null` if no symbol (NOT omitted)
   - `symbol_parameters`: **omitted** if empty (NOT `[]`)
   - `symbol_return`: **omitted** if no return value

5. **String sanitization:**
   - Trim whitespace
   - Remove empty strings
   - Deduplicate arrays

---

## Serialization Format

### JSON Encoding

**Node.js Reference:** `src/codemap/io.js:26-35`

```javascript
fs.writeFileSync(resolvedPath, JSON.stringify(normalized, null, 2));
```

**Format:**
- **Indentation:** 2 spaces
- **No trailing comma**
- **UTF-8 encoding**
- **Unix line endings** (`\n`)

**Go Implementation:**

⚠️ **Important:** Do NOT sort top-level chunk IDs for compatibility with Node.js

```go
import (
    "encoding/json"
    "os"
)

func writeCodemap(path string, codemap map[string]ChunkMetadata) error {
    // DO NOT sort keys - preserve insertion order for Node.js compatibility
    // Use ordered map (e.g., github.com/iancoleman/orderedmap) or write in insertion order
    
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()
    
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")  // 2-space indent
    
    // json.Marshal() will alphabetically sort field keys within each chunk (correct)
    // but we must preserve insertion order for top-level chunk IDs
    return encoder.Encode(codemap)
}
```

**Alternative for ordered output:**
```go
// Use ordered map library to preserve insertion order
import "github.com/iancoleman/orderedmap"

func buildCodemap() *orderedmap.OrderedMap {
    codemap := orderedmap.New()
    
    // Add chunks in the order they're discovered during indexing
    for _, chunk := range discoveredChunks {
        codemap.Set(chunk.ID, chunk.Metadata)
    }
    
    return codemap
}
```

### Whitespace

**Required:**
- 2-space indentation
- No tabs
- Unix line endings (`\n`)
- Final newline at end of file

**Example:**
```json
{
  "src/service.js:indexProject:a681484f": {
    "file": "src/service.js",
    "symbol": "indexProject",
    "sha": "a681484fe8efce38ee22f53ef39c057bb3198732"
  }
}

```
(Note: final newline after closing `}`)

---

## Example Entries

### 1. Simple Function Chunk

```json
"src/service.js:cosineSimilarity:b234c567": {
  "chunkType": "function",
  "dimensions": 1536,
  "encrypted": false,
  "file": "src/service.js",
  "hasDocumentation": false,
  "hasIntent": false,
  "hasPampaTags": true,
  "lang": "javascript",
  "path_weight": 1,
  "provider": "OpenAI",
  "sha": "b234c567890abcdef1234567890abcdef1234567",
  "success_rate": 0,
  "symbol": "cosineSimilarity",
  "symbol_calls": [],
  "symbol_call_targets": [],
  "symbol_callers": ["searchCode"],
  "symbol_neighbors": [],
  "symbol_parameters": ["a", "b"],
  "symbol_return": "number",
  "symbol_signature": "cosineSimilarity(a, b) : number",
  "synonyms": [],
  "variableCount": 0
}
```

### 2. Markdown Section

```json
"AGENTS.md:section_group_105_undefined_partgroup_105funcs:a681484f": {
  "chunkType": "function",
  "dimensions": 1536,
  "encrypted": false,
  "file": "AGENTS.md",
  "hasDocumentation": false,
  "hasIntent": false,
  "hasPampaTags": true,
  "lang": "markdown",
  "path_weight": 1,
  "provider": "OpenAI",
  "sha": "a681484fe8efce38ee22f53ef39c057bb3198732",
  "success_rate": 0,
  "symbol": "section_group_105_undefined_partgroup_105funcs",
  "symbol_calls": [],
  "symbol_call_targets": [],
  "symbol_callers": [],
  "symbol_neighbors": [],
  "symbol_signature": "section_group_105_undefined_partgroup_105funcs()",
  "synonyms": [],
  "variableCount": 0
}
```

### 3. Function with Parameters

```json
"src/cli/commands/search.js:buildScopeFiltersFromOptions:0e671987": {
  "chunkType": "function",
  "dimensions": 1536,
  "encrypted": false,
  "file": "src/cli/commands/search.js",
  "hasDocumentation": false,
  "hasIntent": false,
  "hasPampaTags": true,
  "lang": "javascript",
  "path_weight": 1,
  "provider": "OpenAI",
  "sha": "0e6719877b4b323fa578f1035b50a4947700e13e",
  "success_rate": 0,
  "symbol": "buildScopeFiltersFromOptions",
  "symbol_calls": ["resolveScopeWithPack"],
  "symbol_call_targets": [],
  "symbol_callers": [],
  "symbol_neighbors": [],
  "symbol_parameters": ["options", "projectPath", "sessionPack"],
  "symbol_return": "options",
  "symbol_signature": "buildScopeFiltersFromOptions(options, projectPath, sessionPack) : options",
  "synonyms": [],
  "variableCount": 0
}
```

---

## Edge Cases

### 1. Null Symbol

**Scenario:** No named symbol (e.g., top-level statement)

```json
{
  "file": "src/config.js",
  "symbol": null,
  "sha": "abc123..."
}
```

**Rule:** Use `null`, NOT `""` or omit field

### 2. Empty Arrays

**Scenario:** No symbol calls

**Correct:**
```json
{
  "symbol_calls": []
}
```

**Incorrect:**
```json
{
  "symbol_calls": null  // ❌ Use [] instead
}
```

### 3. Omitted Optional Fields

**Scenario:** No parameters

**Correct:**
```json
{
  "symbol_signature": "foo()"
  // symbol_parameters omitted
}
```

**Incorrect:**
```json
{
  "symbol_parameters": []  // ❌ Should be omitted
}
```

### 4. Windows Paths

**Scenario:** Indexed on Windows

**Stored:**
```json
{
  "file": "src/utils/logger.js"  // ✅ Forward slashes
}
```

**Never:**
```json
{
  "file": "src\\utils\\logger.js"  // ❌ Backslashes
}
```

### 5. Non-ASCII Filenames

**Scenario:** UTF-8 filename

```json
{
  "file": "src/测试/test.js"  // ✅ UTF-8 preserved
}
```

---

## Validation

### JSON Schema (TypeScript/Zod)

**Reference:** `src/codemap/types.js:32-56`

```typescript
const CodemapChunkSchema = z.object({
    file: z.string(),
    symbol: z.union([z.string(), z.null()]).optional(),
    sha: z.string(),
    lang: z.string().optional(),
    chunkType: z.string().optional(),
    provider: z.string().optional(),
    dimensions: z.number().optional(),
    hasPampaTags: z.boolean().optional(),
    hasIntent: z.boolean().optional(),
    hasDocumentation: z.boolean().optional(),
    variableCount: z.number().optional(),
    synonyms: z.array(z.string()).optional(),
    path_weight: z.number().optional(),
    last_used_at: z.union([z.string(), z.null()]).optional(),
    success_rate: z.number().optional(),
    encrypted: z.boolean().optional(),
    symbol_signature: z.string().optional(),
    symbol_parameters: z.array(z.string()).optional(),
    symbol_return: z.string().optional(),
    symbol_calls: z.array(z.string()).optional(),
    symbol_call_targets: z.array(z.string()).optional(),
    symbol_callers: z.array(z.string()).optional(),
    symbol_neighbors: z.array(z.string()).optional()
}).passthrough();

const CodemapSchema = z.record(CodemapChunkSchema);
```

### Go Validation

```go
type ChunkMetadata struct {
    File              string   `json:"file"`
    Symbol            *string  `json:"symbol"`
    SHA               string   `json:"sha"`
    Lang              string   `json:"lang,omitempty"`
    ChunkType         string   `json:"chunkType,omitempty"`
    Provider          string   `json:"provider,omitempty"`
    Dimensions        int      `json:"dimensions,omitempty"`
    HasPampaTags      bool     `json:"hasPampaTags,omitempty"`
    HasIntent         bool     `json:"hasIntent,omitempty"`
    HasDocumentation  bool     `json:"hasDocumentation,omitempty"`
    VariableCount     int      `json:"variableCount,omitempty"`
    Synonyms          []string `json:"synonyms,omitempty"`
    PathWeight        float64  `json:"path_weight,omitempty"`
    SuccessRate       float64  `json:"success_rate,omitempty"`
    Encrypted         bool     `json:"encrypted,omitempty"`
    SymbolSignature   string   `json:"symbol_signature,omitempty"`
    SymbolParameters  []string `json:"symbol_parameters,omitempty"`
    SymbolReturn      string   `json:"symbol_return,omitempty"`
    SymbolCalls       []string `json:"symbol_calls,omitempty"`
    SymbolCallTargets []string `json:"symbol_call_targets,omitempty"`
    SymbolCallers     []string `json:"symbol_callers,omitempty"`
    SymbolNeighbors   []string `json:"symbol_neighbors,omitempty"`
    LastUsedAt        string   `json:"last_used_at,omitempty"`
}
```

---

## Compatibility Checklist

- [ ] ⚠️ **Top-level chunk IDs preserve insertion order** (NOT alphabetically sorted)
- [ ] 2-space JSON indentation
- [ ] Unix line endings (`\n`)
- [ ] Final newline after closing `}`
- [ ] `symbol` field is `null` for no symbol (not omitted)
- [ ] `symbol_parameters` omitted if empty (not `[]`)
- [ ] Forward slashes in `file` paths (all platforms)
- [ ] Boolean fields as `true`/`false` (not strings)
- [ ] ISO 8601 timestamps with `Z` suffix
- [ ] UTF-8 encoding for non-ASCII filenames
- [ ] Alphabetically sorted field keys within each chunk metadata object
- [ ] Arrays always present (not null) for `synonyms`, `symbol_calls`, etc.

---

## Known Compatibility Issues

### 1. Top-Level Key Ordering ⚠️

**Issue:** Original specification claimed alphabetical sorting, but Node.js implementation does NOT sort.

**Actual Behavior:**
- Top-level chunk IDs follow **insertion order** (order chunks were added during indexing)
- This is the default behavior of `JSON.stringify()` in ES2015+ JavaScript

**Impact on Go Implementation:**
- MUST preserve insertion order (use ordered map structure)
- DO NOT alphabetically sort top-level keys
- Field keys within each chunk object ARE alphabetically sorted (correct)

**Discovered:** Stage 0.2 fixture validation (2026-01-28)

**Resolution:** Accept unsorted behavior as canonical; update spec to match implementation

**Example:**
```json
{
  "DEMO_MULTI_PROJECT_EN.md:...": { ... },
  "DEMO_MULTI_PROJECT.md:...": { ... },     // Later alphabetically but inserted first
  "MIGRATION_GUIDE_v1.12.md:...": { ... }   // Would come before DEMO_* if sorted
}
```

Note: Chunk IDs appear in file discovery/indexing order, not alphabetical order.

---

## Reference Implementation Locations

**Node.js:**
- Schema definition: `src/codemap/types.js:32-56`
- Normalization: `src/codemap/types.js:155-242`
- Read/Write: `src/codemap/io.js`
- Metadata creation: `src/service.js:1631-1667`

**Key Functions:**
- `normalizeChunkMetadata()`: Field sanitization
- `normalizeCodemapRecord()`: Top-level normalization
- `writeCodemap()`: JSON serialization

---

## Summary

The codemap JSON format requires:
1. ✅ Top-level chunk IDs preserved in insertion order (not alphabetically sorted)
2. ✅ 2-space indented JSON
3. ✅ Unix line endings
4. ✅ Null for missing symbols (not omitted)
5. ✅ Omit empty optional arrays (except `synonyms`, `symbol_calls`, etc.)
6. ✅ Forward slashes in paths
7. ✅ UTF-8 encoding
8. ✅ ISO 8601 timestamps
9. ✅ Alphabetically sorted object keys within each chunk

**Critical for Compatibility:**
- Alphabetical ordering (both top-level and object keys)
- Null vs omitted semantics
- Array presence rules
- Path separator normalization
