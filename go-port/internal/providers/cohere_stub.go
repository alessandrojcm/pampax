package providers

import "strings"

// CohereProvider is a Stage 4A stub implementation for Cohere embeddings.
type CohereProvider struct {
	apiKey     string
	model      string
	dimensions int
}

// NewCohereProvider creates a Cohere provider stub.
func NewCohereProvider(cfg FactoryConfig) *CohereProvider {
	model := strings.TrimSpace(cfg.CohereModel)
	if model == "" {
		model = defaultCohereModel
	}

	return &CohereProvider{
		apiKey:     strings.TrimSpace(cfg.CohereAPIKey),
		model:      model,
		dimensions: effectiveDimensions(cfg.Dimensions, 1024),
	}
}

func (p *CohereProvider) GenerateEmbedding(text string) ([]float64, error) {
	return fakeEmbedding("cohere:"+p.model, text, p.dimensions)
}

func (p *CohereProvider) GetDimensions() int {
	return p.dimensions
}

func (p *CohereProvider) GetName() string {
	return "Cohere"
}
