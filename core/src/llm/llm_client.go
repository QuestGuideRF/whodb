package llm
import (
	"github.com/clidey/whodb/core/src/bamlconfig"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/llm/providers"
	"github.com/clidey/whodb/core/src/log"
)
func init() {
	providers.RegisterProvider(providers.NewOpenAIProvider())
	providers.RegisterProvider(providers.NewAnthropicProvider())
	providers.RegisterProvider(providers.NewOllamaProvider())
	providers.RegisterProvider(providers.NewLMStudioProvider())
	bamlconfig.RegisterBAMLConfigResolver(providers.GetBAMLConfig)
}
func RegisterGenericProviders(name string, providerId string, models []string, clientType string, baseURL string, apiKey string) {
	provider := providers.NewGenericProvider(providerId, name, models, clientType)
	providers.RegisterProvider(provider)
	env.AddGenericProvider(env.GenericProviderConfig{
		ProviderId: providerId,
		Name:       name,
		ClientType: clientType,
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Models:     models,
	})
}
type LLMType = providers.LLMType
type LLMModel = providers.LLMModel
const (
	Ollama_LLMType    = providers.Ollama_LLMType
	OpenAI_LLMType    = providers.OpenAI_LLMType
	Anthropic_LLMType = providers.Anthropic_LLMType
	LMStudio_LLMType  = providers.LMStudio_LLMType
)
type LLMClient struct {
	Type      LLMType
	APIKey    string
	ProfileId string
}
func getEndpointForProvider(providerType LLMType) string {
	switch providerType {
	case OpenAI_LLMType:
		return env.GetOpenAIEndpoint()
	case Anthropic_LLMType:
		return env.GetAnthropicEndpoint()
	case Ollama_LLMType:
		return env.GetOllamaEndpoint()
	case LMStudio_LLMType:
		return env.GetLMStudioEndpoint()
	default:
		for _, provider := range env.GenericProviders {
			if providers.LLMType(provider.ProviderId) == providerType {
				return provider.BaseURL
			}
		}
		log.Warnf("No endpoint found for provider type: %s", providerType)
		return ""
	}
}
func (c *LLMClient) GetSupportedModels() ([]string, error) {
	provider, err := providers.GetProvider(c.Type)
	if err != nil {
		log.WithError(err).Errorf("Provider not found for type: %s", c.Type)
		return nil, err
	}
	config := &providers.ProviderConfig{
		Type:     c.Type,
		APIKey:   c.APIKey,
		Endpoint: getEndpointForProvider(c.Type),
	}
	return provider.GetSupportedModels(config)
}