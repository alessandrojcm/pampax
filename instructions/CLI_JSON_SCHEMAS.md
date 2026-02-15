# CLI JSON Output Schemas

This document defines the JSON output schemas for all PAMPAX CLI commands in the Go port.

## Design Principles

1. **JSON by default**: All CLI commands output JSON (no human-formatted text required)
2. **Agent-friendly**: Structured, parseable output suitable for jq and programmatic consumption
3. **Consistent error handling**: All errors use the same envelope structure
4. **No schema versioning**: Keep output simple and noise-free (versioning can be added later if needed)

---

## Error Envelope

All CLI errors use this structured format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "hint": "Actionable suggestion for fixing the error"
  }
}
```

### Error Codes (Enum)

| Code | Description |
|------|-------------|
| `INVALID_INPUT` | Invalid command arguments or flags |
| `NOT_FOUND` | Requested resource not found (file, chunk, context pack) |
| `INDEX_MISSING` | No index exists at the specified path |
| `DB_ERROR` | Database operation failed |
| `IO_ERROR` | File system operation failed |
| `CONFIG_ERROR` | Configuration error (invalid config file, missing env var) |
| `EMBEDDING_ERROR` | Embedding generation failed |
| `SEARCH_ERROR` | Search operation failed |
| `INTERNAL_ERROR` | Unexpected internal error |

---

## Command Schemas

### `pampax index`

Indexes a project directory.

**Success output:**
```json
{
  "status": "success",
  "indexed": {
    "files": 150,
    "chunks": 4500,
    "duration_ms": 12340
  },
  "path": "/absolute/path/to/project"
}
```

**Error example:**
```json
{
  "error": {
    "code": "IO_ERROR",
    "message": "Cannot read directory: /path/to/project",
    "hint": "Check that the path exists and you have read permissions"
  }
}
```

---

### `pampax update`

Updates an existing index.

**Success output:**
```json
{
  "status": "success",
  "updated": {
    "files_added": 5,
    "files_modified": 12,
    "files_removed": 3,
    "chunks_added": 150,
    "chunks_removed": 90,
    "duration_ms": 3200
  },
  "path": "/absolute/path/to/project"
}
```

**Error example:**
```json
{
  "error": {
    "code": "INDEX_MISSING",
    "message": "No index found at /path/to/project",
    "hint": "Run 'pampax index' first to create the index"
  }
}
```

---

### `pampax watch`

Watches for file changes and streams update events.

**Output format:** Continuous stream of JSON objects (one per event)

**Event types:**

**File change event:**
```json
{
  "event": "file_changed",
  "path": "src/main.go",
  "change_type": "modified",
  "timestamp": "2026-01-27T10:30:45Z",
  "chunks_updated": 5
}
```

**Index update event:**
```json
{
  "event": "index_updated",
  "timestamp": "2026-01-27T10:30:46Z",
  "changes": {
    "files_modified": 1,
    "chunks_added": 3,
    "chunks_removed": 2
  }
}
```

**Error event:**
```json
{
  "event": "error",
  "error": {
    "code": "IO_ERROR",
    "message": "Cannot watch path: /path/to/project",
    "hint": "Check that the path exists and you have read permissions"
  }
}
```

---

### `pampax search <query>`

Performs hybrid semantic + BM25 search.

**Success output:**
```json
{
  "query": "user authentication logic",
  "results": [
    {
      "chunk_sha": "a1b2c3d4e5f6...",
      "file_path": "src/auth/login.go",
      "language": "go",
      "content": "func AuthenticateUser(username, password string) {...}",
      "score": 0.95,
      "start_line": 42,
      "end_line": 58,
      "symbols": ["AuthenticateUser"],
      "tags": ["authentication", "security"]
    }
  ],
  "total": 15,
  "filters": {
    "path_glob": ["src/**"],
    "lang": ["go"],
    "tags": null
  }
}
```

**Error example:**
```json
{
  "error": {
    "code": "INDEX_MISSING",
    "message": "No index found at current directory",
    "hint": "Run 'pampax index' first"
  }
}
```

---

### `pampax info`

Displays project statistics.

**Success output:**
```json
{
  "project": {
    "path": "/absolute/path/to/project",
    "indexed_at": "2026-01-27T09:15:30Z",
    "last_updated": "2026-01-27T10:30:45Z"
  },
  "stats": {
    "total_files": 150,
    "total_chunks": 4500,
    "total_size_bytes": 1048576,
    "languages": {
      "go": 100,
      "javascript": 30,
      "typescript": 20
    },
    "database_size_bytes": 52428800
  }
}
```

**Error example:**
```json
{
  "error": {
    "code": "INDEX_MISSING",
    "message": "No index found at /path/to/project",
    "hint": "Run 'pampax index' first"
  }
}
```

---

### `pampax context list`

Lists available context packs.

**Success output:**
```json
{
  "context_packs": [
    {
      "name": "stripe-backend",
      "description": "Stripe payment integration backend code",
      "chunks": 45,
      "created_at": "2026-01-20T14:30:00Z"
    },
    {
      "name": "auth-frontend",
      "description": "Frontend authentication components",
      "chunks": 23,
      "created_at": "2026-01-22T09:15:00Z"
    }
  ],
  "total": 2
}
```

---

### `pampax context show <name>`

Shows details of a specific context pack.

**Success output:**
```json
{
  "name": "stripe-backend",
  "description": "Stripe payment integration backend code",
  "created_at": "2026-01-20T14:30:00Z",
  "chunks": [
    {
      "chunk_sha": "a1b2c3d4e5f6...",
      "file_path": "src/payments/stripe.go",
      "language": "go",
      "start_line": 10,
      "end_line": 45
    }
  ],
  "total_chunks": 45
}
```

**Error example:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Context pack 'stripe-backend' not found",
    "hint": "Use 'pampax context list' to see available packs"
  }
}
```

