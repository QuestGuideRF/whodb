package aws
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"golang.org/x/sync/errgroup"
	awsinfra "github.com/clidey/whodb/core/src/aws"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/providers"
)
const maxPaginationPages = 1000
const discoveryTimeout = 60 * time.Second
type DiscoveryExtension func(ctx context.Context, p *Provider) ([]providers.DiscoveredConnection, error)
var discoveryExtensions []DiscoveryExtension
func SetDiscoveryExtension(ext DiscoveryExtension) {
	discoveryExtensions = append(discoveryExtensions, ext)
}
func (p *Provider) GetAWSConfig() aws.Config {
	return p.awsConfig
}
type Config struct {
	ID string
	Name string
	Region string
	AuthMethod awsinfra.AuthMethod
	ProfileName string
	DiscoverRDS bool
	DiscoverElastiCache bool
	DiscoverDocumentDB bool
	DiscoveryTimeout time.Duration
}
func (c *Config) String() string {
	return fmt.Sprintf("Config{ID:%s, Name:%s, Region:%s, AuthMethod:%s, ProfileName:%s, DiscoverRDS:%t, DiscoverElastiCache:%t, DiscoverDocumentDB:%t}",
		c.ID, c.Name, c.Region, c.AuthMethod, c.ProfileName, c.DiscoverRDS, c.DiscoverElastiCache, c.DiscoverDocumentDB)
}
func (c *Config) Validate() error {
	if c.Region == "" {
		return errors.New("aws: region is required")
	}
	if !c.DiscoverRDS && !c.DiscoverElastiCache && !c.DiscoverDocumentDB {
		log.Warnf("AWS provider %s: no discovery flags enabled", c.ID)
	}
	return nil
}
func DefaultConfig(id, name, region string) *Config {
	return &Config{
		ID:                  id,
		Name:                name,
		Region:              region,
		AuthMethod:          awsinfra.AuthMethodDefault,
		DiscoverRDS:         true,
		DiscoverElastiCache: true,
		DiscoverDocumentDB:  true,
	}
}
type Provider struct {
	config    *Config
	awsConfig aws.Config
	rdsClient         *rds.Client
	elasticacheClient *elasticache.Client
	docdbClient       *docdb.Client
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
	return providers.ProviderTypeAWS
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
	cfg, err := awsinfra.LoadAWSConfig(ctx, creds)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}
	p.awsConfig = cfg
	if p.config.DiscoverRDS {
		p.rdsClient = rds.NewFromConfig(cfg)
	}
	if p.config.DiscoverElastiCache {
		p.elasticacheClient = elasticache.NewFromConfig(cfg)
	}
	if p.config.DiscoverDocumentDB {
		p.docdbClient = docdb.NewFromConfig(cfg)
	}
	p.initialized = true
	return nil
}
func (p *Provider) buildInternalCredentials() *engine.Credentials {
	creds := &engine.Credentials{
		Hostname: p.config.Region,
	}
	var advanced []engine.Record
	advanced = append(advanced, engine.Record{
		Key:   awsinfra.AdvancedKeyAuthMethod,
		Value: string(p.config.AuthMethod),
	})
	if p.config.ProfileName != "" {
		advanced = append(advanced, engine.Record{
			Key:   awsinfra.AdvancedKeyProfileName,
			Value: p.config.ProfileName,
		})
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
	log.Infof("AWS Provider DiscoverConnections called for id=%s, region=%s, authMethod=%s, profileName=%s",
		p.config.ID, p.config.Region, p.config.AuthMethod, p.config.ProfileName)
	if err := p.initialize(ctx); err != nil {
		log.Errorf("AWS Provider initialize failed: %v", err)
		return nil, err
	}
	log.Infof("AWS Provider initialized successfully, rdsClient=%v", p.rdsClient != nil)
	taskCount := len(discoveryExtensions)
	if p.config.DiscoverRDS && p.rdsClient != nil {
		taskCount += 3
	}
	if p.config.DiscoverElastiCache && p.elasticacheClient != nil {
		taskCount++
	}
	taskCount++
	if p.config.DiscoverDocumentDB && p.docdbClient != nil {
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
	if p.config.DiscoverRDS && p.rdsClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverRDS(svcCtx)
			results <- discoveryResult{conns, err, "RDS"}
			return nil
		})
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverRDSClusters(svcCtx)
			results <- discoveryResult{conns, err, "RDS Clusters"}
			return nil
		})
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverRDSProxies(svcCtx)
			results <- discoveryResult{conns, err, "RDS Proxies"}
			return nil
		})
	}
	if p.config.DiscoverElastiCache && p.elasticacheClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverElastiCache(svcCtx)
			results <- discoveryResult{conns, err, "ElastiCache"}
			return nil
		})
	}
	g.Go(func() error {
		svcCtx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()
		conns, err := p.discoverOpenSearch(svcCtx)
		results <- discoveryResult{conns, err, "OpenSearch"}
		return nil
	})
	if p.config.DiscoverDocumentDB && p.docdbClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverDocumentDB(svcCtx)
			results <- discoveryResult{conns, err, "DocumentDB"}
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
			log.Errorf("AWS Provider: %s discovery failed: %v", r.name, r.err)
			allErrs = append(allErrs, fmt.Errorf("%s: %w", r.name, r.err))
		} else {
			log.Infof("AWS Provider: %s discovery found %d resources", r.name, len(r.conns))
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
	stsClient := sts.NewFromConfig(p.awsConfig)
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := stsClient.GetCallerIdentity(validateCtx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return awsinfra.HandleAWSError(err)
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
func (p *Provider) connectionID(resourceID string) string {
	return fmt.Sprintf("%s/%s", p.config.ID, resourceID)
}