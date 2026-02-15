# PAMPAX Go Port

This is the Go implementation of PAMPAX (Pragmatic Agentic Memory via Portable Artifact eXchange).

## Project Structure

```
github.com/alessandrojcm/pampax-go/
├── cmd/
│   └── pampax/              # CLI entrypoint (Cobra root)
├── internal/
│   ├── app/                 # Command orchestration (index/search/update/info)
│   ├── config/              # Viper config + env parsing + defaults
│   ├── db/                  # sqlc queries + DB access layer
│   ├── migrations/          # dbmate migration files
│   ├── chunks/              # SHA-1, gzip, encryption, atomic writes
│   ├── codemap/             # Ordered map, normalization, JSON serialization
│   ├── indexer/             # File discovery, chunking, language detection
│   ├── providers/           # Embedding provider interfaces + stubs
│   ├── search/              # Cosine + BM25/hybrid
│   ├── merkle/              # Merkle tree generation
│   ├── compat/              # Node/Go compatibility helpers
│   └── utils/               # Path, UTF-8, timestamps, errors
├── sql/
│   ├── schema.sql           # DB schema for dbmate
│   └── queries/             # sqlc query files
└── test/
    ├── fixtures/            # Golden fixtures (pre-generated from Node)
    ├── compat/              # Cross-implementation compatibility tests
    └── unit/                # Unit tests
```

## Technology Stack

- **Go Version:** 1.25+ (module: go 1.25.5)
- **SQLite:** `modernc.org/sqlite` (pure Go, no CGO)
- **Query Generation:** `github.com/sqlc-dev/sqlc`
- **Migrations:** `github.com/amacneil/dbmate`
- **CLI Framework:** `github.com/spf13/cobra`
- **Config Management:** `github.com/spf13/viper`
- **JSON Validation:** `github.com/go-playground/validator/v10`
- **Logging:** `github.com/rs/zerolog`
- **Crypto:** `golang.org/x/crypto/hkdf`

## Development Status

**Stage 1A - Project Skeleton:** ✅ Complete
- [x] Directory structure created
- [x] go.mod initialized
- [x] sqlc.yaml configured
- [x] Dependencies installed
- [x] .gitignore updated

**Stage 1B - Tooling Configuration:** ✅ Complete
- [x] sqlc CLI tool installed (v1.30.0)
- [x] dbmate CLI tool installed (v2.29.5)
- [x] dbmate configuration created (.env file)
- [x] Makefile with development commands

**Next:** Stage 2A - Codemap Ordered Map & Serialization

## Installation

### Prerequisites

1. Go 1.22 or later
2. sqlc CLI tool:
   ```bash
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```
3. dbmate CLI tool:
   ```bash
   go install github.com/amacneil/dbmate/v2@latest
   ```

### Build

```bash
cd go-port
go build -o pampax ./cmd/pampax
```

## Development

### Generate SQL Code (sqlc)

```bash
cd internal/db
sqlc generate
```

### Run Migrations (dbmate)

```bash
export DATABASE_URL="sqlite:.pampa/pampa.db"
dbmate up
```

### Run Tests

```bash
go test ./...
```

## Implementation Plan

See `../instructions/GO_PORT_STAGE1_PLAN.md` for detailed implementation stages and compatibility requirements.

## Compatibility

This Go implementation maintains byte-level compatibility with the Node.js implementation for all `.pampa/` artifacts:
- Database schema (SQLite)
- Chunk files (gzip + optional AES-256-GCM encryption)
- Codemap JSON structure (insertion order preserved)
- Vector embeddings (JSON array BLOB format)

## License

Same as PAMPAX main repository.
