package azure
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"golang.org/x/sync/errgroup"
	azureinfra "github.com/clidey/whodb/core/src/azure"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/providers"
)
const discoveryTimeout = 60 * time.Second
type DiscoveryExtension func(ctx context.Context, p *Provider) ([]providers.DiscoveredConnection, error)
var discoveryExtensions []DiscoveryExtension
func SetDiscoveryExtension(ext DiscoveryExtension) {
	discoveryExtensions = append(discoveryExtensions, ext)
}
var azureSQLEngineMapper func(string) (engine.DatabaseType, bool)
func SetAzureSQLEngineMapper(fn func(string) (engine.DatabaseType, bool)) {
	azureSQLEngineMapper = fn
}
type Config struct {
	ID string
	Name string
	SubscriptionID string
	TenantID string
	ClientID string
	ClientSecret string
	AuthMethod azureinfra.AuthMethod
	ResourceGroup string
	DiscoverPostgreSQL bool
	DiscoverMySQL bool
	DiscoverRedis bool
	DiscoverCosmosDB bool
	DiscoveryTimeout time.Duration
}
func (c *Config) String() string {
	return fmt.Sprintf("Config{ID:%s, Name:%s, SubscriptionID:%s, AuthMethod:%s, ResourceGroup:%s, DiscoverPostgreSQL:%t, DiscoverMySQL:%t, DiscoverRedis:%t, DiscoverCosmosDB:%t}",
		c.ID, c.Name, c.SubscriptionID, c.AuthMethod, c.ResourceGroup, c.DiscoverPostgreSQL, c.DiscoverMySQL, c.DiscoverRedis, c.DiscoverCosmosDB)
}
func (c *Config) Validate() error {
	if c.SubscriptionID == "" {
		return errors.New("azure: subscription ID is required")
	}
	if !c.DiscoverPostgreSQL && !c.DiscoverMySQL && !c.DiscoverRedis && !c.DiscoverCosmosDB {
		log.Warnf("Azure provider %s: no discovery flags enabled", c.ID)
	}
	return nil
}
func DefaultConfig(id, name, subscriptionID string) *Config {
	return &Config{
		ID:                 id,
		Name:               name,
		SubscriptionID:     subscriptionID,
		AuthMethod:         azureinfra.AuthMethodDefault,
		DiscoverPostgreSQL: true,
		DiscoverMySQL:      true,
		DiscoverRedis:      true,
		DiscoverCosmosDB:   true,
	}
}
type Provider struct {
	config     *Config
	credential azcore.TokenCredential
	postgresClient *armpostgresqlflexibleservers.ServersClient
	mysqlClient    *armmysqlflexibleservers.ServersClient
	redisClient    *armredis.Client
	cosmosClient   *armcosmos.DatabaseAccountsClient
	initMu      sync.Mutex
	initialized bool
}
func New(config *Config) (*Provider, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}
	if config.ID == "" {
		return nil, errors.New("provider ID is required")
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &Provider{
		config: config,
	}, nil
}
func (p *Provider) Type() providers.ProviderType {
	return providers.ProviderTypeAzure
}
func (p *Provider) ID() string {
	return p.config.ID
}
func (p *Provider) Name() string {
	return p.config.Name
}
func (p *Provider) initialize(ctx context.Context) error {
	p.initMu.Lock()
	defer p.initMu.Unlock()
	if p.initialized {
		return nil
	}
	creds := p.buildInternalCredentials()
	credential, _, err := azureinfra.LoadAzureCredential(creds)
	if err != nil {
		return fmt.Errorf("failed to load Azure credential: %w", err)
	}
	p.credential = credential
	if p.config.DiscoverPostgreSQL {
		client, err := armpostgresqlflexibleservers.NewServersClient(p.config.SubscriptionID, credential, nil)
		if err != nil {
			return fmt.Errorf("failed to create PostgreSQL client: %w", azureinfra.HandleAzureError(err))
		}
		p.postgresClient = client
	}
	if p.config.DiscoverMySQL {
		client, err := armmysqlflexibleservers.NewServersClient(p.config.SubscriptionID, credential, nil)
		if err != nil {
			return fmt.Errorf("failed to create MySQL client: %w", azureinfra.HandleAzureError(err))
		}
		p.mysqlClient = client
	}
	if p.config.DiscoverRedis {
		client, err := armredis.NewClient(p.config.SubscriptionID, credential, nil)
		if err != nil {
			return fmt.Errorf("failed to create Redis client: %w", azureinfra.HandleAzureError(err))
		}
		p.redisClient = client
	}
	if p.config.DiscoverCosmosDB {
		client, err := armcosmos.NewDatabaseAccountsClient(p.config.SubscriptionID, credential, nil)
		if err != nil {
			return fmt.Errorf("failed to create Cosmos DB client: %w", azureinfra.HandleAzureError(err))
		}
		p.cosmosClient = client
	}
	p.initialized = true
	return nil
}
func (p *Provider) buildInternalCredentials() *engine.Credentials {
	creds := &engine.Credentials{
		Hostname: p.config.SubscriptionID,
	}
	var advanced []engine.Record
	advanced = append(advanced, engine.Record{
		Key:   azureinfra.AdvancedKeyAuthMethod,
		Value: string(p.config.AuthMethod),
	})
	if p.config.TenantID != "" {
		advanced = append(advanced, engine.Record{Key: azureinfra.AdvancedKeyTenantID, Value: p.config.TenantID})
	}
	if p.config.ClientID != "" {
		advanced = append(advanced, engine.Record{Key: azureinfra.AdvancedKeyClientID, Value: p.config.ClientID})
	}
	if p.config.ClientSecret != "" {
		advanced = append(advanced, engine.Record{Key: azureinfra.AdvancedKeyClientSecret, Value: p.config.ClientSecret})
	}
	if p.config.ResourceGroup != "" {
		advanced = append(advanced, engine.Record{Key: azureinfra.AdvancedKeyResourceGroup, Value: p.config.ResourceGroup})
	}
	creds.Advanced = advanced
	return creds
}
type discoveryResult struct {
	conns []providers.DiscoveredConnection
	err   error
	name  string
}
func (p *Provider) DiscoverConnections(ctx context.Context) ([]providers.DiscoveredConnection, error) {
	log.Infof("Azure Provider DiscoverConnections called for id=%s, subscription=%s, authMethod=%s",
		p.config.ID, p.config.SubscriptionID, p.config.AuthMethod)
	if err := p.initialize(ctx); err != nil {
		log.Errorf("Azure Provider initialize failed: %v", err)
		return nil, err
	}
	taskCount := len(discoveryExtensions)
	if p.config.DiscoverPostgreSQL && p.postgresClient != nil {
		taskCount++
	}
	if p.config.DiscoverMySQL && p.mysqlClient != nil {
		taskCount++
	}
	if p.config.DiscoverRedis && p.redisClient != nil {
		taskCount++
	}
	if p.config.DiscoverCosmosDB && p.cosmosClient != nil {
		taskCount++
	}
	if taskCount == 0 {
		return nil, nil
	}
	timeout := discoveryTimeout
	if p.config.DiscoveryTimeout > 0 {
		timeout = p.config.DiscoveryTimeout
	}
	results := make(chan discoveryResult, taskCount)
	g, gctx := errgroup.WithContext(ctx)
	if p.config.DiscoverPostgreSQL && p.postgresClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverPostgreSQL(svcCtx)
			results <- discoveryResult{conns, err, "PostgreSQL"}
			return nil
		})
	}
	if p.config.DiscoverMySQL && p.mysqlClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverMySQL(svcCtx)
			results <- discoveryResult{conns, err, "MySQL"}
			return nil
		})
	}
	if p.config.DiscoverRedis && p.redisClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverRedis(svcCtx)
			results <- discoveryResult{conns, err, "Redis"}
			return nil
		})
	}
	if p.config.DiscoverCosmosDB && p.cosmosClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverCosmosDB(svcCtx)
			results <- discoveryResult{conns, err, "CosmosDB"}
			return nil
		})
	}
	for _, ext := range discoveryExtensions {
		ext := ext
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := ext(svcCtx, p)
			results <- discoveryResult{conns, err, "extension"}
			return nil
		})
	}
	_ = g.Wait()
	close(results)
	var allConns []providers.DiscoveredConnection
	var allErrs []error
	for r := range results {
		if r.err != nil {
			log.Errorf("Azure Provider: %s discovery failed: %v", r.name, r.err)
			allErrs = append(allErrs, fmt.Errorf("%s: %w", r.name, r.err))
		} else {
			log.Infof("Azure Provider: %s discovery found %d resources", r.name, len(r.conns))
			allConns = append(allConns, r.conns...)
		}
	}
	if len(allErrs) > 0 {
		return allConns, errors.Join(allErrs...)
	}
	return allConns, nil
}
func (p *Provider) TestConnection(ctx context.Context) error {
	if err := p.initialize(ctx); err != nil {
		return err
	}
	subsClient, err := armsubscriptions.NewClient(p.credential, nil)
	if err != nil {
		return azureinfra.HandleAzureError(err)
	}
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err = subsClient.Get(validateCtx, p.config.SubscriptionID, nil)
	if err != nil {
		return azureinfra.HandleAzureError(err)
	}
	return nil
}
func (p *Provider) RefreshConnection(ctx context.Context, connectionID string) (bool, error) {
	return false, nil
}
func (p *Provider) Close(ctx context.Context) error {
	return nil
}
func (p *Provider) GetConfig() *Config {
	return p.config
}
func (p *Provider) GetAzureCredential() azcore.TokenCredential {
	return p.credential
}
func (p *Provider) GetSubscriptionID() string {
	return p.config.SubscriptionID
}
func (p *Provider) GetResourceGroup() string {
	return p.config.ResourceGroup
}
func (p *Provider) connectionID(resourceID string) string {
	return fmt.Sprintf("%s/%s", p.config.ID, resourceID)
}
func extractResourceGroup(resourceID string) string {
	if resourceID == "" {
		return ""
	}
	const prefix = "/resourcegroups/"
	lower := toLower(resourceID)
	idx := indexOf(lower, prefix)
	if idx == -1 {
		return ""
	}
	rest := resourceID[idx+len(prefix):]
	end := indexOf(rest, "/")
	if end == -1 {
		return rest
	}
	return rest[:end]
}
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}