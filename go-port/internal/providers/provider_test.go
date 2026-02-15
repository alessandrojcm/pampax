package providers

import "testing"

func TestNewEmbeddingProviderExplicit(t *testing.T) {
	cfg := FactoryConfig{Dimensions: 256}

	provider, err := NewEmbeddingProvider("openai", cfg)
	if err != nil {
		t.Fatalf("NewEmbeddingProvider() error = %v", err)
	}

	if provider.GetName() != "OpenAI" {
		t.Fatalf("unexpected provider name: %q", provider.GetName())
	}
	if provider.GetDimensions() != 256 {
		t.Fatalf("unexpected dimensions: %d", provider.GetDimensions())
	}
}

func TestNewEmbeddingProviderAutoSelection(t *testing.T) {
	tests := []struct {
		name     string
		cfg      FactoryConfig
		expected string
	}{
		{
			name:     "prefers openai when api key exists",
			cfg:      FactoryConfig{OpenAIAPIKey: "sk-test"},
			expected: "OpenAI",
		},
		{
			name:     "falls back to cohere with cohere key",
			cfg:      FactoryConfig{CohereAPIKey: "cohere-key", OllamaBaseURL: ""},
			expected: "Cohere",
		},
		{
			name:     "uses ollama when configured",
			cfg:      FactoryConfig{OllamaBaseURL: "http://localhost:11434"},
			expected: "Ollama",
		},
		{
			name:     "falls back to local provider",
			cfg:      FactoryConfig{OllamaBaseURL: ""},
			expected: "Transformers.js (Local)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewEmbeddingProvider("auto", tt.cfg)
			if err != nil {
				t.Fatalf("NewEmbeddingProvider() error = %v", err)
			}
			if provider.GetName() != tt.expected {
				t.Fatalf("provider.GetName() = %q, want %q", provider.GetName(), tt.expected)
			}
		})
	}
}

func TestNewEmbeddingProviderRejectsUnknown(t *testing.T) {
	if _, err := NewEmbeddingProvider("mystery", FactoryConfig{}); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestGenerateEmbeddingDeterministic(t *testing.T) {
	provider := NewTransformersProvider(FactoryConfig{Dimensions: 8})

	a, err := provider.GenerateEmbedding("hello")
	if err != nil {
		t.Fatalf("GenerateEmbedding() error = %v", err)
	}
	b, err := provider.GenerateEmbedding("hello")
	if err != nil {
		t.Fatalf("GenerateEmbedding() error = %v", err)
	}

	if len(a) != 8 {
		t.Fatalf("expected 8 dimensions, got %d", len(a))
	}
	if len(b) != 8 {
		t.Fatalf("expected 8 dimensions, got %d", len(b))
	}

	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("embedding not deterministic at %d: %f != %f", i, a[i], b[i])
		}
	}
}
