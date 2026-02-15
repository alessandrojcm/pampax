# SQLite Schema Contract

**Version:** 1.0  
**Created:** 2026-01-28  
**Status:** Reference Specification for Go Port Compatibility

---

## Overview

This document defines the exact SQLite database schema used by PAMPAX for storing code chunks, embeddings, and semantic metadata. The Go implementation MUST produce byte-compatible databases that are interchangeable with the Node.js implementation.

## Database File

- **Location:** `.pampa/pampa.db` (relative to project root)
- **Format:** SQLite3 database file
- **SQLite Version:** 3.44.4+ (tested with 3.44.4)
- **Compatibility:** Standard SQLite3 format with default settings

## Pragmas and Settings

### Required Pragmas (Observed Values)

```sql
PRAGMA page_size = 4096;           -- Default page size
PRAGMA journal_mode = delete;      -- Default journaling mode
PRAGMA encoding = 'UTF-8';         -- Text encoding
PRAGMA foreign_keys = OFF;         -- Foreign keys disabled (default)
```

### Compatibility Notes

- **Page size:** While 4096 is default, Go implementation should accept existing databases with different page sizes
- **Journal mode:** `delete` mode is default and portable across platforms
- **Encoding:** UTF-8 is mandatory for text fields
- **Foreign keys:** Currently disabled; Go implementation should match this behavior

### Platform Portability

- Database files are **cross-platform compatible** (Windows/Linux/macOS)
- No platform-specific pragmas are used
- Byte order is handled automatically by SQLite

---

## Table Schemas

### 1. `code_chunks` - Primary Chunk Storage

**Purpose:** Stores code chunks with embeddings and metadata.

