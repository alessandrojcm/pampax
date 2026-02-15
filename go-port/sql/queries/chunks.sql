-- name: InsertChunk :exec
INSERT INTO code_chunks (
    id,
    file_path,
    symbol,
    sha,
    lang,
    chunk_type,
    embedding,
    embedding_provider,
    embedding_dimensions,
    pampa_tags,
    pampa_intent,
    pampa_description,
    doc_comments,
    variables_used,
    context_info
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetChunkBySHA :one
SELECT
    id,
    file_path,
    symbol,
    sha,
    lang,
    chunk_type,
    embedding,
    embedding_provider,
    embedding_dimensions,
    pampa_tags,
    pampa_intent,
    pampa_description,
    doc_comments,
    variables_used,
    context_info,
    created_at,
    updated_at
FROM code_chunks
WHERE sha = ?
LIMIT 1;

-- name: GetChunksByFilePath :many
SELECT
    id,
    file_path,
    symbol,
    sha,
    lang,
    chunk_type,
    embedding,
    embedding_provider,
    embedding_dimensions,
    pampa_tags,
    pampa_intent,
    pampa_description,
    doc_comments,
    variables_used,
    context_info,
    created_at,
    updated_at
FROM code_chunks
WHERE file_path = ?
ORDER BY id;

-- name: GetChunksByProvider :many
SELECT
    id,
    file_path,
    symbol,
    sha,
    lang,
    chunk_type,
    embedding,
    embedding_provider,
    embedding_dimensions,
    pampa_tags,
    pampa_intent,
    pampa_description,
    doc_comments,
    variables_used,
    context_info,
    created_at,
    updated_at
FROM code_chunks
WHERE embedding_provider = ?
ORDER BY id;

-- name: DeleteChunk :exec
DELETE FROM code_chunks
WHERE id = ?;