---

### `pampax context use <name>`

Applies a context pack to the current search context.

**Success output:**
```json
{
  "status": "success",
  "context_pack": "stripe-backend",
  "chunks_loaded": 45,
  "message": "Context pack 'stripe-backend' is now active"
}
```

**Error example:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Context pack 'stripe-backend' not found",
    "hint": "Use 'pampax context list' to see available packs"
  }
}
```

---

### `pampax chunk get <sha>`

Retrieves a specific chunk by its SHA hash.

**Success output:**
```json
{
  "chunk_sha": "a1b2c3d4e5f6...",
  "file_path": "src/auth/login.go",
  "language": "go",
  "content": "func AuthenticateUser(username, password string) (User, error) {\n    // Implementation\n}",
  "start_line": 42,
  "end_line": 58,
  "symbols": ["AuthenticateUser"],
  "tags": ["authentication", "security"],
  "metadata": {
    "file_size_bytes": 2048,
    "indexed_at": "2026-01-27T09:15:30Z"
  }
}
```

**Error example:**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Chunk with SHA 'a1b2c3d4e5f6...' not found",
    "hint": "Use 'pampax search' to find valid chunk SHAs"
  }
}
```

---

## Implementation Notes

1. **Consistency**: All commands follow the same error envelope structure
2. **Timestamps**: Use ISO 8601 format with timezone (`2026-01-27T10:30:45Z`)
3. **Paths**: Always use absolute paths in output
4. **File sizes**: Report in bytes (consumers can format as needed)
5. **Durations**: Report in milliseconds for precision
6. **Watch mode**: Stream one JSON object per line (newline-delimited JSON)
7. **Chunk SHA**: Use full SHA hash (not truncated) for unambiguous identification

---

## Testing Requirements

Each command schema must have:
1. **Schema validation tests**: JSON output validates against schema
2. **Golden fixture tests**: Stable output snapshots for regression testing
3. **Error envelope tests**: All error paths produce correct error codes
4. **Cross-platform tests**: Output format consistent across macOS/Linux/Windows

---

## Future Considerations

- **Pagination**: For commands returning large result sets (defer until needed)
- **Filtering**: Additional filter options in search (already supported via flags)
- **Streaming search**: For very large result sets (defer until needed)
- **Schema versioning**: Add if backward compatibility becomes a concern
