package providers
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
)
const (
	Ollama_LLMType LLMType = "Ollama"
)
type OllamaProvider struct{}
func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{}
}
func (p *OllamaProvider) GetType() LLMType {
	return Ollama_LLMType
}
func (p *OllamaProvider) GetProtocol() string { return "openai" }
func (p *OllamaProvider) GetDefaultEndpoint() string {
	return env.GetOllamaEndpoint()
}
func (p *OllamaProvider) ValidateConfig(config *ProviderConfig) error {
	if config.Endpoint == "" {
		config.Endpoint = p.GetDefaultEndpoint()
	}
	return nil
}
func (p *OllamaProvider) GetSupportedModels(config *ProviderConfig) ([]string, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/tags", config.Endpoint)
	resp, err := sendHTTPRequest("GET", url, nil, nil)
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch models from Ollama at %s", url)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Ollama models endpoint returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("failed to fetch models: %s", string(body))
	}
	var modelsResp struct {
		Models []struct {
			Name string `json:"model"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		log.WithError(err).Error("Failed to decode Ollama models response")
		return nil, err
	}
	models := []string{}
	for _, model := range modelsResp.Models {
		models = append(models, model.Name)
	}
	return models, nil
}
func (p *OllamaProvider) CreateBAMLClient(config *ProviderConfig, model string) (string, map[string]any, error) {
	if err := p.ValidateConfig(config); err != nil {
		return "", nil, err
	}
	baseURL := strings.TrimSuffix(config.Endpoint, "/api") + "/v1"
	return "openai-generic", map[string]any{
		"base_url":           baseURL,
		"model":              model,
		"default_role":       "user",
		"request_timeout_ms": 60000,
	}, nil
}