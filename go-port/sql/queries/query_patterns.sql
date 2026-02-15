-- name: InsertQueryPattern :exec
INSERT INTO query_patterns (
    pattern,
    frequency,
    typical_results
) VALUES (?, ?, ?)
ON CONFLICT(pattern) DO UPDATE SET
    frequency = excluded.frequency,
    typical_results = excluded.typical_results,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetQueryPatternByPattern :one
SELECT
    id,
    pattern,
    frequency,
    typical_results,
    created_at,
    updated_at
FROM query_patterns
WHERE pattern = ?
LIMIT 1;
