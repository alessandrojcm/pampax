package providers

import (
	"errors"
	"fmt"
	"strings"
)

const (
	defaultOpenAIModel       = "text-embedding-3-large"
	defaultTransformersModel = "Xenova/all-MiniLM-L6-v2"
	defaultOllamaModel       = "nomic-embed-text"
	defaultCohereModel       = "embed-english-v3.0"
)

// EmbeddingProvider defines the contract for embedding providers.
type EmbeddingProvider interface {
	GenerateEmbedding(text string) ([]float64, error)
	GetDimensions() int
	GetName() string
}

// FactoryConfig includes provider-related configuration.
type FactoryConfig struct {
	OpenAIAPIKey         string
	OpenAIBaseURL        string
	OpenAIEmbeddingModel string
	TransformersModel    string
	OllamaBaseURL        string
	OllamaModel          string
	CohereAPIKey         string
	CohereModel          string
	Dimensions           int
}

// NewEmbeddingProvider builds a provider implementation based on the requested name.
func NewEmbeddingProvider(providerName string, cfg FactoryConfig) (EmbeddingProvider, error) {
	resolved := strings.ToLower(strings.TrimSpace(providerName))
	if resolved == "" || resolved == "auto" {
		resolved = resolveAutoProvider(cfg)
	}

	switch resolved {
	case "openai":
		return NewOpenAIProvider(cfg), nil
	case "transformers", "local":
		return NewTransformersProvider(cfg), nil
	case "ollama":
		return NewOllamaProvider(cfg), nil
	case "cohere":
		return NewCohereProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q: must be one of [auto, openai, transformers, local, ollama, cohere]", providerName)
	}
}

func resolveAutoProvider(cfg FactoryConfig) string {
	if strings.TrimSpace(cfg.OpenAIAPIKey) != "" {
		return "openai"
	}
	if strings.TrimSpace(cfg.CohereAPIKey) != "" {
		return "cohere"
	}
	if strings.TrimSpace(cfg.OllamaBaseURL) != "" {
		return "ollama"
	}
	return "transformers"
}

func effectiveDimensions(configDims int, defaultDims int) int {
	if configDims > 0 {
		return configDims
	}
	return defaultDims
}

func fakeEmbedding(seed string, text string, dimensions int) ([]float64, error) {
	if dimensions <= 0 {
		return nil, errors.New("embedding dimensions must be greater than 0")
	}

	combined := seed + "\n" + text
	values := make([]float64, dimensions)
	raw := []byte(combined)
	if len(raw) == 0 {
		raw = []byte("stub")
	}

	for i := 0; i < dimensions; i++ {
		b := raw[i%len(raw)]
		values[i] = (float64(int(b)%200) / 100.0) - 1.0
	}

	return values, nil
}
