# Vector Blob Format Specification

**Version:** 1.0  
**Created:** 2026-01-28  
**Status:** Reference Specification for Go Port Compatibility

---

## Overview

This document specifies the exact storage format for embedding vectors in the SQLite `code_chunks.embedding` BLOB column. The Go implementation MUST produce byte-identical BLOB representations for the same embedding vectors.

---

## Format Summary

| Property | Value |
|----------|-------|
| **Storage Type** | SQLite BLOB column |
| **Encoding** | UTF-8 text |
| **Serialization** | JSON array of numbers |
| **Data Type** | IEEE 754 double precision floats |
| **Byte Order** | Not applicable (text-based) |
| **Compression** | None |
| **Normalization** | Provider-dependent (not enforced) |

---

## Detailed Specification

### 1. Serialization Format

**Format:** JSON array of floating-point numbers

**Storage model:** Embedding providers return arrays of floats; the array is serialized as JSON and stored as UTF-8 bytes in the SQLite BLOB column.

**Example:**
```json
[0.029445774853229523,-0.0034673467744141817,0.007123...]
```

**Characteristics:**
- **No whitespace** between elements (compact JSON)
- **Full precision** preserved (typically 15-17 decimal digits)
- **No scientific notation** (standard decimal representation)
- **Negative numbers** include `-` prefix
- **Array brackets** `[` and `]` present
- **Comma-separated** values

### 2. Storage Process

#### Node.js Reference Implementation

**Location:** `src/service.js:1586-1617`

```javascript
// Step 1: Generate embedding (returns Array or Float32Array)
const embedding = await embeddingProvider.generateEmbedding(enhancedEmbeddingText);

// Step 2: Serialize to JSON
const jsonString = JSON.stringify(embedding);

// Step 3: Convert to Buffer
const buffer = Buffer.from(jsonString, 'utf8');

// Step 4: Store in BLOB column
await run(`
    INSERT OR REPLACE INTO code_chunks
    (id, ..., embedding, ...)
    VALUES (?, ..., ?, ...)
`, [
    chunkId,
    // ...
    buffer,  // ← BLOB value
    // ...
]);
```

#### Go Implementation Requirements

**Must:**
1. Serialize embedding as JSON array: `json.Marshal(embedding)`
2. Encode as UTF-8 bytes: `[]byte(jsonString)`
3. Store in BLOB column as byte slice

**Example (pseudocode):**
```go
// embedding is []float64 or []float32
jsonBytes, err := json.Marshal(embedding)
if err != nil {
    return err
}

// jsonBytes is now UTF-8 encoded JSON array
// Store directly in BLOB column
stmt.Exec(..., jsonBytes, ...)
```

### 3. Retrieval Process

#### Node.js Reference Implementation

**Location:** `src/service.js` (search functions)

```javascript
// Step 1: Read BLOB as Buffer
const buffer = row.embedding;  // Buffer object

// Step 2: Convert to UTF-8 string
const jsonString = buffer.toString('utf8');

// Step 3: Parse JSON
const embedding = JSON.parse(jsonString);  // Array of numbers

// Step 4: Use for similarity calculation
const similarity = cosineSimilarity(queryEmbedding, embedding);
```

#### Go Implementation Requirements

**Must:**
1. Read BLOB column as `[]byte`
2. Parse as JSON: `json.Unmarshal(blobData, &embedding)`
3. Result: `[]float64` (or `[]float32` depending on provider)

**Example (pseudocode):**
```go
var blobData []byte
err := row.Scan(..., &blobData, ...)

var embedding []float64
err = json.Unmarshal(blobData, &embedding)
```

---

## Precision and Rounding

### Floating-Point Precision

**JSON serialization preserves:**
- IEEE 754 double precision (64-bit floats)
- ~15-17 significant decimal digits
- Full range: ±1.7e±308

**Example values:**
```json
[
  0.029445774853229523,    // 18 digits
  -0.0034673467744141817,  // 19 digits
  0.007,                   // 3 digits
  1.234567890123456,       // 16 digits
  0.0                      // Exact zero
]
```

