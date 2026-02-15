package providers

import "strings"

// TransformersProvider is a Stage 4A stub for local embeddings.
type TransformersProvider struct {
	model      string
	dimensions int
}

// NewTransformersProvider creates a local transformers provider stub.
func NewTransformersProvider(cfg FactoryConfig) *TransformersProvider {
	model := strings.TrimSpace(cfg.TransformersModel)
	if model == "" {
		model = defaultTransformersModel
	}

	return &TransformersProvider{
		model:      model,
		dimensions: effectiveDimensions(cfg.Dimensions, 384),
	}
}

func (p *TransformersProvider) GenerateEmbedding(text string) ([]float64, error) {
	return fakeEmbedding("transformers:"+p.model, text, p.dimensions)
}

func (p *TransformersProvider) GetDimensions() int {
	return p.dimensions
}

func (p *TransformersProvider) GetName() string {
	return "Transformers.js (Local)"
}
