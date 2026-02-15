# Stage 1 — Go Scaffold Implementation Plan

**Version:** 1.0  
**Created:** 2026-02-05  
**Status:** Ready for Implementation  
**Module:** `github.com/alessandrojcm/pampax-go`

---

## Overview

This document defines the implementation plan for **Stage 1** of the PAMPAX Go port. Stage 1 focuses on establishing the project scaffold, implementing compatibility-critical modules, and ensuring byte-level parity with the Node.js implementation for all artifacts defined in Stage 0.

**Stage 1 Scope (CLI Commands):**
1. `index` — produce `.pampa/` artifacts (db + chunks + codemap + merkle)
2. `update` — reindex (full reindex for parity)
3. `search` — read db + chunks and return results (top-10 matching Node)
4. `info` — basic stats / health (db exists, chunk count, provider info)

**Out of Scope (Stage 1):**
- `watch` (requires filesystem watcher & debouncing)
- `context list/show/use` (context packs are more advanced)
- `mcp` (server)

---

## Technology Stack

### Core Libraries
- **Go Version:** 1.26
- **SQLite:** `modernc.org/sqlite` (pure Go, no CGO)
- **Query Generation:** `github.com/sqlc-dev/sqlc`
- **Migrations:** `github.com/amacneil/dbmate`

### CLI & Config
- **CLI Framework:** `github.com/spf13/cobra`
- **Config Management:** `github.com/spf13/viper`

### Validation & Logging
- **JSON Validation:** `github.com/go-playground/validator/v10`
- **Logging:** `github.com/rs/zerolog`
  - JSON logs by default (agent-friendly)
  - Console format via `--pretty` flag

### Crypto & Compression
- **Stdlib:** `crypto/sha1`, `crypto/aes`, `crypto/cipher`, `crypto/rand`
- **HKDF:** `golang.org/x/crypto/hkdf`
- **Gzip:** `compress/gzip` (stdlib)

### JSON & Ordering
- **JSON:** `encoding/json` (stdlib)
- **Ordered Map:** Custom implementation (unit tested)

---

## Project Structure

```
github.com/alessandrojcm/pampax-go/
├── cmd/
│   └── pampax/
│       └── main.go                    # CLI entrypoint (Cobra root)
├── internal/
│   ├── app/                           # Command orchestration (index/search/update/info)
│   ├── config/                        # Viper config + env parsing + defaults
│   ├── db/                            # sqlc queries + DB access layer
│   │   └── sqlc.yaml                  # sqlc config
│   ├── migrations/                    # dbmate migration files
│   ├── chunks/                        # SHA-1, gzip, encryption, atomic writes
│   ├── codemap/                       # Ordered map, normalization, JSON serialization
│   ├── indexer/                       # File discovery, chunking, language detection
│   ├── providers/                     # Embedding provider interfaces + stubs
│   ├── search/                        # Cosine + BM25/hybrid (stubs for Stage 1)
│   ├── merkle/                        # Merkle tree generation
│   ├── compat/                        # Node/Go compatibility helpers
│   └── utils/                         # Path, UTF-8, timestamps, errors
├── sql/
│   ├── schema.sql                     # DB schema for dbmate
│   └── queries/                       # sqlc query files
│       ├── chunks.sql
│       ├── intention_cache.sql
│       └── query_patterns.sql
├── test/
│   ├── fixtures/                      # Golden fixtures (pre-generated from Node)
│   ├── compat/                        # Cross-implementation compatibility tests
│   └── unit/                          # Unit tests (ordered map, JSON encoding, etc.)
├── go.mod
├── go.sum
└── README.md
```

---

## Stage 0 — Planning & Constraints ✅

**Status:** Complete

**Confirmed Decisions:**
- CLI Scope: `index`, `update`, `search`, `info`
- JSON Validation: **warn and skip invalid JSON fields**
- Logging: JSON by default; `--pretty` flag for console
- Symbol handling: **empty string in DB**, **null in codemap**
- Codemap top-level keys: **insertion order preserved**
- Module path: `github.com/alessandrojcm/pampax-go`

---

## Stage 1 — Project Skeleton

**Dependencies:** None (can run in parallel)

### 1A) Repository Layout

**References:** None (setup only)

**Goals:**
- Establish directory structure
- Initialize Go module

