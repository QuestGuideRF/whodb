package providers
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"github.com/clidey/whodb/core/src/log"
)
const (
	Anthropic_LLMType LLMType = "Anthropic"
)
type AnthropicProvider struct{}
func NewAnthropicProvider() *AnthropicProvider {
	return &AnthropicProvider{}
}
func (p *AnthropicProvider) GetType() LLMType {
	return Anthropic_LLMType
}
func (p *AnthropicProvider) GetProtocol() string { return "anthropic" }
func (p *AnthropicProvider) GetDefaultEndpoint() string {
	return "https://api.anthropic.com/v1"
}
func (p *AnthropicProvider) ValidateConfig(config *ProviderConfig) error {
	if config.APIKey == "" {
		return errors.New("API key is required for Anthropic")
	}
	if config.Endpoint == "" {
		config.Endpoint = p.GetDefaultEndpoint()
	}
	return nil
}
func (p *AnthropicProvider) GetSupportedModels(config *ProviderConfig) ([]string, error) {
	if err := p.ValidateConfig(config); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/models", config.Endpoint)
	headers := map[string]string{
		"x-api-key":         config.APIKey,
		"anthropic-version": "2023-06-01",
		"Content-Type":      "application/json",
	}
	resp, err := sendHTTPRequest("GET", url, nil, headers)
	if err != nil {
		log.WithError(err).Errorf("Failed to fetch models from Anthropic at %s", url)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("Anthropic models endpoint returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("failed to fetch models: %s", string(body))
	}
	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		log.WithError(err).Error("Failed to decode Anthropic models response")
		return nil, err
	}
	var models []string
	for _, model := range modelsResp.Data {
		if strings.HasPrefix(model.ID, "claude-") {
			models = append(models, model.ID)
		}
	}
	return models, nil
}
func (p *AnthropicProvider) CreateBAMLClient(config *ProviderConfig, model string) (string, map[string]any, error) {
	opts := map[string]any{"model": model}
	if config.APIKey != "" {
		opts["api_key"] = config.APIKey
	}
	if config.Endpoint != "" && config.Endpoint != p.GetDefaultEndpoint() {
		opts["base_url"] = config.Endpoint
	}
	return "anthropic", opts, nil
}