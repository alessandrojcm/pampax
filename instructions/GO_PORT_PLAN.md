# Go Port Migration Plan

**Goal:** Port PAMPAX to Go for performance (≥2× indexing speed) and cross-platform binary distribution, while maintaining full compatibility with existing Node.js implementation.

## Requirements Summary
- **Performance:** ≥2× indexing speed improvement
- **Compatibility:** Exact SQLite schema and data format compatibility
- **Distribution:** Single cross-platform binary with model download on first run
- **Features:** Full feature parity (embeddings, reranker, tree-sitter, symbol boost, hybrid search)
- **Configuration:** Add global config support (Viper) while maintaining env-var compatibility
- **Testing:** Parity tests against Node.js golden fixtures

## Technical Decisions
- **Tree-sitter:** go-tree-sitter (or best available Go binding)
- **Embeddings:** Local in-binary inference (ONNX/ggml/gguf)
- **Reranker:** Local in-binary cross-encoder
- **Models:** Downloaded on first run, stored in cache
- **Concurrency:** Aggressive parallelism for performance (internals need not match Node)
- **Licensing:** Permissive models preferred

---

## Stage 0 — Compatibility Contract (serial)
**Goals:** Establish exact compatibility requirements and performance baseline.

**Deliverables:**
- [ ] Document current SQLite schema with exact data types and constraints
- [ ] Document vector blob format and serialization
- [ ] Capture golden fixtures:
  - [ ] `.pampa/` database snapshot
  - [ ] Chunk files from sample repos
  - [ ] `pampa.codemap.json` outputs
  - [ ] Search query outputs (top-K results with scores)
- [ ] Define representative test repos (small/medium/large)
- [ ] Measure baseline performance metrics:
  - [ ] Indexing throughput (files/sec, chunks/sec)
  - [ ] Memory usage during indexing
  - [ ] Search latency (p50, p95, p99)
  - [ ] Disk usage

**Success Criteria:**
- Complete schema documentation with examples
- Golden fixtures from ≥3 representative repos
- Baseline metrics documented with reproducible benchmarks

---

## Stage 1 — Storage & Schema Layer (serial)
**Goals:** Implement Go data layer with exact schema compatibility.

**Deliverables:**
- [ ] SQLite wrapper with connection pooling
- [ ] Schema creation matching Node implementation exactly
- [ ] Vector serialization/deserialization (exact binary format)
- [ ] Chunk file I/O (read/write compatibility)
- [ ] Codemap JSON serialization
- [ ] Database migration utilities (if needed)

**Testing:**
- [ ] Round-trip tests: write from Go, read from Node
- [ ] Round-trip tests: write from Node, read from Go
- [ ] Binary comparison of vector blobs
- [ ] Schema validation against golden fixtures

**Success Criteria:**
- All compatibility tests pass
- No data loss in round-trip conversions
- Byte-for-byte vector compatibility verified

---

## Stage 2 — File Discovery & Ignore Rules (serial)
**Goals:** Match file inclusion/exclusion behavior exactly.

**Deliverables:**
- [ ] File walker with concurrent directory traversal
- [ ] Default ignore patterns matching Node implementation
- [ ] `.pampignore` parser and evaluator
- [ ] `.gitignore` parser and evaluator
- [ ] Symlink handling (cross-platform)
- [ ] Hidden file detection (cross-platform)

**Testing:**
- [ ] Parity tests: identical file lists vs Node for same repos
- [ ] Edge cases: symlinks, hidden files, permissions
- [ ] Cross-platform tests (Windows/macOS/Linux)
- [ ] Performance benchmarks vs Node walker

**Success Criteria:**
- File lists match Node output exactly on test repos
- Performance ≥1.5× faster than Node walker
- All edge cases handled correctly

---

## Stage 3 — Chunking + Metadata (serial)
**Goals:** Produce equivalent chunks and metadata.

**Deliverables:**
- [ ] Chunking algorithm (same size/overlap logic)
- [ ] Language detection (match Node results)
- [ ] File path normalization (cross-platform)
- [ ] Metadata extraction (language, tags, paths)
- [ ] Hash computation (for deduplication)

**Testing:**
- [ ] Chunk boundary parity tests
- [ ] Metadata field-by-field comparison
- [ ] Language detection accuracy tests
- [ ] Hash collision tests

