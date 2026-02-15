# PAMPAX Performance Baselines

This directory contains baseline performance metrics for PAMPAX search functionality. These baselines establish reference points for:

1. **Cross-version compatibility** - Ensure new versions maintain or improve performance
2. **Cross-platform validation** - Compare Node.js vs Go port implementations
3. **Regression detection** - Identify performance degradation in new releases

## Baseline Files

### `node_baseline_2026-01-28.json`

**Purpose:** Node.js v22 reference baseline for Go port Stage 0 compatibility testing

**Key Metrics:**
- **Base (vector-only):** P@1: 0.750, MRR@5: 0.750, nDCG@10: 0.829
- **Hybrid (BM25+vector):** P@1: 0.750, MRR@5: 0.800, nDCG@10: 0.847
- **Hybrid+CE (with reranker):** P@1: 0.750, MRR@5: 0.875, nDCG@10: 0.908

**Environment:**
- PAMPAX: v1.17.1
- Node.js: v22.22.0
- Platform: Docker (Debian 11, aarch64)
- CPU: 10 cores, 7.7GB RAM

**Test Data:**
- Synthetic fixtures (`test/benchmarks/fixtures/`)
- Mock embeddings (4 dimensions)
- TestProvider (deterministic)

## Metric Definitions

### Precision@1 (P@1)
**What:** Percentage of queries where the top result is relevant  
**Range:** 0.0 to 1.0 (higher is better)  
**Interpretation:** Measures accuracy of the single best result

### Mean Reciprocal Rank@5 (MRR@5)
**What:** Average of `1 / rank_of_first_relevant_result` (top 5 only)  
**Range:** 0.0 to 1.0 (higher is better)  
**Interpretation:** Measures how quickly users find a relevant result

### Normalized Discounted Cumulative Gain@10 (nDCG@10)
**What:** Ranking quality metric that rewards relevant results appearing earlier  
**Range:** 0.0 to 1.0 (higher is better)  
**Interpretation:** Measures overall ranking quality of top 10 results

## Usage

### Comparing Against Baseline

```bash
# Run current benchmarks
npm run bench

# Compare against baseline
node test/baselines/compare.js test/baselines/node_baseline_2026-01-28.json
```

### Generating New Baselines

```bash
# Run benchmarks and capture output
cd test/fixtures
docker-compose run --rm pampax-fixture-gen npm run bench > bench_output.txt

# Extract metrics and create baseline JSON
# (Manual process - see node_baseline_2026-01-28.json as template)
```

### Go Port Validation

```bash
# 1. Run Go implementation on same fixtures
go-pampax bench --fixture=test/fixtures/small/

# 2. Compare metrics against Node.js baseline
# Required: P@1, MRR@5, nDCG@10 match within ±0.001 tolerance

# 3. Verify ranking order matches
diff <(go-pampax search "query" | jq '.results[].sha') \
     <(node pampax search "query" | jq '.results[].sha')
```

## Baseline Acceptance Criteria

For a Go implementation to be considered "compatible" with Node.js:

1. **Exact Metrics:** P@1, MRR@5, nDCG@10 match within ±0.001 (floating-point tolerance)
2. **Ranking Order:** Top-5 results per query appear in same order (SHA comparison)
3. **Edge Cases:** Handle empty queries, missing embeddings, and invalid inputs identically
4. **Performance:** No requirement for speed parity (Go expected to be faster)

## Notes

- **Synthetic Data:** Current baselines use mock fixtures with 4D embeddings
- **Real-World Validation:** Future baselines should use actual codebases with 768D embeddings
- **Version Tracking:** Create new baseline files for major version changes (e.g., v2.0.0)
- **Platform Variance:** Cross-platform baselines may differ due to floating-point precision

## Related Documentation

- [Stage 0 Compatibility Contract](../../instructions/GO_PORT_STAGE0_DETAILS.md)
- [Benchmark Test Suite](../benchmarks/bench.test.js)
- [IR Metrics Implementation](../../src/metrics/ir.js)
- [Fixture Documentation](../fixtures/README.md)