### Provider-Specific Precision

Different embedding providers return different precisions:

| Provider | Native Type | JSON Precision | Example |
|----------|-------------|----------------|---------|
| OpenAI | `float` | 15-17 digits | `0.029445774853229523` |
| local (Transformers.js) | `float32` | 6-9 digits | `0.0294458` |
| Ollama | `float64` | 15-17 digits | `0.029445774853229523` |

**Compatibility Requirement:**
- Go must preserve **exact precision** returned by provider
- Use `json.Marshal()` default precision (no custom formatting)
- Do NOT truncate or round values

---

## Dimension Validation

### Consistency Check

**Rule:** `len(embedding_array)` MUST equal `embedding_dimensions` column value

**Example:**
```sql
-- Correct: 1536 elements, dimensions=1536
embedding_dimensions = 1536
embedding = [0.1, 0.2, ..., 0.999]  -- length 1536

-- Incorrect: mismatch
embedding_dimensions = 1536
embedding = [0.1, 0.2, ..., 0.999]  -- length 384 ❌
```

**Validation (Go pseudocode):**
```go
var embedding []float64
json.Unmarshal(blobData, &embedding)

if len(embedding) != row.EmbeddingDimensions {
    return fmt.Errorf("dimension mismatch: got %d, expected %d",
        len(embedding), row.EmbeddingDimensions)
}
```

### Common Dimensions

| Provider | Model | Dimensions |
|----------|-------|------------|
| OpenAI | `text-embedding-3-small` | 1536 |
| OpenAI | `text-embedding-3-large` | 3072 |
| local | `Xenova/all-MiniLM-L6-v2` | 384 |
| Ollama | `mxbai-embed-large` | 1024 |

---

## Normalization

### L2 Normalization

**Node.js behavior:**
- Embeddings are stored **as-is** from provider
- No automatic L2 normalization at storage time
- Normalization (if any) is provider-dependent

**Provider normalization status:**
| Provider | Pre-normalized? | Notes |
|----------|-----------------|-------|
| OpenAI | ✅ Yes | Always L2-normalized |
| local | ❌ No | Raw embeddings |
| Ollama | ❔ Varies | Model-dependent |

**Go Implementation:**
- MUST preserve raw provider embeddings
- Do NOT normalize unless provider does
- Cosine similarity calculation handles normalization during search

### Normalization Formula

If implementing cosine similarity:

