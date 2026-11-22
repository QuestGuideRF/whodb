package envconfig
import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/migrate"
	"github.com/clidey/whodb/core/src/types"
)
func GetDefaultDatabaseCredentials(databaseType string) []types.DatabaseCredentials {
	uppercaseDatabaseType := strings.ToUpper(databaseType)
	credEnvVar := fmt.Sprintf("WHODB_%s", uppercaseDatabaseType)
	credEnvValue := os.Getenv(credEnvVar)
	if credEnvValue == "" {
		return findAllDatabaseCredentials(databaseType)
	}
	var creds []types.DatabaseCredentials
	err := json.Unmarshal([]byte(credEnvValue), &creds)
	if err != nil {
		log.Error("🔴 [Database Error] Failed to parse database credentials from environment variable! Error: ", err)
		return nil
	}
	return creds
}
func findAllDatabaseCredentials(databaseType string) []types.DatabaseCredentials {
	uppercaseDatabaseType := strings.ToUpper(databaseType)
	i := 1
	var profiles []types.DatabaseCredentials
	for {
		databaseProfile := os.Getenv(fmt.Sprintf("WHODB_%s_%d", uppercaseDatabaseType, i))
		if databaseProfile == "" {
			break
		}
		var creds types.DatabaseCredentials
		err := json.Unmarshal([]byte(databaseProfile), &creds)
		if err != nil {
			log.Error("Unable to parse database credential: ", err)
			break
		}
		profiles = append(profiles, creds)
		i++
	}
	return profiles
}
func GetAWSProvidersFromEnv() ([]env.AWSProviderEnvConfig, error) {
	val := os.Getenv("WHODB_AWS_PROVIDER")
	if val == "" {
		return nil, nil
	}
	var configs []env.AWSProviderEnvConfig
	if err := json.Unmarshal([]byte(val), &configs); err != nil {
		log.Error("[AWS Provider] Failed to parse WHODB_AWS_PROVIDER: ", err)
		return nil, err
	}
	for i := range configs {
		if configs[i].DiscoverRDS == nil {
			t := true
			configs[i].DiscoverRDS = &t
		}
		if configs[i].DiscoverElastiCache == nil {
			t := true
			configs[i].DiscoverElastiCache = &t
		}
		if configs[i].DiscoverDocumentDB == nil {
			t := true
			configs[i].DiscoverDocumentDB = &t
		}
	}
	return configs, nil
}
func GetAzureProvidersFromEnv() ([]env.AzureProviderEnvConfig, error) {
	val := os.Getenv("WHODB_AZURE_PROVIDER")
	if val == "" {
		return nil, nil
	}
	var configs []env.AzureProviderEnvConfig
	if err := json.Unmarshal([]byte(val), &configs); err != nil {
		log.Error("[Azure Provider] Failed to parse WHODB_AZURE_PROVIDER: ", err)
		return nil, err
	}
	for i := range configs {
		if configs[i].DiscoverPostgreSQL == nil {
			t := true
			configs[i].DiscoverPostgreSQL = &t
		}
		if configs[i].DiscoverMySQL == nil {
			t := true
			configs[i].DiscoverMySQL = &t
		}
		if configs[i].DiscoverRedis == nil {
			t := true
			configs[i].DiscoverRedis = &t
		}
		if configs[i].DiscoverCosmosDB == nil {
			t := true
			configs[i].DiscoverCosmosDB = &t
		}
	}
	return configs, nil
}
func GetGCPProvidersFromEnv() ([]env.GCPProviderEnvConfig, error) {
	val := os.Getenv("WHODB_GCP_PROVIDER")
	if val == "" {
		return nil, nil
	}
	var configs []env.GCPProviderEnvConfig
	if err := json.Unmarshal([]byte(val), &configs); err != nil {
		log.Error("[GCP Provider] Failed to parse WHODB_GCP_PROVIDER: ", err)
		return nil, err
	}
	for i := range configs {
		if configs[i].DiscoverCloudSQL == nil {
			t := true
			configs[i].DiscoverCloudSQL = &t
		}
		if configs[i].DiscoverAlloyDB == nil {
			t := true
			configs[i].DiscoverAlloyDB = &t
		}
		if configs[i].DiscoverMemorystore == nil {
			t := true
			configs[i].DiscoverMemorystore = &t
		}
	}
	return configs, nil
}
func ParseGenericProviders() []env.GenericProviderConfig {
	var providers []env.GenericProviderConfig
	processed := make(map[string]bool)
	for _, envVar := range os.Environ() {
		if !strings.HasPrefix(envVar, "WHODB_AI_GENERIC_") {
			continue
		}
		parts := strings.SplitN(envVar, "=", 2)
		key := parts[0]
		keyParts := strings.Split(key, "_")
		if len(keyParts) < 5 {
			continue
		}
		idParts := keyParts[3 : len(keyParts)-1]
		providerID := strings.Join(idParts, "_")
		if processed[providerID] {
			continue
		}
		prefix := fmt.Sprintf("WHODB_AI_GENERIC_%s_", providerID)
		name := os.Getenv(prefix + "NAME")
		clientType := os.Getenv(prefix + "TYPE")
		baseURL := os.Getenv(prefix + "BASE_URL")
		apiKey := os.Getenv(prefix + "API_KEY")
		modelsStr := os.Getenv(prefix + "MODELS")
		if baseURL == "" || modelsStr == "" {
			log.Warnf("Incomplete generic provider config for %s, skipping (missing base_url or models)", providerID)
			continue
		}
		models := common.FilterList(strings.Split(modelsStr, ","), func(item string) bool {
			return strings.TrimSpace(item) != ""
		})
		if len(models) == 0 {
			log.Warnf("No models specified for generic provider %s, skipping", providerID)
			continue
		}
		if name == "" {
			name = providerID
		}
		if clientType == "" {
			clientType = "openai-generic"
		}
		providers = append(providers, env.GenericProviderConfig{
			ProviderId: strings.ToLower(providerID),
			Name:       name,
			ClientType: clientType,
			BaseURL:    baseURL,
			APIKey:     apiKey,
			Models:     models,
		})
		processed[providerID] = true
	}
	if len(providers) > 0 {
		log.Infof("Discovered %d generic AI provider(s)", len(providers))
		for _, provider := range providers {
			log.Infof("  - %s (%s) with %d model(s)", provider.Name, provider.ProviderId, len(provider.Models))
		}
	}
	return providers
}
func GetConfiguredChatProviders() []env.ChatProvider {
	var providers []env.ChatProvider
	if len(env.OpenAIAPIKey) > 0 {
		name := env.OpenAIName
		if name == "" {
			name = "OpenAI"
		}
		providers = append(providers, env.ChatProvider{
			Type:       "OpenAI",
			Name:       name,
			APIKey:     env.OpenAIAPIKey,
			Endpoint:   env.GetOpenAIEndpoint(),
			ProviderId: "openai-1",
		})
	}
	if len(env.AnthropicAPIKey) > 0 {
		name := env.AnthropicName
		if name == "" {
			name = "Anthropic"
		}
		providers = append(providers, env.ChatProvider{
			Type:       "Anthropic",
			Name:       name,
			APIKey:     env.AnthropicAPIKey,
			Endpoint:   env.GetAnthropicEndpoint(),
			ProviderId: "anthropic-1",
		})
	}
	if os.Getenv("WHODB_OPENAI_COMPATIBLE_ENDPOINT") != "" || os.Getenv("WHODB_OPENAI_COMPATIBLE_API_KEY") != "" || os.Getenv("WHODB_CUSTOM_MODELS") != "" {
		migrate.DeprecatedOpenAICompatibleEnv = true
	}
	for _, genericProvider := range env.GenericProviders {
		providers = append(providers, env.ChatProvider{
			Type:       genericProvider.ProviderId,
			Name:       genericProvider.Name,
			APIKey:     genericProvider.APIKey,
			Endpoint:   common.ResolveLocalURL(genericProvider.BaseURL),
			ProviderId: genericProvider.ProviderId,
			ClientType: genericProvider.ClientType,
			IsGeneric:  true,
		})
	}
	name := env.LMStudioName
	if name == "" {
		name = "LM Studio"
	}
	providers = append(providers, env.ChatProvider{
		Type:       "LMStudio",
		Name:       name,
		APIKey:     env.LMStudioAPIKey,
		Endpoint:   env.GetLMStudioEndpoint(),
		ProviderId: "lmstudio-1",
	})
	name = env.OllamaName
	if name == "" {
		name = "Ollama"
	}
	providers = append(providers, env.ChatProvider{
		Type:       "Ollama",
		Name:       name,
		APIKey:     "",
		Endpoint:   env.GetOllamaEndpoint(),
		ProviderId: "ollama-1",
	})
	return providers
}
type SupplementalCredentialResolver func(providerId string) (token string, endpoint string, found bool)
var supplementalResolver SupplementalCredentialResolver
func RegisterSupplementalCredentialResolver(fn SupplementalCredentialResolver) {
	supplementalResolver = fn
}
func ResolveProviderCredentials(providerId, requestToken, requestEndpoint, requestModelType string) env.ResolvedCredentials {
	result := env.ResolvedCredentials{
		ModelType: requestModelType,
		Token:     requestToken,
		Endpoint:  requestEndpoint,
	}
	if providerId == "" {
		return result
	}
	for _, provider := range GetConfiguredChatProviders() {
		if provider.ProviderId != providerId {
			continue
		}
		if result.Token == "" {
			result.Token = provider.APIKey
		}
		if result.Endpoint == "" {
			result.Endpoint = provider.Endpoint
		}
		return result
	}
	if result.Token == "" && supplementalResolver != nil {
		token, endpoint, found := supplementalResolver(providerId)
		if found {
			result.Token = token
			if result.Endpoint == "" {
				result.Endpoint = endpoint
			}
		}
	}
	return result
}