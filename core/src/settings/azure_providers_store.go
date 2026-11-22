package settings
import (
	"github.com/clidey/whodb/core/src/common/config"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
)
type azureSection struct {
	Providers []persistedAzureProviderConfig `json:"providers"`
}
type persistedAzureProviderConfig struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	SubscriptionID     string `json:"subscriptionId"`
	TenantID           string `json:"tenantId,omitempty"`
	ClientID           string `json:"clientId,omitempty"`
	AuthMethod         string `json:"authMethod"`
	ResourceGroup      string `json:"resourceGroup,omitempty"`
	DiscoverPostgreSQL bool   `json:"discoverPostgreSQL"`
	DiscoverMySQL      bool   `json:"discoverMySQL"`
	DiscoverRedis      bool   `json:"discoverRedis"`
	DiscoverCosmosDB   bool   `json:"discoverCosmosDB"`
}
func saveAzureProvidersToFile() error {
	azureProvidersMu.RLock()
	providersList := make([]persistedAzureProviderConfig, 0, len(azureProviders))
	for _, state := range azureProviders {
		providersList = append(providersList, persistedAzureProviderConfig{
			ID:                 state.Config.ID,
			Name:               state.Config.Name,
			SubscriptionID:     state.Config.SubscriptionID,
			TenantID:           state.Config.TenantID,
			ClientID:           state.Config.ClientID,
			AuthMethod:         state.Config.AuthMethod,
			ResourceGroup:      state.Config.ResourceGroup,
			DiscoverPostgreSQL: state.Config.DiscoverPostgreSQL,
			DiscoverMySQL:      state.Config.DiscoverMySQL,
			DiscoverRedis:      state.Config.DiscoverRedis,
			DiscoverCosmosDB:   state.Config.DiscoverCosmosDB,
		})
	}
	azureProvidersMu.RUnlock()
	section := azureSection{Providers: providersList}
	opts := getConfigOptions()
	if err := config.WriteSection(config.SectionAzure, section, opts); err != nil {
		log.Warnf("Failed to save Azure providers: %v", err)
		return err
	}
	log.Debugf("Saved %d Azure provider(s) to config", len(providersList))
	return nil
}
func LoadAzureProvidersFromFile() error {
	if !env.IsAzureProviderEnabled {
		return nil
	}
	opts := getConfigOptions()
	var section azureSection
	if err := config.ReadSection(config.SectionAzure, &section, opts); err != nil {
		log.Warnf("Failed to read Azure providers from config: %v", err)
		return err
	}
	if len(section.Providers) == 0 {
		return nil
	}
	skipAzurePersist = true
	defer func() { skipAzurePersist = false }()
	loadedCount := 0
	for _, cfg := range section.Providers {
		providerCfg := &AzureProviderConfig{
			ID:                 cfg.ID,
			Name:               cfg.Name,
			SubscriptionID:     cfg.SubscriptionID,
			TenantID:           cfg.TenantID,
			ClientID:           cfg.ClientID,
			AuthMethod:         cfg.AuthMethod,
			ResourceGroup:      cfg.ResourceGroup,
			DiscoverPostgreSQL: cfg.DiscoverPostgreSQL,
			DiscoverMySQL:      cfg.DiscoverMySQL,
			DiscoverRedis:      cfg.DiscoverRedis,
			DiscoverCosmosDB:   cfg.DiscoverCosmosDB,
		}
		azureProvidersMu.RLock()
		_, exists := azureProviders[cfg.ID]
		azureProvidersMu.RUnlock()
		if exists {
			log.Debugf("Skipping persisted Azure provider %s - already registered", cfg.ID)
			continue
		}
		_, err := AddAzureProvider(providerCfg, "")
		if err != nil {
			log.Warnf("Failed to load persisted Azure provider %s: %v", cfg.Name, err)
			continue
		}
		loadedCount++
	}
	if loadedCount > 0 {
		log.Infof("Loaded %d Azure provider(s) from config", loadedCount)
	}
	return nil
}