**Success Criteria:**
- Chunk boundaries match Node output
- Metadata fields identical
- Language detection matches ≥99% of Node results

---

## Stage 4 — Tree-sitter & Symbol Extraction (can parallelize with Stage 5)
**Goals:** Preserve symbol boost and call-graph features.

**Deliverables:**
- [ ] Tree-sitter Go bindings integration
- [ ] Parser registry for supported languages
- [ ] Symbol extraction (functions, classes, methods, etc.)
- [ ] Call graph construction
- [ ] Symbol signature generation
- [ ] Fallback behavior for unsupported languages

**Testing:**
- [ ] Symbol extraction parity tests per language
- [ ] Call graph correctness tests
- [ ] Edge cases: nested symbols, generics, macros
- [ ] Performance benchmarks vs Node

**Success Criteria:**
- Symbol extraction matches Node for supported languages
- Graceful degradation for unsupported languages
- Performance ≥2× faster than Node tree-sitter

---

## Stage 5 — Embeddings (can parallelize with Stage 4)
**Goals:** Local embeddings inside Go binary with model download.

**Deliverables:**
- [ ] Choose inference backend (ONNX Runtime Go / ggml / gguf)
- [ ] Model downloader with progress reporting
- [ ] Model cache management
- [ ] Embedding pipeline (tokenization → inference → normalization)
- [ ] Batch processing for efficiency
- [ ] API client fallback (OpenAI, Ollama, Cohere)

**Testing:**
- [ ] Embedding parity tests (cosine similarity with Node outputs)
- [ ] Tolerance definition (acceptable numeric deviation)
- [ ] Performance benchmarks (throughput, latency, memory)
- [ ] Model download/cache tests

**Success Criteria:**
- Embeddings within acceptable tolerance of Node outputs
- Performance ≥1.5× faster than Node (local mode)
- Model download works on all platforms

**Open Questions:**
- **Model format:** ONNX vs ggml/gguf? (ONNX has better Go support)
- **Tolerance:** What cosine similarity threshold is acceptable? (suggest ≥0.99)

---

## Stage 6 — Search Stack (serial, depends on Stages 4 & 5)
**Goals:** Match hybrid search, symbol boost, and reranker.

**Deliverables:**
- [ ] BM25 index construction
- [ ] Vector similarity search
- [ ] Reciprocal Rank Fusion (RRF) implementation
- [ ] Symbol boost scoring
- [ ] Local reranker integration (cross-encoder)
- [ ] Search filters (path_glob, lang, tags, context packs)
- [ ] Result ranking and deduplication

**Testing:**
- [ ] Search parity tests on golden fixtures
- [ ] Top-K ordering comparison (with tolerance)
- [ ] Scoring function validation
- [ ] Performance benchmarks (latency at various K values)

**Success Criteria:**
- Top-10 results match Node ≥90% of the time
- Search latency ≤50% of Node latency
- Reranker produces similar score distributions

**Open Questions:**
- **Top-K tolerance:** How much ordering variation is acceptable?
- **Reranker model:** Use same model as Node or optimize for Go?

---

## Stage 7 — CLI & MCP Parity (serial, depends on all previous stages)
**Goals:** Provide JSON-first CLI with parity to CLI-sensible MCP tools, plus agent-friendly structured output and schema validation.

**Deliverables:**

### CLI Surface & Parity
- [ ] CLI framework (cobra or similar)
- [ ] Commands: `index`, `update`, `watch`, `search`, `info`, `context`, `chunk get`
- [ ] `chunk get <sha>` retrieves chunk by SHA (obtained from `search` results)
- [ ] `chunk get` returns same payload as MCP `get_code_at_chunk` (minus MCP transport wrapper)
- [ ] Flag parsing matching Node CLI where applicable

### CLI ↔ MCP Mapping Table
Document explicit mapping between CLI commands and MCP tools:

| CLI Command | MCP Tool(s) | Notes |
|-------------|-------------|-------|
| `pampax index` | `index_project` | Full indexing |
| `pampax update` | `update_project` | Incremental update |
| `pampax watch` | N/A | CLI-specific (file watcher) |
| `pampax search` | `search_code` | Semantic + hybrid search |
| `pampax info` | `get_project_stats` | Project statistics |
| `pampax context list` | `list_context_packs` | List context packs |
| `pampax context show` | `show_context_pack` | Show pack contents |
| `pampax context use` | `use_context_pack` | Apply context pack |
| `pampax chunk get` | `get_code_at_chunk` | Retrieve chunk by SHA |

