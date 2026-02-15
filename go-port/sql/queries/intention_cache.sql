-- name: InsertIntention :exec
INSERT INTO intention_cache (
    query_normalized,
    original_query,
    target_sha,
    confidence,
    usage_count
) VALUES (?, ?, ?, ?, ?);

-- name: GetIntentionsByQuery :many
SELECT
    id,
    query_normalized,
    original_query,
    target_sha,
    confidence,
    usage_count,
    created_at,
    last_used
FROM intention_cache
WHERE query_normalized = ?
ORDER BY usage_count DESC, last_used DESC;
