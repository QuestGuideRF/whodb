package providers
import (
	"errors"
	"fmt"
)
type LLMType string
type LLMModel string
type AIProvider interface {
	GetType() LLMType
	GetProtocol() string
	GetDefaultEndpoint() string
	ValidateConfig(config *ProviderConfig) error
	GetSupportedModels(config *ProviderConfig) ([]string, error)
	CreateBAMLClient(config *ProviderConfig, model string) (clientType string, opts map[string]any, err error)
}
type ProviderConfig struct {
	Type       LLMType
	APIKey     string
	Endpoint   string
	ProviderId string
	Metadata   map[string]string
}
type ProviderRegistry struct {
	providers map[LLMType]AIProvider
}
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[LLMType]AIProvider),
	}
}
func (r *ProviderRegistry) Register(provider AIProvider) {
	r.providers[provider.GetType()] = provider
}
func (r *ProviderRegistry) Get(providerType LLMType) (AIProvider, error) {
	if provider, ok := r.providers[providerType]; ok {
		return provider, nil
	}
	return nil, fmt.Errorf("unsupported provider type: %s", providerType)
}
func (r *ProviderRegistry) Has(providerType LLMType) bool {
	_, ok := r.providers[providerType]
	return ok
}
func (r *ProviderRegistry) List() []LLMType {
	types := make([]LLMType, 0, len(r.providers))
	for providerType := range r.providers {
		types = append(types, providerType)
	}
	return types
}
var providerRegistry *ProviderRegistry
func init() {
	providerRegistry = NewProviderRegistry()
}
func GetProvider(providerType LLMType) (AIProvider, error) {
	if providerRegistry == nil {
		return nil, errors.New("provider registry not initialized")
	}
	return providerRegistry.Get(providerType)
}
func RegisterProvider(provider AIProvider) {
	if providerRegistry == nil {
		providerRegistry = NewProviderRegistry()
	}
	providerRegistry.Register(provider)
}
func HasProvider(providerType LLMType) bool {
	if providerRegistry == nil {
		return false
	}
	return providerRegistry.Has(providerType)
}
func ListProviders() []LLMType {
	if providerRegistry == nil {
		return []LLMType{}
	}
	return providerRegistry.List()
}
func GetBAMLConfig(providerType string, apiKey, endpoint, model string) (string, map[string]any, error) {
	provider, err := GetProvider(LLMType(providerType))
	if err != nil {
		return "", nil, err
	}
	config := &ProviderConfig{
		Type:     LLMType(providerType),
		APIKey:   apiKey,
		Endpoint: endpoint,
	}
	return provider.CreateBAMLClient(config, model)
}