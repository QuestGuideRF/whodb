package providers
import (
	"fmt"
)
type GenericProvider struct {
	providerId string
	name       string
	models     []string
	clientType string
}
func NewGenericProvider(providerId, name string, models []string, clientType string) *GenericProvider {
	if clientType == "" {
		clientType = "openai-generic"
	}
	return &GenericProvider{
		providerId: providerId,
		name:       name,
		models:     models,
		clientType: clientType,
	}
}
func (p *GenericProvider) GetType() LLMType {
	return LLMType(p.providerId)
}
func (p *GenericProvider) GetProtocol() string {
	switch p.clientType {
	case "anthropic":
		return "anthropic"
	case "aws-bedrock":
		return "aws-bedrock"
	default:
		return "openai"
	}
}
func (p *GenericProvider) GetDefaultEndpoint() string {
	return ""
}
func (p *GenericProvider) ValidateConfig(config *ProviderConfig) error {
	if config.Endpoint == "" {
		return fmt.Errorf("endpoint is required for generic provider %s", p.name)
	}
	return nil
}
func (p *GenericProvider) GetSupportedModels(config *ProviderConfig) ([]string, error) {
	return p.models, nil
}
func (p *GenericProvider) CreateBAMLClient(config *ProviderConfig, model string) (string, map[string]any, error) {
	if err := p.ValidateConfig(config); err != nil {
		return "", nil, err
	}
	opts := map[string]any{
		"base_url":           config.Endpoint,
		"model":              model,
		"default_role":       "user",
		"request_timeout_ms": 60000,
	}
	if config.APIKey != "" {
		opts["api_key"] = config.APIKey
	}
	return p.clientType, opts, nil
}