**Tasks:**
1. Create directory tree (see structure above)
2. Initialize `go.mod` with module path `github.com/alessandrojcm/pampax-go`
3. Add `.gitignore` for Go artifacts (if not present)

**Deliverables:**
- [x] Directory structure created
- [x] `go.mod` initialized with Go 1.25.5
- [x] `.gitignore` updated for Go

---

### 1B) Tooling Configuration

**References:** None (setup only)

**Goals:**
- Configure sqlc, dbmate, and development tools

**Tasks:**
1. Create `internal/db/sqlc.yaml`:
   ```yaml
   version: "2"
   sql:
     - engine: "sqlite"
       queries: "../../sql/queries"
       schema: "../../sql/schema.sql"
       gen:
         go:
           package: "db"
           out: "."
           emit_json_tags: true
           emit_prepared_queries: false
           emit_interface: true
           emit_exact_table_names: false
   ```

2. Create dbmate config (or use env vars):
   - `DATABASE_URL` = `sqlite:.pampa/pampa.db`
   - Migrations dir = `internal/migrations`

3. Install dependencies:
   ```bash
   go get modernc.org/sqlite
   go get github.com/spf13/cobra
   go get github.com/spf13/viper
   go get github.com/go-playground/validator/v10
   go get github.com/rs/zerolog
   go get golang.org/x/crypto/hkdf
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   go install github.com/amacneil/dbmate/v2@latest
   ```

**Deliverables:**
- [x] `internal/db/sqlc.yaml` configured
- [x] dbmate config set up (.env file with DATABASE_URL and DBMATE_MIGRATIONS_DIR)
- [x] Dependencies added to `go.mod`
- [x] CLI tools installed (sqlc v1.30.0, dbmate v2.29.5)

---

## Stage 2 — Compatibility-Critical Modules

**Dependencies:** Stage 1 complete

### 2A) Codemap Ordered Map & Serialization

**References:**
- `instructions/stage0_artifacts/04_codemap_json_schema.md` — JSON structure, field ordering, null vs omitted semantics, insertion order preservation
- `instructions/stage0_artifacts/05_edge_cases_platform_differences.md` — Path normalization (forward slashes), UTF-8 handling

**Goals:**
- Custom ordered map preserving insertion order
- JSON serialization with 2-space indent + final newline
- Alphabetical keys inside chunk metadata objects

**Tasks:**
1. Implement `internal/codemap/orderedmap.go`:
   - Type: `OrderedMap` with insertion order tracking
   - Methods: `Set(key, value)`, `Get(key)`, `Keys()`, `MarshalJSON()`

2. Implement `internal/codemap/serialize.go`:
   - Function to write codemap with correct formatting
   - 2-space indent, Unix newlines, final newline

3. Implement `internal/codemap/types.go`:
   - `ChunkMetadata` struct matching Stage 0 schema
   - JSON tags with `omitempty` where appropriate
   - Handle `symbol` as pointer (null vs empty string)

4. Implement normalization rules:
   - Null vs omitted semantics
   - Empty array handling (synonyms, symbol_calls, etc.)
   - Field presence rules

**Unit Tests:**
- [x] Insertion order preserved in JSON output
- [x] Top-level keys NOT alphabetically sorted
- [x] Chunk metadata fields ARE alphabetically sorted
- [x] 2-space indent, Unix newlines, final newline
- [x] `symbol: null` when no symbol (not omitted)
- [x] `symbol_parameters` omitted when empty (not `[]`)
- [x] Forward slashes in paths on all platforms

**Deliverables:**
- [x] `internal/codemap/orderedmap.go` implemented
- [x] `internal/codemap/serialize.go` implemented
- [x] `internal/codemap/types.go` implemented
- [x] Unit tests passing

---

### 2B) Chunk Storage & Crypto

**References:**
- `instructions/stage0_artifacts/03_chunk_file_format.md` — SHA-1 naming, gzip compression, AES-256-GCM encryption, PAMPAE1 header format, HKDF key derivation
- `instructions/stage0_artifacts/05_edge_cases_platform_differences.md` — Line ending preservation, BOM preservation, UTF-8 handling, atomic writes

**Goals:**
- SHA-1 on raw UTF-8 bytes (preserve BOM & line endings)
- gzip default compression
- AES-256-GCM + HKDF (PAMPAE1 magic header)
- Atomic writes

