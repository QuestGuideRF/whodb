package settings
import (
	"github.com/clidey/whodb/core/src/common/config"
	"github.com/clidey/whodb/core/src/common/datadir"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
)
type awsSection struct {
	Providers []persistedProviderConfig `json:"providers"`
}
type persistedProviderConfig struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Region              string `json:"region"`
	AuthMethod          string `json:"authMethod"`
	ProfileName         string `json:"profileName,omitempty"`
	DiscoverRDS         bool   `json:"discoverRDS"`
	DiscoverElastiCache bool   `json:"discoverElastiCache"`
	DiscoverDocumentDB  bool   `json:"discoverDocumentDB"`
}
func getConfigOptions() datadir.Options {
	return datadir.Options{
		AppName:           "whodb",
		EnterpriseEdition: env.IsEnterpriseEdition,
		Development:       env.IsDevelopment,
	}
}
func saveProvidersToFile() error {
	awsProvidersMu.RLock()
	providers := make([]persistedProviderConfig, 0, len(awsProviders))
	for _, state := range awsProviders {
		providers = append(providers, persistedProviderConfig{
			ID:                  state.Config.ID,
			Name:                state.Config.Name,
			Region:              state.Config.Region,
			AuthMethod:          state.Config.AuthMethod,
			ProfileName:         state.Config.ProfileName,
			DiscoverRDS:         state.Config.DiscoverRDS,
			DiscoverElastiCache: state.Config.DiscoverElastiCache,
			DiscoverDocumentDB:  state.Config.DiscoverDocumentDB,
		})
	}
	awsProvidersMu.RUnlock()
	section := awsSection{Providers: providers}
	opts := getConfigOptions()
	if err := config.WriteSection(config.SectionAWS, section, opts); err != nil {
		log.Warnf("Failed to save AWS providers: %v", err)
		return err
	}
	log.Debugf("Saved %d AWS provider(s) to config", len(providers))
	return nil
}
func LoadProvidersFromFile() error {
	if !env.IsAWSProviderEnabled {
		return nil
	}
	opts := getConfigOptions()
	var section awsSection
	if err := config.ReadSection(config.SectionAWS, &section, opts); err != nil {
		log.Warnf("Failed to read AWS providers from config: %v", err)
		return err
	}
	if len(section.Providers) == 0 {
		return nil
	}
	skipPersist = true
	defer func() { skipPersist = false }()
	loadedCount := 0
	for _, cfg := range section.Providers {
		providerCfg := &AWSProviderConfig{
			ID:                  cfg.ID,
			Name:                cfg.Name,
			Region:              cfg.Region,
			AuthMethod:          cfg.AuthMethod,
			ProfileName:         cfg.ProfileName,
			DiscoverRDS:         cfg.DiscoverRDS,
			DiscoverElastiCache: cfg.DiscoverElastiCache,
			DiscoverDocumentDB:  cfg.DiscoverDocumentDB,
		}
		awsProvidersMu.RLock()
		_, exists := awsProviders[cfg.ID]
		awsProvidersMu.RUnlock()
		if exists {
			log.Debugf("Skipping persisted provider %s - already registered", cfg.ID)
			continue
		}
		_, err := AddAWSProvider(providerCfg)
		if err != nil {
			log.Warnf("Failed to load persisted provider %s: %v", cfg.Name, err)
			continue
		}
		loadedCount++
	}
	if loadedCount > 0 {
		log.Infof("Loaded %d AWS provider(s) from config", loadedCount)
	}
	return nil
}