package settings
import (
	"github.com/clidey/whodb/core/src/common/config"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
)
type gcpSection struct {
	Providers []persistedGCPProviderConfig `json:"providers"`
}
type persistedGCPProviderConfig struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	ProjectID             string `json:"projectId"`
	Region                string `json:"region"`
	AuthMethod            string `json:"authMethod"`
	ServiceAccountKeyPath string `json:"serviceAccountKeyPath,omitempty"`
	DiscoverCloudSQL      bool   `json:"discoverCloudSQL"`
	DiscoverAlloyDB       bool   `json:"discoverAlloyDB"`
	DiscoverMemorystore   bool   `json:"discoverMemorystore"`
}
func saveGCPProvidersToFile() error {
	gcpProvidersMu.RLock()
	providersList := make([]persistedGCPProviderConfig, 0, len(gcpProviders))
	for _, state := range gcpProviders {
		providersList = append(providersList, persistedGCPProviderConfig{
			ID:                    state.Config.ID,
			Name:                  state.Config.Name,
			ProjectID:             state.Config.ProjectID,
			Region:                state.Config.Region,
			AuthMethod:            state.Config.AuthMethod,
			ServiceAccountKeyPath: state.Config.ServiceAccountKeyPath,
			DiscoverCloudSQL:      state.Config.DiscoverCloudSQL,
			DiscoverAlloyDB:       state.Config.DiscoverAlloyDB,
			DiscoverMemorystore:   state.Config.DiscoverMemorystore,
		})
	}
	gcpProvidersMu.RUnlock()
	section := gcpSection{Providers: providersList}
	opts := getConfigOptions()
	if err := config.WriteSection(config.SectionGCP, section, opts); err != nil {
		log.Warnf("Failed to save GCP providers: %v", err)
		return err
	}
	log.Debugf("Saved %d GCP provider(s) to config", len(providersList))
	return nil
}
func LoadGCPProvidersFromFile() error {
	if !env.IsGCPProviderEnabled {
		return nil
	}
	opts := getConfigOptions()
	var section gcpSection
	if err := config.ReadSection(config.SectionGCP, &section, opts); err != nil {
		log.Warnf("Failed to read GCP providers from config: %v", err)
		return err
	}
	if len(section.Providers) == 0 {
		return nil
	}
	gcpSkipPersist = true
	defer func() { gcpSkipPersist = false }()
	loadedCount := 0
	for _, cfg := range section.Providers {
		providerCfg := &GCPProviderConfig{
			ID:                    cfg.ID,
			Name:                  cfg.Name,
			ProjectID:             cfg.ProjectID,
			Region:                cfg.Region,
			AuthMethod:            cfg.AuthMethod,
			ServiceAccountKeyPath: cfg.ServiceAccountKeyPath,
			DiscoverCloudSQL:      cfg.DiscoverCloudSQL,
			DiscoverAlloyDB:       cfg.DiscoverAlloyDB,
			DiscoverMemorystore:   cfg.DiscoverMemorystore,
		}
		gcpProvidersMu.RLock()
		_, exists := gcpProviders[cfg.ID]
		gcpProvidersMu.RUnlock()
		if exists {
			log.Debugf("Skipping persisted GCP provider %s - already registered", cfg.ID)
			continue
		}
		_, err := AddGCPProvider(providerCfg)
		if err != nil {
			log.Warnf("Failed to load persisted GCP provider %s: %v", cfg.Name, err)
			continue
		}
		loadedCount++
	}
	if loadedCount > 0 {
		log.Infof("Loaded %d GCP provider(s) from config", loadedCount)
	}
	return nil
}