**Tasks:**
1. Implement `internal/chunks/sha.go`:
   - `ComputeSHA(code string) string` — SHA-1 of raw bytes

2. Implement `internal/chunks/gzip.go`:
   - `Compress(data []byte) ([]byte, error)`
   - `Decompress(data []byte) ([]byte, error)`

3. Implement `internal/chunks/encrypt.go`:
   - `DeriveChunkKey(masterKey, salt []byte) ([]byte, error)` — HKDF-SHA256
   - `Encrypt(gzipped, masterKey []byte) ([]byte, error)` — AES-256-GCM
   - `Decrypt(payload, masterKey []byte) ([]byte, error)`
   - Magic header: `PAMPAE1`

4. Implement `internal/chunks/storage.go`:
   - `WriteChunk(chunkDir, sha, code string, encrypted bool, masterKey []byte) error`
   - `ReadChunk(chunkDir, sha string, encrypted bool, masterKey []byte) (string, error)`
   - `RemoveChunk(chunkDir, sha string) error` — removes both .gz and .gz.enc
   - Atomic writes: temp file + rename

**Unit Tests:**
- [x] SHA-1 matches Node.js for same content (including `\r\n`)
- [x] Gzip roundtrip preserves exact bytes
- [x] Encrypted roundtrip with correct magic header
- [x] HKDF key derivation matches Node.js
- [x] AES-GCM authentication tag verified
- [x] Atomic write (temp + rename)
- [x] BOM preservation

**Deliverables:**
- [x] `internal/chunks/{sha,gzip,encrypt,storage}.go` implemented
- [x] Unit tests passing

---

### 2C) DB Schema + sqlc Queries

**References:**
- `instructions/stage0_artifacts/01_sqlite_schema_contract.md` — Exact table schemas, column types, indices, constraints, pragmas (UTF-8, page size, journal mode)
- `instructions/stage0_artifacts/02_vector_blob_format.md` — Embedding BLOB format (JSON array bytes), dimension validation, precision requirements
- `instructions/stage0_artifacts/05_edge_cases_platform_differences.md` — Symbol empty string vs NULL, timestamp formats, JSON field handling

**Goals:**
- Schema matches Stage 0 exactly
- sqlc for all queries
- JSON validation on all JSON fields (warn & skip invalid)

**Tasks:**
1. Create `sql/schema.sql` matching Stage 0:
   - `code_chunks` table with all fields
   - `intention_cache` table
   - `query_patterns` table
   - All indices from Stage 0

2. Create dbmate migration in `internal/migrations/`:
   - `001_create_schema.sql` (same as schema.sql)

3. Create sqlc queries in `sql/queries/`:
   - `chunks.sql`:
     - `InsertChunk`
     - `GetChunkBySHA`
     - `GetChunksByFilePath`
     - `GetChunksByProvider`
     - `DeleteChunk`
   - `intention_cache.sql`:
     - `InsertIntention`
     - `GetIntentionsByQuery`
   - `query_patterns.sql`:
     - `InsertQueryPattern`
     - `GetQueryPatternByPattern`

4. Generate sqlc types:
   ```bash
   cd internal/db && sqlc generate
   ```

5. Implement `internal/db/validation.go`:
   - Use `go-playground/validator` for JSON fields
   - Validate: `pampa_tags`, `variables_used`, `context_info`
   - Behavior: **warn and skip invalid JSON fields** (log with zerolog)

**Unit Tests:**
- [x] Schema pragmas match (UTF-8, page size, journal mode)
- [x] All tables and indices created
- [x] JSON validation warns and skips invalid fields
- [x] Embedding BLOB serialization (JSON array bytes, no whitespace)
- [x] Symbol empty string in DB (not NULL)

**Deliverables:**
- [x] `sql/schema.sql` created
- [x] dbmate migration created
- [x] sqlc queries created
- [x] sqlc types generated
- [x] `internal/db/validation.go` implemented
- [x] Unit tests passing

---

## Stage 3 — CLI & Config

**Dependencies:** Stage 2 complete

### 3A) CLI Commands

**References:** None (application layer — uses implementation from Stage 2)

**Goals:**
- Cobra skeleton with Stage-1 commands

**Tasks:**
1. Implement `cmd/pampax/main.go`:
   - Cobra root command
   - Persistent flags: `--pretty`, `--config`, `--verbose`

