# Stage 0 - Compatibility Contract (Complete)

**Status:** ✅ 100% Complete  
**Date:** 2026-01-28  
**Parent Plan:** [GO_PORT_STAGE0_DETAILS.md](../GO_PORT_STAGE0_DETAILS.md)

---

## Purpose

This directory is the **canonical reference** for Stage 0 (Compatibility Contract) of the PAMPAX Go port. It defines the artifact formats and cross-platform behaviors needed to ensure Node ↔ Go interoperability.

---

## Required Reading (Stage 1 Engineers)

Read these documents in order:

1. **Database schema** – [01_sqlite_schema_contract.md](./01_sqlite_schema_contract.md)
2. **Embedding storage** – [02_vector_blob_format.md](./02_vector_blob_format.md)
3. **Chunk files** – [03_chunk_file_format.md](./03_chunk_file_format.md)
4. **Codemap schema** – [04_codemap_json_schema.md](./04_codemap_json_schema.md)
5. **Cross‑platform edge cases** – [05_edge_cases_platform_differences.md](./05_edge_cases_platform_differences.md)

---

## Compatibility & Acceptance Criteria

The Go implementation is considered compatible only if all critical artifacts are interchangeable with Node.js.

### Interoperability Matrix

| Artifact | Node → Go | Go → Node | Priority |
|----------|-----------|-----------|----------|
| `.pampa/pampa.db` | ✅ | ✅ | 🔴 Critical |
| `.pampa/chunks/*.gz` | ✅ | ✅ | 🔴 Critical |
| `pampa.codemap.json` | ✅ | ✅ | 🟡 High |
| `.pampa/merkle.json` | ✅ | ✅ | 🟢 Medium |
| Search outputs (JSON) | ✅ | ✅ | 🔵 Low |

### Acceptance Rules (Summary)

**Database (`.pampa/pampa.db`)** – Strict
- Schema must match exactly (tables, columns, indices, constraints).
- Embedding BLOBs must be JSON arrays stored as UTF‑8 bytes.

**Chunks (`.pampa/chunks/*.gz`)** – Strict
- SHA‑1 filenames must match **decompressed** content hash.
- Decompressed content must be byte‑identical to original source.

**Codemap (`pampa.codemap.json`)** – Strict with normalization
- Must match schema in [04_codemap_json_schema.md](./04_codemap_json_schema.md).
- **Top‑level keys preserve insertion order** (no alphabetical sorting).

**Merkle (`.pampa/merkle.json`)** – Flexible
- Root hash must match; structure must be semantically equivalent.

**Search outputs** – Strict for top‑10
- Top‑10 order must match exactly; scores within tolerance.

### Tolerances

- **Vector cosine similarity:** ≥ 0.99
- **Search score:** ±0.01 absolute or ±1% relative (whichever is larger)
- **Top‑10 ordering:** exact match required

---

## Validation Inputs (High‑Level)

Use the golden fixture under `test/fixtures/small/` and the benchmark baseline in `test/baselines/node_baseline_2026-01-28.json` as the canonical Node.js references for compatibility checks. These are the primary artifacts for validating the Go implementation.

---

## Next Steps (Stage 1)

- Implement the Go indexing pipeline using the five canonical specs above.
- Validate outputs against the Stage 0 fixture and baseline artifacts.
- Ensure all compatibility criteria in this README are met before declaring parity.

---

**Status:** ✅ **Stage 0 100% Complete** — Ready for Stage 1 (Go Implementation)
