package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const rerankerEnvPrefix = "PAMPAX_RERANKER_"

var configValidator = validator.New(validator.WithRequiredStructEnabled())

type Config struct {
	EncryptionKey        string            `validate:"omitempty"`
	OpenAIAPIKey         string            `validate:"omitempty"`
	OpenAIBaseURL        string            `validate:"required,url"`
	OpenAIEmbeddingModel string            `validate:"required"`
	TransformersModel    string            `validate:"required"`
	OllamaBaseURL        string            `validate:"required,url"`
	OllamaModel          string            `validate:"required"`
	CohereAPIKey         string            `validate:"omitempty"`
	CohereModel          string            `validate:"required"`
	MaxTokens            int               `validate:"gte=1"`
	Dimensions           int               `validate:"gte=1"`
	RateLimit            int               `validate:"gte=1"`
	RerankerMode         string            `validate:"oneof=off transformers api"`
	Reranker             map[string]string `validate:"-"`
}

func Load(configFile string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvPrefix("PAMPAX")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("openai.base_url", "https://api.openai.com/v1")
	v.SetDefault("openai.embedding_model", "text-embedding-3-large")
	v.SetDefault("transformers.model", "Xenova/all-MiniLM-L6-v2")
	v.SetDefault("ollama.base_url", "http://localhost:11434")
	v.SetDefault("ollama.model", "nomic-embed-text")
	v.SetDefault("cohere.model", "embed-english-v3.0")
	v.SetDefault("max_tokens", 8191)
	v.SetDefault("dimensions", 1536)
	v.SetDefault("rate_limit", 60)
	v.SetDefault("reranker.mode", "off")

	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	loaded := &Config{
		EncryptionKey:        strings.TrimSpace(v.GetString("encryption_key")),
		OpenAIAPIKey:         strings.TrimSpace(v.GetString("openai.api_key")),
		OpenAIBaseURL:        strings.TrimSpace(v.GetString("openai.base_url")),
		OpenAIEmbeddingModel: strings.TrimSpace(v.GetString("openai.embedding_model")),
		TransformersModel:    strings.TrimSpace(v.GetString("transformers.model")),
		OllamaBaseURL:        strings.TrimSpace(v.GetString("ollama.base_url")),
		OllamaModel:          strings.TrimSpace(v.GetString("ollama.model")),
		CohereAPIKey:         strings.TrimSpace(v.GetString("cohere.api_key")),
		CohereModel:          strings.TrimSpace(v.GetString("cohere.model")),
		MaxTokens:            v.GetInt("max_tokens"),
		Dimensions:           v.GetInt("dimensions"),
		RateLimit:            v.GetInt("rate_limit"),
		RerankerMode:         strings.TrimSpace(v.GetString("reranker.mode")),
		Reranker:             collectRerankerEnv(),
	}

	if loaded.OpenAIBaseURL == "" {
		loaded.OpenAIBaseURL = "https://api.openai.com/v1"
	}
	if loaded.OpenAIEmbeddingModel == "" {
		loaded.OpenAIEmbeddingModel = "text-embedding-3-large"
	}
	if loaded.TransformersModel == "" {
		loaded.TransformersModel = "Xenova/all-MiniLM-L6-v2"
	}
	if loaded.OllamaBaseURL == "" {
		loaded.OllamaBaseURL = "http://localhost:11434"
	}
	if loaded.OllamaModel == "" {
		loaded.OllamaModel = "nomic-embed-text"
	}
	if loaded.CohereModel == "" {
		loaded.CohereModel = "embed-english-v3.0"
	}
	if loaded.MaxTokens <= 0 {
		loaded.MaxTokens = 8191
	}
	if loaded.Dimensions <= 0 {
		loaded.Dimensions = 1536
	}
	if loaded.RateLimit <= 0 {
		loaded.RateLimit = 60
	}
	if loaded.RerankerMode == "" {
		loaded.RerankerMode = "off"
	}

	if err := configValidator.Struct(loaded); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return loaded, nil
}

func collectRerankerEnv() map[string]string {
	values := map[string]string{}

	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, rerankerEnvPrefix) {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.TrimPrefix(parts[0], rerankerEnvPrefix))
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		values[strings.ToLower(key)] = value
	}

	return values
}