2. Implement `cmd/pampax/index.go`:
   - Command: `pampax index [path]`
   - Flags: `--provider`, `--encryption-key`, etc.

3. Implement `cmd/pampax/update.go`:
   - Command: `pampax update [path]`
   - Full reindex for Stage 1

4. Implement `cmd/pampax/search.go`:
   - Command: `pampax search <query>`
   - Flags: `--lang`, `--path`, `--top`, etc.

5. Implement `cmd/pampax/info.go`:
   - Command: `pampax info`
   - Display: chunk count, provider, dimensions, db size

**Deliverables:**
- [x] `cmd/pampax/main.go` root command
- [x] `cmd/pampax/{index,update,search,info}.go` commands
- [x] Flags match Node.js CLI where applicable

---

### 3B) Config & Logging

**References:**
- `instructions/stage0_artifacts/03_chunk_file_format.md` — Environment variables: `PAMPAX_ENCRYPTION_KEY` format (base64/hex)
- `instructions/stage0_artifacts/05_edge_cases_platform_differences.md` — Environment variable handling, path separator conventions

**Goals:**
- Viper config / env parsing
- Zerolog JSON by default; console on `--pretty`

**Tasks:**
1. Implement `internal/config/config.go`:
   - Viper setup for env vars (matching Node.js env vars from Stage 0)
   - Config struct with defaults

2. Implement `internal/utils/logging.go`:
   - Zerolog setup
   - JSON output by default
   - Console writer when `--pretty` flag set

3. Environment variables to support:
   - `PAMPAX_ENCRYPTION_KEY`
   - `PAMPAX_OPENAI_API_KEY`
   - `PAMPAX_OPENAI_BASE_URL`
   - `PAMPAX_OPENAI_EMBEDDING_MODEL`
   - `PAMPAX_MAX_TOKENS`
   - `PAMPAX_DIMENSIONS`
   - `PAMPAX_RATE_LIMIT`
   - `PAMPAX_RERANKER_*`

**Deliverables:**
- [x] `internal/config/config.go` with Viper setup
- [x] `internal/utils/logging.go` with Zerolog
- [x] `--pretty` flag toggles console logs

---

## Stage 4 — Provider Stubs & Search

**Dependencies:** Stage 3 complete

### 4A) Provider Stubs

**References:**
- `instructions/stage0_artifacts/02_vector_blob_format.md` — Provider-specific dimensions (OpenAI: 1536/3072, local: 384, Ollama: 1024), precision requirements

**Goals:**
- Provider interface with stubs for OpenAI/local/Ollama

**Tasks:**
1. Implement `internal/providers/provider.go`:
   - Interface: `EmbeddingProvider`
   - Method: `GenerateEmbedding(text string) ([]float64, error)`
   - Method: `GetDimensions() int`
   - Method: `GetName() string`

2. Implement stubs:
   - `internal/providers/openai_stub.go`
   - `internal/providers/local_stub.go`
   - `internal/providers/ollama_stub.go`
   - Each returns fake embeddings for Stage 1

**Deliverables:**
- [x] `internal/providers/provider.go` interface
- [x] Stub implementations (return fake embeddings)

---

### 4B) Search Stub

**References:**
- `instructions/stage0_artifacts/02_vector_blob_format.md` — Cosine similarity calculation, L2 normalization

**Goals:**
- Search pipeline stub (cosine, BM25, hybrid placeholders)

**Tasks:**
1. Implement `internal/search/cosine.go`:
   - `CosineSimilarity(a, b []float64) float64`

2. Implement `internal/search/search.go`:
   - Stub function: `Search(query string, options SearchOptions) ([]Result, error)`
   - For Stage 1: return top-10 by cosine similarity only

**Deliverables:**
- [ ] `internal/search/cosine.go` implemented
- [ ] `internal/search/search.go` stub

---

## Stage 5 — Automated Tests

**Dependencies:** Stage 4 complete

### 5A) Compatibility Tests