**MCP tools NOT exposed to CLI:** None (all CLI-sensible tools are exposed)

### JSON-First Output
- [ ] JSON output is **default** for all CLI commands (no human-formatted output required)
- [ ] Structured JSON error envelope with enum error codes
- [ ] Error codes: `INVALID_INPUT`, `NOT_FOUND`, `INDEX_MISSING`, `DB_ERROR`, `IO_ERROR`, `CONFIG_ERROR`, `EMBEDDING_ERROR`, `SEARCH_ERROR`, `INTERNAL_ERROR`
- [ ] `watch` streams JSON objects continuously as changes are detected (one object per event)

### JSON Schema Definitions
- [ ] Define JSON schemas for each CLI command output
- [ ] Store schemas in `instructions/` directory
- [ ] Schema includes:
  - [ ] Output payload structure and required fields
  - [ ] Error envelope structure
  - [ ] Field types and constraints
- [ ] Example schema structure:
  ```json
  {
    "command": "search",
    "output": {
      "results": [...],
      "total": 10,
      "query": "..."
    }
  }
  ```
  ```json
  {
    "error": {
      "code": "INDEX_MISSING",
      "message": "No index found at path",
      "hint": "Run 'pampax index' first"
    }
  }
  ```

### MCP Server
- [ ] MCP server implementation (stdio transport)
- [ ] MCP tool handlers matching Node behavior
- [ ] Progress reporting during indexing

**Testing:**
- [ ] CLI integration tests (end-to-end)
- [ ] JSON schema validation tests for each command output
- [ ] Golden JSON fixture tests for stable outputs
- [ ] Error envelope tests for common failure paths
- [ ] CLI/MCP mapping tests (verify parity coverage)
- [ ] Watch output tests (streaming JSON objects on change events)
- [ ] MCP protocol compliance tests
- [ ] Cross-platform CLI tests
- [ ] `chunk get` payload parity tests vs MCP

**Success Criteria:**
- All CLI outputs validate against defined JSON schemas
- CLI command set matches mapping table (all CLI-sensible MCP tools exposed)
- `chunk get` returns same payload as MCP `get_code_at_chunk` (minus transport wrapper)
- JSON output is default for all CLI commands
- Errors always match structured JSON envelope with enum codes
- MCP protocol passes all compatibility tests
- `watch` correctly streams JSON objects on file changes

---

## Stage 8 — Global Config (serial)
**Goals:** Introduce Viper-based config without breaking env-vars.

**Deliverables:**
- [ ] Config file format (TOML recommended)
- [ ] Config search paths (cross-platform):
  - [ ] `./pampax.toml` (project-local)
  - [ ] `~/.config/pampax/config.toml` (user-global)
  - [ ] `~/.pampax.toml` (legacy fallback)
- [ ] Viper integration with precedence:
  1. CLI flags
  2. Environment variables
  3. Config file (project-local)
  4. Config file (user-global)
  5. Defaults
- [ ] Config validation and schema
- [ ] Config file generation (`pampax config init`)

**Testing:**
- [ ] Precedence tests (flag > env > file)
- [ ] Cross-platform config path tests
- [ ] Backward compatibility tests (env-only usage)

**Success Criteria:**
- Env-var usage still works unchanged
- Config files provide better UX for non-MCP usage
- Clear documentation for config options

**Open Questions:**
- **Format:** TOML or YAML? (TOML recommended for simplicity)

---

## Stage 9 — Performance Optimization (serial, final stage)
**Goals:** Achieve ≥2× indexing speed improvement.

**Deliverables:**
- [ ] Profiling report (CPU, memory, I/O)
- [ ] Concurrency tuning:
  - [ ] File scanning parallelism
  - [ ] Chunk processing parallelism
  - [ ] Embedding batch processing
  - [ ] Database write batching
- [ ] Memory optimizations:
  - [ ] Streaming file reads
  - [ ] Bounded queues
  - [ ] Memory pooling
- [ ] I/O optimizations:
  - [ ] Batch database inserts
  - [ ] Async file operations
- [ ] Benchmark suite comparing Go vs Node