```sql
CREATE TABLE code_chunks (
    -- Primary identification
    id TEXT PRIMARY KEY,
    file_path TEXT NOT NULL,
    symbol TEXT NOT NULL,
    sha TEXT NOT NULL,
    lang TEXT NOT NULL,
    chunk_type TEXT DEFAULT 'function',
    
    -- Embedding data
    embedding BLOB,
    embedding_provider TEXT,
    embedding_dimensions INTEGER,
    
    -- Enhanced semantic metadata
    pampa_tags TEXT,           -- JSON array of semantic tags
    pampa_intent TEXT,         -- Natural language intent description
    pampa_description TEXT,    -- Human-readable description
    doc_comments TEXT,         -- JSDoc/PHPDoc comments
    variables_used TEXT,       -- JSON array of important variables
    context_info TEXT,         -- Additional context metadata
    
    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### Field Specifications

| Field | Type | Nullable | Default | Description |
|-------|------|----------|---------|-------------|
| `id` | TEXT | NO | - | **Primary key**. Chunk identifier (format: `{file}:{symbol}:{hash}`) |
| `file_path` | TEXT | NO | - | Project-relative file path (uses `/` separator on all platforms) |
| `symbol` | TEXT | NO | - | Function/class/method name (empty string if no symbol) |
| `sha` | TEXT | NO | - | SHA-1 hash of chunk content (hex-encoded, 40 chars) |
| `lang` | TEXT | NO | - | Language identifier (e.g., `javascript`, `python`, `go`) |
| `chunk_type` | TEXT | YES | `'function'` | Chunk classification (e.g., `function`, `class`, `method`) |
| `embedding` | BLOB | YES | NULL | **Vector embedding** (see Vector Blob Format section) |
| `embedding_provider` | TEXT | YES | NULL | Provider name (e.g., `OpenAI`, `local`, `Ollama`) |
| `embedding_dimensions` | INTEGER | YES | NULL | Vector dimension count (e.g., `384`, `1536`, `3072`) |
| `pampa_tags` | TEXT | YES | NULL | **JSON array** of semantic tags (e.g., `["stripe","payment"]`) |
| `pampa_intent` | TEXT | YES | NULL | Natural language intent description |
| `pampa_description` | TEXT | YES | NULL | Human-readable description |
| `doc_comments` | TEXT | YES | NULL | Extracted documentation comments (raw text) |
| `variables_used` | TEXT | YES | NULL | **JSON array** of important variable objects |
| `context_info` | TEXT | YES | NULL | **JSON object** with additional context |
| `created_at` | DATETIME | YES | `CURRENT_TIMESTAMP` | ISO 8601 timestamp (e.g., `2026-01-28 12:34:56`) |
| `updated_at` | DATETIME | YES | `CURRENT_TIMESTAMP` | ISO 8601 timestamp |

#### Constraints and Invariants

1. **Primary Key:**
   - `id` must be unique
   - Format: `{file_path}:{symbol}:{sha_suffix}` (implementation-specific)
   - Example: `src/service.js:indexProject:a681484f`

2. **Path Normalization:**
   - `file_path` uses forward slashes (`/`) on **all platforms** (including Windows)
   - Paths are **relative** to project root
   - No leading `/` or `./`
   - Example: `src/utils/logger.js` (NOT `./src/utils/logger.js`)

3. **SHA Format:**
   - 40-character hexadecimal string (lowercase)
   - SHA-1 hash of chunk content
   - Example: `5ea95a5a78779486d1fccdab927a7d64f5cf1599`

4. **JSON Fields:**
   - `pampa_tags`: Array of strings, e.g., `["auth", "login", "user"]`
   - `variables_used`: Array of objects with `type`, `name`, `value` keys
   - `context_info`: Object with arbitrary metadata
   - **Must be valid JSON or NULL** (not empty strings)

5. **Symbol Field:**
   - Empty string (`''`) if no symbol (NOT NULL)
   - Contains function/class/method name
   - May include parent context (e.g., `ClassName::methodName`)

6. **Timestamps:**
   - ISO 8601 format: `YYYY-MM-DD HH:MM:SS`
   - UTC timezone (implicit)
   - Generated by SQLite `CURRENT_TIMESTAMP` function

#### Indices on `code_chunks`

```sql
-- Single-column indices
CREATE INDEX idx_file_path ON code_chunks(file_path);
CREATE INDEX idx_symbol ON code_chunks(symbol);
CREATE INDEX idx_lang ON code_chunks(lang);
CREATE INDEX idx_provider ON code_chunks(embedding_provider);
CREATE INDEX idx_chunk_type ON code_chunks(chunk_type);
CREATE INDEX idx_pampa_tags ON code_chunks(pampa_tags);
CREATE INDEX idx_pampa_intent ON code_chunks(pampa_intent);

-- Composite index for provider/dimension filtering
CREATE INDEX idx_lang_provider
    ON code_chunks(lang, embedding_provider, embedding_dimensions);
```

**Index Purpose:**
- `idx_file_path`: File-scoped searches
- `idx_symbol`: Symbol-based lookups
- `idx_lang`: Language filtering
- `idx_provider`: Provider-specific queries
- `idx_chunk_type`: Chunk type filtering
- `idx_pampa_tags`: Tag-based search
- `idx_pampa_intent`: Intent-based search
- `idx_lang_provider`: Efficient provider/dimension filtering for vector search

---

### 2. `intention_cache` - Query Intent Mapping

**Purpose:** Caches user query intent to chunk mappings for learning user patterns.

```sql
CREATE TABLE intention_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query_normalized TEXT NOT NULL,
    original_query TEXT NOT NULL,
    target_sha TEXT NOT NULL,
    confidence REAL DEFAULT 1.0,
    usage_count INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### Field Specifications

| Field | Type | Nullable | Default | Description |
|-------|------|----------|---------|-------------|
| `id` | INTEGER | NO | AUTOINCREMENT | Primary key (auto-incremented) |
| `query_normalized` | TEXT | NO | - | Normalized query string (lowercase, trimmed) |
| `original_query` | TEXT | NO | - | Original user query (preserves casing) |
| `target_sha` | TEXT | NO | - | SHA of target chunk |
| `confidence` | REAL | YES | `1.0` | Confidence score (0.0-1.0) |
| `usage_count` | INTEGER | YES | `1` | Number of times pattern observed |
| `created_at` | DATETIME | YES | `CURRENT_TIMESTAMP` | First observation timestamp |
| `last_used` | DATETIME | YES | `CURRENT_TIMESTAMP` | Most recent usage timestamp |

