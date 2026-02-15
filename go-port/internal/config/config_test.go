package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("PAMPAX_OPENAI_BASE_URL", "")
	t.Setenv("PAMPAX_OPENAI_EMBEDDING_MODEL", "")
	t.Setenv("PAMPAX_TRANSFORMERS_MODEL", "")
	t.Setenv("PAMPAX_OLLAMA_BASE_URL", "")
	t.Setenv("PAMPAX_OLLAMA_MODEL", "")
	t.Setenv("PAMPAX_COHERE_MODEL", "")
	t.Setenv("PAMPAX_MAX_TOKENS", "")
	t.Setenv("PAMPAX_DIMENSIONS", "")
	t.Setenv("PAMPAX_RATE_LIMIT", "")
	t.Setenv("PAMPAX_RERANKER_MODE", "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.OpenAIBaseURL != "https://api.openai.com/v1" {
		t.Fatalf("unexpected default OpenAIBaseURL: %q", cfg.OpenAIBaseURL)
	}
	if cfg.OpenAIEmbeddingModel != "text-embedding-3-large" {
		t.Fatalf("unexpected default OpenAIEmbeddingModel: %q", cfg.OpenAIEmbeddingModel)
	}
	if cfg.TransformersModel != "Xenova/all-MiniLM-L6-v2" {
		t.Fatalf("unexpected default TransformersModel: %q", cfg.TransformersModel)
	}
	if cfg.OllamaBaseURL != "http://localhost:11434" {
		t.Fatalf("unexpected default OllamaBaseURL: %q", cfg.OllamaBaseURL)
	}
	if cfg.OllamaModel != "nomic-embed-text" {
		t.Fatalf("unexpected default OllamaModel: %q", cfg.OllamaModel)
	}
	if cfg.CohereModel != "embed-english-v3.0" {
		t.Fatalf("unexpected default CohereModel: %q", cfg.CohereModel)
	}
	if cfg.MaxTokens != 8191 {
		t.Fatalf("unexpected default MaxTokens: %d", cfg.MaxTokens)
	}
	if cfg.Dimensions != 1536 {
		t.Fatalf("unexpected default Dimensions: %d", cfg.Dimensions)
	}
	if cfg.RateLimit != 60 {
		t.Fatalf("unexpected default RateLimit: %d", cfg.RateLimit)
	}
	if cfg.RerankerMode != "off" {
		t.Fatalf("unexpected default RerankerMode: %q", cfg.RerankerMode)
	}
}

func TestLoadReadsEnvAndConfigFile(t *testing.T) {
	t.Setenv("PAMPAX_OPENAI_API_KEY", "env-key")
	t.Setenv("PAMPAX_OPENAI_EMBEDDING_MODEL", "")
	t.Setenv("PAMPAX_OLLAMA_BASE_URL", "")
	t.Setenv("PAMPAX_OLLAMA_MODEL", "")
	t.Setenv("PAMPAX_COHERE_API_KEY", "cohere-key")
	t.Setenv("PAMPAX_COHERE_MODEL", "")
	t.Setenv("PAMPAX_MAX_TOKENS", "")
	t.Setenv("PAMPAX_DIMENSIONS", "")
	t.Setenv("PAMPAX_RATE_LIMIT", "")
	t.Setenv("PAMPAX_RERANKER_MODE", "")
	t.Setenv("PAMPAX_RERANKER_API_KEY", "rk-key")
	t.Setenv("PAMPAX_RERANKER_MODEL", "cross-encoder")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pampax.yaml")
	configContent := []byte("openai:\n  embedding_model: text-embedding-3-small\ntransformers:\n  model: Xenova/all-mpnet-base-v2\nollama:\n  base_url: http://127.0.0.1:11434\n  model: nomic-embed-text\ncohere:\n  model: embed-multilingual-v3.0\nmax_tokens: 4096\ndimensions: 3072\nrate_limit: 20\nreranker:\n  mode: transformers\n")
	if err := os.WriteFile(configPath, configContent, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.OpenAIAPIKey != "env-key" {
		t.Fatalf("expected env api key, got %q", cfg.OpenAIAPIKey)
	}
	if cfg.OpenAIEmbeddingModel != "text-embedding-3-small" {
		t.Fatalf("expected file model, got %q", cfg.OpenAIEmbeddingModel)
	}
	if cfg.TransformersModel != "Xenova/all-mpnet-base-v2" {
		t.Fatalf("expected transformers model, got %q", cfg.TransformersModel)
	}
	if cfg.OllamaBaseURL != "http://127.0.0.1:11434" {
		t.Fatalf("expected ollama base URL, got %q", cfg.OllamaBaseURL)
	}
	if cfg.OllamaModel != "nomic-embed-text" {
		t.Fatalf("expected ollama model, got %q", cfg.OllamaModel)
	}
	if cfg.CohereAPIKey != "cohere-key" {
		t.Fatalf("expected cohere api key, got %q", cfg.CohereAPIKey)
	}
	if cfg.CohereModel != "embed-multilingual-v3.0" {
		t.Fatalf("expected cohere model, got %q", cfg.CohereModel)
	}
	if cfg.MaxTokens != 4096 {
		t.Fatalf("expected file max tokens, got %d", cfg.MaxTokens)
	}
	if cfg.Dimensions != 3072 {
		t.Fatalf("expected file dimensions, got %d", cfg.Dimensions)
	}
	if cfg.RateLimit != 20 {
		t.Fatalf("expected file rate limit, got %d", cfg.RateLimit)
	}
	if cfg.RerankerMode != "transformers" {
		t.Fatalf("expected file reranker mode, got %q", cfg.RerankerMode)
	}
	if cfg.Reranker["api_key"] != "rk-key" {
		t.Fatalf("expected reranker api_key from env, got %q", cfg.Reranker["api_key"])
	}
	if cfg.Reranker["model"] != "cross-encoder" {
		t.Fatalf("expected reranker model from env, got %q", cfg.Reranker["model"])
	}
}

func TestLoadRejectsInvalidRerankerMode(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	configContent := []byte("reranker:\n  mode: invalid\n")
	if err := os.WriteFile(configPath, configContent, 0o644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "validate config") {
		t.Fatalf("expected validation error message, got %v", err)
	}
}