**Testing:**
- [ ] Performance regression tests
- [ ] Memory leak tests
- [ ] Stress tests (large repos, many files)
- [ ] Comparative benchmarks (Go vs Node)

**Success Criteria:**
- ≥2× indexing speed improvement over Node
- Memory usage ≤1.5× Node (acceptable for performance gain)
- No performance regressions in search

**Open Questions:**
- **Representative repo size:** What's "large" for benchmarks? (suggest 10k+ files)

---

## Testing Strategy

### Parity Testing
- **Golden Fixtures:** Generate from Node implementation on representative repos
- **Comparison Points:**
  - File discovery lists
  - Chunk boundaries and content
  - Symbol extraction results
  - Embedding vectors (with tolerance)
  - Search results (top-K with ordering tolerance)
  - Database schema and content
  - Codemap JSON structure

### Performance Testing
- **Baseline Repos:**
  - Small: ~100 files, ~10K LOC
  - Medium: ~1K files, ~100K LOC
  - Large: ~10K files, ~1M LOC
- **Metrics:**
  - Indexing throughput (files/sec, MB/sec)
  - Search latency (p50, p95, p99)
  - Memory usage (peak, average)
  - Disk usage

### Cross-Platform Testing
- **Platforms:** macOS, Linux, Windows
- **Test Points:**
  - File path handling
  - Ignore pattern matching
  - Symlink behavior
  - Config file locations
  - Binary distribution

---

## Edge Cases & Constraints

### Large Monorepos
- **Challenge:** Indexing time, memory usage, database size
- **Mitigation:** Streaming processing, batch operations, progress reporting

### Binary Files
- **Challenge:** Wasted processing on non-text files
- **Mitigation:** Early detection and skip (match Node behavior)

### Unsupported Languages
- **Challenge:** Limited tree-sitter grammar coverage
- **Mitigation:** Graceful fallback, text-only chunking

### Windows-Specific Issues
- **Challenge:** Path separators, symlinks, permissions
- **Mitigation:** Cross-platform testing, path normalization

### Model Compatibility
- **Challenge:** Different models may produce incompatible embeddings
- **Mitigation:** Document model requirements, version checks

### Frequent File Churn
- **Challenge:** Watch mode performance with many changes
- **Mitigation:** Debouncing, incremental updates, Merkle tree optimization

---

## Open Questions Summary

1. **Embedding inference backend:** ONNX Runtime (simpler Go integration) vs ggml/gguf (better performance)?
2. **Chunk compatibility:** Byte-for-byte identical or logically compatible?
3. **Embedding tolerance:** What cosine similarity threshold is acceptable for parity? (suggest ≥0.99)
4. **Top-K ordering tolerance:** How much variation in search result order is acceptable? (suggest top-5 must match)
5. **Representative repo size:** What defines "large" for benchmarks? (suggest 10K+ files)
6. **Config file format:** TOML or YAML? (TOML recommended)
7. **Model licensing:** Any specific restrictions or preferences?

---

## Success Metrics

### Performance Goals
- ✅ Indexing speed: ≥2× faster than Node
- ✅ Search latency: ≤50% of Node latency
- ✅ Memory usage: ≤1.5× Node (acceptable trade-off)

### Compatibility Goals
- ✅ Schema: 100% identical
- ✅ File discovery: 100% identical file lists
- ✅ Embeddings: ≥0.99 cosine similarity
- ✅ Search top-10: ≥90% overlap with Node results

### Quality Goals
- ✅ Cross-platform: Works on macOS, Linux, Windows
- ✅ Single binary: No runtime dependencies (except models)
- ✅ Model download: Automatic on first run
- ✅ Backward compatible: Can read existing `.pampa/` data

---

## Next Steps

1. **Answer open questions** (see above)
2. **Execute Stage 0** (establish baseline and golden fixtures)
3. **Set up Go project structure:**
   - Module initialization
   - CI/CD pipeline
   - Testing framework
   - Benchmark infrastructure
4. **Begin Stage 1** (storage layer implementation)

---

## Notes

- **Incremental approach:** Each stage builds on previous stages
- **Parallelization:** Stages 4 & 5 can be developed concurrently
- **Testing first:** Write parity tests before implementation
- **Documentation:** Update README as Go features are completed
- **Backward compatibility:** Maintain during entire development process
