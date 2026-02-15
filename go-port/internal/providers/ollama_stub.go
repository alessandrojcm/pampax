package providers

import "strings"

// OllamaProvider is a Stage 4A stub implementation for Ollama embeddings.
type OllamaProvider struct {
	baseURL    string
	model      string
	dimensions int
}

// NewOllamaProvider creates an Ollama provider stub.
func NewOllamaProvider(cfg FactoryConfig) *OllamaProvider {
	model := strings.TrimSpace(cfg.OllamaModel)
	if model == "" {
		model = defaultOllamaModel
	}

	baseURL := strings.TrimSpace(cfg.OllamaBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaProvider{
		baseURL:    baseURL,
		model:      model,
		dimensions: effectiveDimensions(cfg.Dimensions, 1024),
	}
}

func (p *OllamaProvider) GenerateEmbedding(text string) ([]float64, error) {
	return fakeEmbedding("ollama:"+p.model+":"+p.baseURL, text, p.dimensions)
}

func (p *OllamaProvider) GetDimensions() int {
	return p.dimensions
}

func (p *OllamaProvider) GetName() string {
	return "Ollama"
}