**References:**
- `instructions/stage0_artifacts/01_sqlite_schema_contract.md` — DB schema validation, pragma checks
- `instructions/stage0_artifacts/02_vector_blob_format.md` — Embedding BLOB validation, dimension checks, cosine similarity
- `instructions/stage0_artifacts/03_chunk_file_format.md` — Chunk SHA-1 validation, gzip decompression, encrypted chunk format
- `instructions/stage0_artifacts/04_codemap_json_schema.md` — Codemap JSON structure, ordering, field presence
- `instructions/stage0_artifacts/05_edge_cases_platform_differences.md` — All edge cases (paths, line endings, UTF-8, timestamps)
- `test/fixtures/small/` and `test/baselines/node_baseline_2026-01-28.json` — Golden fixtures for testing

**Goals:**
- Use **pre-generated Node fixtures** (Option B)
- Automated compatibility checks

**Tasks:**
1. Copy Node fixtures to `test/fixtures/`:
   - Use `test/fixtures/small/` from Node repo
   - Use `test/baselines/node_baseline_2026-01-28.json`

2. Implement `test/compat/db_test.go`:
   - Go reads Node `.pampa/pampa.db`
   - Validate schema matches
   - Validate embedding BLOB format (JSON arrays)

3. Implement `test/compat/chunks_test.go`:
   - Go reads Node `.pampa/chunks/*.gz`
   - Validate SHA-1 matches decompressed content
   - Go reads Node `.pampa/chunks/*.gz.enc`
   - Validate decryption + magic header

4. Implement `test/compat/codemap_test.go`:
   - Go reads Node `pampa.codemap.json`
   - Validate top-level insertion order preserved
   - Validate field keys alphabetically sorted
   - Validate null vs omitted semantics

5. Implement `test/compat/search_test.go`:
   - Compare Go search results to Node baseline JSON
   - Validate top-10 ordering
   - Validate score tolerance (±0.01 or ±1%)

**Deliverables:**
- [ ] Node fixtures copied to `test/fixtures/`
- [ ] Compatibility tests implemented
- [ ] Tests passing

---

### 5B) Unit Tests

**References:**
- `instructions/stage0_artifacts/04_codemap_json_schema.md` — Ordered map tests, JSON formatting
- `instructions/stage0_artifacts/05_edge_cases_platform_differences.md` — Path normalization tests, UTF-8 validation
- `instructions/stage0_artifacts/02_vector_blob_format.md` — Embedding BLOB serialization tests

**Goals:**
- Unit tests for critical modules

**Tasks:**
1. `test/unit/orderedmap_test.go`:
   - Insertion order
   - JSON serialization
   - Field ordering

2. `test/unit/json_validation_test.go`:
   - Valid JSON fields accepted
   - Invalid JSON fields warned and skipped

3. `test/unit/path_normalization_test.go`:
   - Windows paths → forward slashes
   - Relative path normalization

4. `test/unit/embedding_blob_test.go`:
   - JSON array serialization
   - No whitespace in output
   - Dimension validation

**Deliverables:**
- [ ] Unit tests implemented
- [ ] Tests passing

---

## Stage 6 — Index Update & Verification

**Dependencies:** Stage 5 complete

**References:**
- `AGENTS.md` — PAMPAX workflow guidelines for indexing and verification
- `instructions/stage0_artifacts/README.md` — Compatibility acceptance criteria and validation inputs

**Goals:**
- Keep PAMPAX index synchronized after code changes (per AGENTS.md)

**Tasks:**
1. After implementing Stage 1-5, run:
   ```bash
   pampax update /path/to/pampax-go
   ```

2. Verify indexing succeeded:
   ```bash
   pampax info
   pampax search "ordered map implementation"
   ```

3. Update `AGENTS.md` with Go-specific notes if needed

**Deliverables:**
- [ ] Index updated after code changes
- [ ] Verification via search

---

## Test Matrix (Node ↔ Go Compatibility)

| Artifact | Node → Go | Go → Node | Priority | Validation | Notes |
|---------|-----------|-----------|----------|-----------|-------|
| `.pampa/pampa.db` | ✅ | ✅ | 🔴 Critical | Schema match, embedding BLOB format | JSON arrays in BLOB, UTF-8 encoding |
| `.pampa/chunks/*.gz` | ✅ | ✅ | 🔴 Critical | SHA-1, gzip, byte-identical decompressed | Preserve BOM + line endings |
| `.pampa/chunks/*.gz.enc` | ✅ | ✅ | 🔴 Critical | PAMPAE1 magic header, HKDF, AES-GCM | 16-byte salt, 12-byte IV, 16-byte tag |
| `pampa.codemap.json` | ✅ | ✅ | 🟡 High | Top-level insertion order, field sorting | 2-space indent, final newline |
| `.pampa/merkle.json` | ✅ | ✅ | 🟢 Medium | Root hash match, semantic equivalence | Structure may differ |
| Search outputs (JSON) | ✅ | ✅ | 🔵 Low | Top-10 order, score tolerance | ±0.01 absolute or ±1% relative |