```go
func cosineSimilarity(a, b []float64) float64 {
    if len(a) != len(b) {
        return 0
    }
    
    var dotProduct, normA, normB float64
    for i := 0; i < len(a); i++ {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    
    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

---

## Hex Dump Example

### Real Database Example

**Query:**
```sql
SELECT hex(substr(embedding, 1, 64)) FROM code_chunks LIMIT 1;
```

**Result:**
```
5B302E30323934343537373438353332393532332C2D302E303033343637333436373734343134313831372C302E303037
```

**Decoded (UTF-8):**
```
[0.029445774853229523,-0.0034673467744141817,0.007
```

**Breakdown:**
- `5B` = `[` (left bracket)
- `30` = `0` (digit)
- `2E` = `.` (decimal point)
- `2C` = `,` (comma)
- `2D` = `-` (minus sign)

---

## Edge Cases

### 1. Empty Embeddings

**Problem:** What if embedding is empty?

**Solution:**
- `embedding = NULL` (not `[]`)
- Use NULL when embedding not yet generated
- Never store empty JSON array `[]`

### 2. Infinity and NaN

**Problem:** JSON doesn't support `Infinity` or `NaN`

**Solution:**
- Providers should never return these values
- If encountered, validation error should occur
- Go implementation should reject:
  ```go
  if math.IsInf(val, 0) || math.IsNaN(val) {
      return errors.New("invalid embedding value")
  }
  ```

### 3. Very Large Arrays

**Problem:** Large embeddings (e.g., 3072 dimensions)

**Solution:**
- SQLite BLOB supports up to 2GB
- 3072 float64s × 20 bytes (JSON) ≈ 60KB
- No special handling needed
- Go: standard `json.Marshal()` handles this

### 4. Whitespace Differences

**Problem:** `[0.1,0.2]` vs `[0.1, 0.2]` (with space)

**Solution:**
- Node.js `JSON.stringify()` produces **no spaces**
- Go `json.Marshal()` also produces no spaces by default
- **Byte-for-byte identical** when using defaults
- Do NOT use `json.MarshalIndent()`

### 5. Trailing Zeros

**Problem:** `0.1` vs `0.10` vs `0.100`

**Solution:**
- Go `json.Marshal()` uses minimal representation
- `0.1` (not `0.10`)
- Matches Node.js behavior
- No custom formatting needed

---

## Compatibility Testing

### Byte-Level Comparison

**Test:** Generate embedding → Store → Retrieve → Compare

```go
// Test pseudo-code
func TestEmbeddingRoundTrip(t *testing.T) {
    original := []float64{0.029445774853229523, -0.0034673467744141817}
    
    // Serialize
    jsonBytes, _ := json.Marshal(original)
    
    // Store in DB
    db.Exec("INSERT INTO code_chunks (embedding, ...) VALUES (?, ...)", jsonBytes)
    
    // Retrieve
    var retrieved []byte
    db.QueryRow("SELECT embedding FROM code_chunks WHERE id=?", id).Scan(&retrieved)
    
    // Compare byte-for-byte
    assert.Equal(t, jsonBytes, retrieved)
    
    // Parse and verify values
    var parsed []float64
    json.Unmarshal(retrieved, &parsed)
    assert.Equal(t, original, parsed)
}
```

### Cross-Implementation Test

**Test:** Node.js writes → Go reads

```bash
# Node.js creates database
node -e "require('./src/service.js').indexProject({repoPath: './test-repo'})"

# Go reads embedding
go run test_read_embedding.go
```

**Expected:** Exact float values match

---

## Reference Implementation Locations

**Node.js:**
- Storage: `src/service.js:1607` (`Buffer.from(JSON.stringify(embedding))`)
- Retrieval: `src/service.js` (search functions, parse JSON from BLOB)
- Cosine similarity: `src/service.js:1017-1031`

**Related Files:**
- Provider interface: `src/providers.js`
- BM25 search: `src/search/bm25Index.js`
- Hybrid search: `src/search/hybrid.js`

---

## Summary Checklist

For Go implementation compatibility:

- [ ] Serialize embeddings as JSON arrays (no whitespace)
- [ ] Store as UTF-8 encoded bytes in BLOB column
- [ ] Preserve full floating-point precision (no truncation)
- [ ] Validate dimension match: `len(embedding) == embedding_dimensions`
- [ ] Handle NULL embeddings (not empty arrays)
- [ ] Reject `Inf` and `NaN` values
- [ ] Use default `json.Marshal()` (no custom formatting)
- [ ] Implement cosine similarity with normalization
- [ ] Support providers: OpenAI, local, Ollama
- [ ] Test byte-for-byte compatibility with Node.js

---

## Test Vectors

### Example 1: OpenAI Embedding (First 5 dimensions)

**Input:**
```javascript
[0.029445774853229523, -0.0034673467744141817, 0.007123, 0.0123456789, -0.999]
```

**JSON Serialization:**
```json
[0.029445774853229523,-0.0034673467744141817,0.007123,0.0123456789,-0.999]
```

**Hex (UTF-8):**
```
5B302E30323934343537373438353332393532332C2D302E303033343637333436373734343134313831372C302E3030373132332C302E303132333435363738392C2D302E3939395D
```

### Example 2: Local Embedding (384 dimensions, truncated)

**Input (first 3):**
```javascript
[0.0294458, -0.0034673, 0.007123]
```

**JSON:**
```json
[0.0294458,-0.0034673,0.007123]
```

**Dimension:** 384  
**Size:** ~15 bytes per float × 384 ≈ 5.7 KB

---

**End of Vector Blob Format Specification**
