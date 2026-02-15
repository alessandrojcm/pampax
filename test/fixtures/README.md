# PAMPAX Stage 0.2 - Golden Fixtures

This directory contains golden fixtures for PAMPAX Go port compatibility testing.

## Directory Structure

```
test/fixtures/
├── Dockerfile                    # Docker environment for fixture generation
├── docker-compose.yml            # Docker Compose configuration
├── setup-docker.sh               # Helper script to build Docker image
├── generate-fixtures.js          # Fixture generation script
├── README.md                     # This file
├── small/                        # Small repo fixture (~100 files, ~10K LOC)
│   ├── manifest.json             # Metadata and reproducibility info
│   ├── .pampa/                   # Database and chunks
│   │   ├── pampa.db             # SQLite database
│   │   └── chunks/              # Gzipped chunk files
│   ├── pampa.codemap.json        # Codemap JSON
│   └── search_outputs/           # Search query snapshots
│       ├── query_001.json
│       ├── query_002.json
│       └── ...
├── medium/                       # Medium repo fixture (~1K files, ~100K LOC)
│   └── (same structure)
└── large/                        # Large repo fixture (~10K files, ~1M LOC)
    └── (same structure)
```

## Prerequisites

- Docker
- Docker Compose
- Git (for cloning test repositories)

## Quick Start

### 1. Build Docker Environment

```bash
cd test/fixtures
./setup-docker.sh
```

This builds a Docker image with:
- Node.js 22 LTS
- All native dependencies (sqlite3, tree-sitter, etc.)
- PAMPAX source code and dependencies

### 2. Generate Fixtures

#### Option A: Using the PAMPAX repository itself (small fixture)

```bash
docker-compose run --rm pampax-fixture-gen \
  node test/fixtures/generate-fixtures.js /pampax small
```

This indexes the PAMPAX repository itself as a small fixture.

#### Option B: Using an external repository

```bash
# Clone a test repository
git clone https://github.com/example/repo ./test-repos/my-repo

# Generate fixture
docker-compose run --rm pampax-fixture-gen \
  node test/fixtures/generate-fixtures.js /test-repos/my-repo medium
```

#### Option C: Interactive mode

```bash
# Start interactive shell
docker-compose run --rm pampax-fixture-gen /bin/bash

# Inside container, run fixture generation
node test/fixtures/generate-fixtures.js /pampax small
node test/fixtures/generate-fixtures.js /test-repos/medium-repo medium
node test/fixtures/generate-fixtures.js /test-repos/large-repo large

# Exit when done
exit
```

### 3. Verify Fixtures

After generation, verify the fixtures were created:

```bash
ls -lh test/fixtures/small/
ls -lh test/fixtures/small/.pampa/
ls -lh test/fixtures/small/search_outputs/
cat test/fixtures/small/manifest.json
```

## Fixture Requirements

### Small Fixture
- **Size:** ~100 files, ~10K LOC
- **Purpose:** Fast validation, smoke tests
- **Examples:** Small libraries, utilities, or PAMPAX itself

### Medium Fixture
- **Size:** ~1K files, ~100K LOC
- **Purpose:** Comprehensive testing, performance benchmarks
- **Examples:** Medium-sized applications, frameworks

### Large Fixture
- **Size:** ~10K files, ~1M LOC
- **Purpose:** Stress testing, scalability validation
- **Examples:** Large monorepos, enterprise applications

## Fixture Contents

Each fixture contains:

### 1. manifest.json
Metadata for reproducibility:
- PAMPAX version
- Node version
- System specifications
- Git commit hash
- Embedding provider/model
- Repository statistics

### 2. .pampa/pampa.db
SQLite database with:
- `code_chunks` table with embeddings
- All indices and constraints
- Metadata about indexing

### 3. .pampa/chunks/
Gzipped chunk files:
- Named by SHA-1 hash
- Contain raw source code
- Compressed with gzip

### 4. pampa.codemap.json
Codemap structure with:
- File paths and metadata
- Chunk IDs and relationships
- Alphabetically sorted keys

### 5. search_outputs/
Search query snapshots:
- Benchmark queries from `test/benchmarks/fixtures/queries.js`
- Top-10 results with scores
- Used for validating search output compatibility

## Usage in Tests

### Node.js Validation Tests

```javascript
import { validateFixture } from './validate-fixtures.js';

// Load and validate a fixture
const fixture = await loadFixture('small');
await validateFixture(fixture);
```

### Go Port Compatibility Tests

```go
// Load Node-generated fixture
fixture := LoadFixture("small")

// Index same repository with Go implementation
goResults := IndexRepository(fixture.RepoPath)

// Compare artifacts
CompareDatabase(fixture.DB, goResults.DB)
CompareChunks(fixture.Chunks, goResults.Chunks)
CompareCodemap(fixture.Codemap, goResults.Codemap)
CompareSearchResults(fixture.SearchOutputs, goResults.SearchOutputs)
```

## Fixture Validation

Fixtures should be validated to ensure:

1. **Database schema matches specification**
   - Tables, columns, types correct
   - Indices present
   - Constraints enforced

2. **Chunks are valid**
   - Files exist and are gzipped
   - SHA-1 hashes match filenames
   - Content can be decompressed

3. **Codemap is valid**
   - JSON is well-formed
   - Keys are alphabetically sorted
   - Required fields present

4. **Search outputs are valid**
   - Results match expected format
   - Scores are reasonable
   - No errors in responses

## Regenerating Fixtures

Fixtures should be regenerated when:

- PAMPAX version changes significantly
- Embedding model changes
- Schema or format changes
- New test scenarios are needed

To regenerate:

```bash
# Remove old fixture
rm -rf test/fixtures/small

# Generate new fixture
docker-compose run --rm pampax-fixture-gen \
  node test/fixtures/generate-fixtures.js /pampax small
```

## Troubleshooting

### Docker build fails

```bash
# Clean Docker build cache
docker-compose build --no-cache
```

### Native modules fail to compile

This should not happen inside Docker, but if it does:
- Check Docker has enough memory allocated
- Verify Node version is 22.x
- Check build tools are installed

### Fixture generation hangs

- Increase Docker memory limit
- Check if repository is too large
- Monitor disk space

### Search queries fail

- Verify embeddings were generated correctly
- Check database is not corrupted
- Ensure local transformer model downloaded

## Notes

- **Deterministic builds:** Use `transformers` (local) provider for reproducibility
- **Portability:** Fixtures are platform-independent (paths normalized to forward slashes)
- **Version control:** Fixtures are stored in the repository for easy access
- **Size limits:** Keep fixtures reasonably sized (< 100MB each) for git

## Related Documentation

- [GO_PORT_STAGE0_DETAILS.md](../../instructions/GO_PORT_STAGE0_DETAILS.md) - Overall Stage 0 plan
- [stage0_artifacts/](../../instructions/stage0_artifacts/) - Artifact specifications
- [test/benchmarks/](../benchmarks/) - Benchmark suite