#### Indices on `intention_cache`

```sql
CREATE INDEX idx_query_normalized ON intention_cache(query_normalized);
CREATE INDEX idx_target_sha ON intention_cache(target_sha);
CREATE INDEX idx_usage_count ON intention_cache(usage_count DESC);
```

---

### 3. `query_patterns` - Pattern Frequency Analysis

**Purpose:** Tracks frequently observed query patterns for optimization.

```sql
CREATE TABLE query_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pattern TEXT NOT NULL UNIQUE,
    frequency INTEGER DEFAULT 1,
    typical_results TEXT,  -- JSON array of common results
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

#### Field Specifications

| Field | Type | Nullable | Default | Description |
|-------|------|----------|---------|-------------|
| `id` | INTEGER | NO | AUTOINCREMENT | Primary key (auto-incremented) |
| `pattern` | TEXT | NO | - | Query pattern (UNIQUE constraint) |
| `frequency` | INTEGER | YES | `1` | Observation count |
| `typical_results` | TEXT | YES | NULL | **JSON array** of common result SHAs |
| `created_at` | DATETIME | YES | `CURRENT_TIMESTAMP` | First observation |
| `updated_at` | DATETIME | YES | `CURRENT_TIMESTAMP` | Last update |

#### Indices on `query_patterns`

```sql
CREATE INDEX idx_pattern_frequency ON query_patterns(frequency DESC);
```

---

### 4. `sqlite_sequence` - Internal SQLite Table

**Purpose:** Auto-generated by SQLite for AUTOINCREMENT tracking.

```sql
CREATE TABLE sqlite_sequence(name, seq);
```

**Note:** This table is **automatically managed by SQLite** and should not be manually created or modified. It tracks the highest ID values for tables using `AUTOINCREMENT`.

---

## Embedding Storage

Embeddings are stored in the `embedding` BLOB column as UTF-8 encoded JSON arrays of floats.
The exact serialization rules are specified in:

- [02_vector_blob_format.md](./02_vector_blob_format.md)

---

## Migration and Versioning

### Current Version

- **Schema Version:** 1.0 (implicit, no version table)
- **Node.js Reference:** PAMPAX v1.15.0+

### Future Considerations

- No schema version tracking currently exists
- Go implementation should consider adding:
  ```sql
  CREATE TABLE schema_version (
      version INTEGER PRIMARY KEY,
      applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );
  ```

### Breaking Changes

Any changes to:
- Column names or types
- Table names
- Index definitions
- BLOB format (JSON → binary)

...require a **full re-index** and are considered breaking changes.

---

## Reference Implementation

**Source Files:**
- Schema definition: `src/service.js:922-1014` (Node.js)
- Embedding storage: `src/service.js:1607` (Node.js)
- Embedding retrieval: Search in `searchCode()` function

**Testing:**
- Validate schema: `sqlite3 .pampa/pampa.db ".schema"`
- Check pragmas: `sqlite3 .pampa/pampa.db "PRAGMA encoding;"`
- Inspect data: `sqlite3 .pampa/pampa.db "SELECT * FROM code_chunks LIMIT 1;"`

---

## Summary

This schema contract defines:
1. ✅ Exact table structures with column types and constraints
2. ✅ Index definitions for performance
3. ✅ Vector blob format (JSON-serialized arrays)
4. ✅ Path normalization rules (always `/`)
5. ✅ Timestamp format (ISO 8601 UTC)
6. ✅ NULL vs empty string semantics
7. ✅ Platform portability requirements
8. ✅ Validation checklist for Go implementation

**Critical for Compatibility:**
- Vector BLOB format (JSON strings, not binary)
- Path separator normalization (always `/`)
- JSON field validation (NULL vs `"[]"`)
- Timestamp format (no timezone suffix)
