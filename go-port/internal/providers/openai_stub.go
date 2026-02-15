package providers

import "strings"

// OpenAIProvider is a Stage 4A stub implementation for OpenAI embeddings.
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	model      string
	dimensions int
}

// NewOpenAIProvider creates an OpenAI provider stub.
func NewOpenAIProvider(cfg FactoryConfig) *OpenAIProvider {
	model := strings.TrimSpace(cfg.OpenAIEmbeddingModel)
	if model == "" {
		model = defaultOpenAIModel
	}

	baseURL := strings.TrimSpace(cfg.OpenAIBaseURL)
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIProvider{
		apiKey:     strings.TrimSpace(cfg.OpenAIAPIKey),
		baseURL:    baseURL,
		model:      model,
		dimensions: effectiveDimensions(cfg.Dimensions, 3072),
	}
}

func (p *OpenAIProvider) GenerateEmbedding(text string) ([]float64, error) {
	return fakeEmbedding("openai:"+p.model+":"+p.baseURL, text, p.dimensions)
}

func (p *OpenAIProvider) GetDimensions() int {
	return p.dimensions
}

func (p *OpenAIProvider) GetName() string {
	return "OpenAI"
}
