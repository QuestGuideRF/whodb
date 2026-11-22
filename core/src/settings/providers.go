package settings
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	awsinfra "github.com/clidey/whodb/core/src/aws"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/envconfig"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/providers"
	awsprovider "github.com/clidey/whodb/core/src/providers/aws"
)
var (
	ErrProviderNotFound = errors.New("aws provider not found")
	ErrProviderAlreadyExists = errors.New("aws provider already exists")
)
type AWSProviderConfig struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Region              string `json:"region"`
	AuthMethod          string `json:"authMethod"`
	ProfileName         string `json:"profileName,omitempty"`
	DiscoverRDS         bool   `json:"discoverRDS"`
	DiscoverElastiCache bool   `json:"discoverElastiCache"`
	DiscoverDocumentDB  bool   `json:"discoverDocumentDB"`
}
type AWSProviderState struct {
	Config          *AWSProviderConfig
	Provider        *awsprovider.Provider
	Status          string
	LastDiscoveryAt *time.Time
	DiscoveredCount int
	Error           string
}
var (
	awsProvidersMu sync.RWMutex
	awsProviders   = make(map[string]*AWSProviderState)
	skipPersist    bool
)
func GetAWSProviders() []*AWSProviderState {
	awsProvidersMu.RLock()
	defer awsProvidersMu.RUnlock()
	result := make([]*AWSProviderState, 0, len(awsProviders))
	for _, state := range awsProviders {
		result = append(result, state)
	}
	return result
}
func GetAWSProvider(id string) (*AWSProviderState, error) {
	awsProvidersMu.RLock()
	defer awsProvidersMu.RUnlock()
	state, exists := awsProviders[id]
	if !exists {
		return nil, ErrProviderNotFound
	}
	return state, nil
}
func AddAWSProvider(cfg *AWSProviderConfig) (*AWSProviderState, error) {
	awsProvidersMu.Lock()
	if _, exists := awsProviders[cfg.ID]; exists {
		awsProvidersMu.Unlock()
		return nil, ErrProviderAlreadyExists
	}
	providerCfg := configToProviderConfig(cfg)
	provider, err := awsprovider.New(providerCfg)
	if err != nil {
		awsProvidersMu.Unlock()
		return nil, err
	}
	registry := providers.GetDefaultRegistry()
	if err := registry.Register(provider); err != nil {
		awsProvidersMu.Unlock()
		return nil, err
	}
	state := &AWSProviderState{
		Config:   cfg,
		Provider: provider,
		Status:   "Connected",
	}
	awsProviders[cfg.ID] = state
	awsProvidersMu.Unlock()
	if !skipPersist {
		if err := saveProvidersToFile(); err != nil {
			log.Warnf("Failed to persist provider after add: %v", err)
		}
	}
	return state, nil
}
func UpdateAWSProvider(id string, cfg *AWSProviderConfig) (*AWSProviderState, error) {
	cfg.ID = id
	providerCfg := configToProviderConfig(cfg)
	provider, err := awsprovider.New(providerCfg)
	if err != nil {
		return nil, err
	}
	awsProvidersMu.Lock()
	if _, exists := awsProviders[id]; !exists {
		awsProvidersMu.Unlock()
		return nil, ErrProviderNotFound
	}
	registry := providers.GetDefaultRegistry()
	if err := registry.Unregister(id); err != nil {
		log.Warnf("Failed to unregister provider %s during update: %v", id, err)
	}
	if err := registry.Register(provider); err != nil {
		awsProvidersMu.Unlock()
		return nil, err
	}
	state := &AWSProviderState{
		Config:   cfg,
		Provider: provider,
		Status:   "Connected",
	}
	awsProviders[id] = state
	awsProvidersMu.Unlock()
	if !skipPersist {
		if err := saveProvidersToFile(); err != nil {
			log.Warnf("Failed to persist provider after update: %v", err)
		}
	}
	return state, nil
}
func RemoveAWSProvider(id string) error {
	awsProvidersMu.Lock()
	if _, exists := awsProviders[id]; !exists {
		awsProvidersMu.Unlock()
		return ErrProviderNotFound
	}
	registry := providers.GetDefaultRegistry()
	if err := registry.Unregister(id); err != nil {
		log.Warnf("Failed to unregister provider %s during removal: %v", id, err)
	}
	delete(awsProviders, id)
	awsProvidersMu.Unlock()
	if !skipPersist {
		if err := saveProvidersToFile(); err != nil {
			log.Warnf("Failed to persist provider after remove: %v", err)
		}
	}
	return nil
}
func TestAWSProvider(id string) (string, error) {
	awsProvidersMu.RLock()
	state, exists := awsProviders[id]
	if !exists {
		awsProvidersMu.RUnlock()
		return "Disconnected", ErrProviderNotFound
	}
	provider := state.Provider
	awsProvidersMu.RUnlock()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	testErr := provider.TestConnection(ctx)
	awsProvidersMu.Lock()
	state, exists = awsProviders[id]
	if !exists {
		awsProvidersMu.Unlock()
		return "Disconnected", ErrProviderNotFound
	}
	if testErr != nil {
		state.Status = "Error"
		state.Error = testErr.Error()
		awsProvidersMu.Unlock()
		return "Error", testErr
	}
	state.Status = "Connected"
	state.Error = ""
	awsProvidersMu.Unlock()
	return "Connected", nil
}
func RefreshAWSProvider(id string) (*AWSProviderState, error) {
	log.Infof("RefreshAWSProvider called for id=%s", id)
	awsProvidersMu.Lock()
	state, exists := awsProviders[id]
	if !exists {
		awsProvidersMu.Unlock()
		log.Warnf("RefreshAWSProvider: provider not found id=%s", id)
		return nil, ErrProviderNotFound
	}
	state.Status = "Discovering"
	log.Infof("RefreshAWSProvider: provider found, id=%s, region=%s, authMethod=%s",
		state.Config.ID, state.Config.Region, state.Config.AuthMethod)
	awsProvidersMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	registry := providers.GetDefaultRegistry()
	log.Infof("RefreshAWSProvider: calling registry.RefreshDiscovery for id=%s", id)
	conns, err := registry.RefreshDiscovery(ctx, id)
	log.Infof("RefreshAWSProvider: registry.RefreshDiscovery returned %d connections, err=%v", len(conns), err)
	if len(conns) > 0 {
		awsprovider.CheckConnectivity(conns)
	}
	awsProvidersMu.Lock()
	defer awsProvidersMu.Unlock()
	state, exists = awsProviders[id]
	if !exists {
		return nil, ErrProviderNotFound
	}
	now := time.Now()
	state.LastDiscoveryAt = &now
	state.DiscoveredCount = len(conns)
	if err != nil {
		if len(conns) > 0 {
			state.Status = "Connected"
			state.Error = err.Error()
			return state, nil
		}
		state.Status = "Error"
		state.Error = err.Error()
		return state, err
	}
	state.Status = "Connected"
	state.Error = ""
	return state, nil
}
func configToProviderConfig(cfg *AWSProviderConfig) *awsprovider.Config {
	return &awsprovider.Config{
		ID:                  cfg.ID,
		Name:                cfg.Name,
		Region:              cfg.Region,
		AuthMethod:          awsinfra.AuthMethod(cfg.AuthMethod),
		ProfileName:         cfg.ProfileName,
		DiscoverRDS:         cfg.DiscoverRDS,
		DiscoverElastiCache: cfg.DiscoverElastiCache,
		DiscoverDocumentDB:  cfg.DiscoverDocumentDB,
	}
}
func GenerateProviderID(name, region string) string {
	return "aws-" + sanitizeID(name) + "-" + region
}
func InitAWSProvidersFromEnv() error {
	if !env.IsAWSProviderEnabled {
		return nil
	}
	envConfigs, err := envconfig.GetAWSProvidersFromEnv()
	if err != nil {
		return err
	}
	if len(envConfigs) == 0 {
		return nil
	}
	for i, envCfg := range envConfigs {
		name := envCfg.Name
		if name == "" {
			name = fmt.Sprintf("AWS-%d", i+1)
		}
		id := GenerateProviderID(name, envCfg.Region)
		authMethod := "default"
		if envCfg.ProfileName != "" {
			authMethod = "profile"
		}
		discoverRDS := true
		if envCfg.DiscoverRDS != nil {
			discoverRDS = *envCfg.DiscoverRDS
		}
		discoverElastiCache := true
		if envCfg.DiscoverElastiCache != nil {
			discoverElastiCache = *envCfg.DiscoverElastiCache
		}
		discoverDocumentDB := true
		if envCfg.DiscoverDocumentDB != nil {
			discoverDocumentDB = *envCfg.DiscoverDocumentDB
		}
		cfg := &AWSProviderConfig{
			ID:                  id,
			Name:                name,
			Region:              envCfg.Region,
			AuthMethod:          authMethod,
			ProfileName:         envCfg.ProfileName,
			DiscoverRDS:         discoverRDS,
			DiscoverElastiCache: discoverElastiCache,
			DiscoverDocumentDB:  discoverDocumentDB,
		}
		if _, err := AddAWSProvider(cfg); err != nil {
			log.Warnf("Failed to initialize AWS provider %s: %v", name, err)
		} else {
			log.Infof("Initialized AWS provider: %s (%s)", name, envCfg.Region)
		}
	}
	return nil
}
func sanitizeID(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		}
	}
	return string(result)
}