### Tolerance Thresholds

- **Vector cosine similarity:** ≥ 0.99
- **Search score:** ±0.01 absolute or ±1% relative (whichever is larger)
- **Top-10 ordering:** exact match required
- **Embedding dimensions:** exact match (length must equal `embedding_dimensions` column)

---

## Compatibility Rules Summary (from Stage 0)

### Database
- **Symbol field:** empty string `""` in DB (NOT NULL)
- **Embedding BLOB:** JSON array bytes, no whitespace
- **JSON fields:** must be valid JSON or NULL (warn & skip invalid)
- **Timestamps:** `YYYY-MM-DD HH:MM:SS` (no Z suffix)

### Chunks
- **SHA-1:** computed on raw UTF-8 bytes (preserve BOM + line endings)
- **Gzip:** default compression level
- **Encryption:** PAMPAE1 magic header, HKDF-SHA256, AES-256-GCM
- **Filenames:** `{sha}.gz` or `{sha}.gz.enc`

### Codemap
- **Symbol field:** `null` in codemap when no symbol
- **Top-level keys:** insertion order preserved (NOT alphabetically sorted)
- **Chunk metadata fields:** alphabetically sorted by `json.Marshal()`
- **Formatting:** 2-space indent, Unix newlines, final newline
- **Paths:** forward slashes `/` on all platforms
- **Omitted fields:** `symbol_parameters` omitted if empty (not `[]`)
- **Present fields:** `synonyms`, `symbol_calls`, `symbol_callers`, `symbol_neighbors` always present (may be `[]`)

### Platform-Specific
- **Line endings:** preserved exactly (no normalization)
- **BOM:** preserved if present
- **Paths:** stored with `/`, accessed with OS-native separators
- **Symlinks:** not followed during indexing
- **Case:** preserved as-is (no normalization)

---

## Success Criteria (Stage 1 Complete)

- [ ] All CLI commands (`index`, `update`, `search`, `info`) implemented
- [ ] Database schema matches Node.js byte-for-byte
- [ ] Chunks (plain + encrypted) roundtrip correctly
- [ ] Codemap JSON formatting matches Node.js exactly
- [ ] Embedding BLOB format matches Node.js
- [ ] All compatibility tests passing
- [ ] All unit tests passing
- [ ] Go can read Node artifacts without errors
- [ ] Node can read Go artifacts without errors
- [ ] Search results match Node baseline within tolerance

---

## Notes for Implementation

### Ordered Map Strategy
- Custom ordered map is required (no external dependency)
- Must preserve insertion order for top-level chunk IDs
- `json.Marshal()` will alphabetically sort fields within each chunk (correct behavior)
- Unit tests critical to verify ordering

### JSON Validation Strategy
- Use `go-playground/validator` struct tags
- Custom validation functions for JSON fields
- Log warnings with zerolog (don't fail indexing)
- Skip invalid fields and continue

### Logging Strategy
- JSON logs by default (machine-readable)
- `--pretty` flag enables console writer (human-readable)
- Log levels: debug, info, warn, error
- Structured logging for all operations

### Error Handling
- Missing chunks: return error
- Encrypted chunks without key: return error
- Invalid UTF-8: return error
- Graceful handling of filesystem errors

---

## References

**Stage 0 Artifacts:**
- `instructions/stage0_artifacts/01_sqlite_schema_contract.md`
- `instructions/stage0_artifacts/02_vector_blob_format.md`
- `instructions/stage0_artifacts/03_chunk_file_format.md`
- `instructions/stage0_artifacts/04_codemap_json_schema.md`
- `instructions/stage0_artifacts/05_edge_cases_platform_differences.md`

**Node.js Reference:**
- Repository: https://github.com/lemon07r/pampax
- Fixtures: `test/fixtures/small/`
- Baseline: `test/baselines/node_baseline_2026-01-28.json`

---

**End of Stage 1 Plan**
