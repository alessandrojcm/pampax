PRAGMA page_size = 4096;
PRAGMA journal_mode = delete;
PRAGMA encoding = 'UTF-8';
PRAGMA foreign_keys = OFF;

CREATE TABLE code_chunks (
    id TEXT PRIMARY KEY,
    file_path TEXT NOT NULL,
    symbol TEXT NOT NULL,
    sha TEXT NOT NULL,
    lang TEXT NOT NULL,
    chunk_type TEXT DEFAULT 'function',
    embedding BLOB,
    embedding_provider TEXT,
    embedding_dimensions INTEGER,
    pampa_tags TEXT,
    pampa_intent TEXT,
    pampa_description TEXT,
    doc_comments TEXT,
    variables_used TEXT,
    context_info TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_file_path ON code_chunks(file_path);
CREATE INDEX idx_symbol ON code_chunks(symbol);
CREATE INDEX idx_lang ON code_chunks(lang);
CREATE INDEX idx_provider ON code_chunks(embedding_provider);
CREATE INDEX idx_chunk_type ON code_chunks(chunk_type);
CREATE INDEX idx_pampa_tags ON code_chunks(pampa_tags);
CREATE INDEX idx_pampa_intent ON code_chunks(pampa_intent);
CREATE INDEX idx_lang_provider ON code_chunks(lang, embedding_provider, embedding_dimensions);

CREATE TABLE intention_cache (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query_normalized TEXT NOT NULL,
    original_query TEXT NOT NULL,
    target_sha TEXT NOT NULL,
    confidence REAL DEFAULT 1.0,
    usage_count INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_query_normalized ON intention_cache(query_normalized);
CREATE INDEX idx_target_sha ON intention_cache(target_sha);
CREATE INDEX idx_usage_count ON intention_cache(usage_count DESC);

CREATE TABLE query_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pattern TEXT NOT NULL UNIQUE,
    frequency INTEGER DEFAULT 1,
    typical_results TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pattern_frequency ON query_patterns(frequency DESC);
