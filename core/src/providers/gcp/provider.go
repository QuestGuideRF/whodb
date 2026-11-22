package gcp
import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	alloydb "cloud.google.com/go/alloydb/apiv1"
	memcache "cloud.google.com/go/memcache/apiv1"
	redis "cloud.google.com/go/redis/apiv1"
	gcpinfra "github.com/clidey/whodb/core/src/gcp"
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
type Config struct {
	ID string
	Name string
	ProjectID string
	Region string
	AuthMethod gcpinfra.AuthMethod
	ServiceAccountKeyPath string
	DiscoverCloudSQL bool
	DiscoverAlloyDB bool
	DiscoverMemorystore bool
	DiscoveryTimeout time.Duration
}
func (c *Config) String() string {
	return fmt.Sprintf("Config{ID:%s, Name:%s, ProjectID:%s, Region:%s, AuthMethod:%s, DiscoverCloudSQL:%t, DiscoverAlloyDB:%t, DiscoverMemorystore:%t}",
		c.ID, c.Name, c.ProjectID, c.Region, c.AuthMethod, c.DiscoverCloudSQL, c.DiscoverAlloyDB, c.DiscoverMemorystore)
}
func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return errors.New("gcp: project ID is required")
	}
	if c.Region == "" {
		return errors.New("gcp: region is required")
	}
	if !c.DiscoverCloudSQL && !c.DiscoverAlloyDB && !c.DiscoverMemorystore {
		log.Warnf("GCP provider %s: no discovery flags enabled", c.ID)
	}
	return nil
}
func DefaultConfig(id, name, projectID, region string) *Config {
	return &Config{
		ID:                  id,
		Name:                name,
		ProjectID:           projectID,
		Region:              region,
		AuthMethod:          gcpinfra.AuthMethodDefault,
		DiscoverCloudSQL:    true,
		DiscoverAlloyDB:     true,
		DiscoverMemorystore: true,
	}
}
type Provider struct {
	config *Config
	clientOpts []option.ClientOption
	sqladminService   *sqladmin.Service
	alloydbClient     *alloydb.AlloyDBAdminClient
	memorystoreClient *redis.CloudRedisClient
	memcachedClient   *memcache.CloudMemcacheClient
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
	return providers.ProviderTypeGCP
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
	var opts []option.ClientOption
	if p.config.AuthMethod == gcpinfra.AuthMethodServiceAccountKey && p.config.ServiceAccountKeyPath != "" {
		opts = append(opts, option.WithCredentialsFile(p.config.ServiceAccountKeyPath))
	}
	p.clientOpts = opts
	if p.config.DiscoverCloudSQL {
		svc, err := sqladmin.NewService(ctx, opts...)
		if err != nil {
			return fmt.Errorf("failed to create Cloud SQL client: %w", gcpinfra.HandleGCPError(err))
		}
		p.sqladminService = svc
	}
	if p.config.DiscoverAlloyDB {
		client, err := alloydb.NewAlloyDBAdminClient(ctx, opts...)
		if err != nil {
			return fmt.Errorf("failed to create AlloyDB client: %w", gcpinfra.HandleGCPError(err))
		}
		p.alloydbClient = client
	}
	if p.config.DiscoverMemorystore {
		client, err := redis.NewCloudRedisClient(ctx, opts...)
		if err != nil {
			return fmt.Errorf("failed to create Memorystore client: %w", gcpinfra.HandleGCPError(err))
		}
		p.memorystoreClient = client
		mcClient, err := memcache.NewCloudMemcacheClient(ctx, opts...)
		if err != nil {
			log.Warnf("GCP Provider: failed to create Memcached client (non-fatal): %v", err)
		} else {
			p.memcachedClient = mcClient
		}
	}
	p.initialized = true
	return nil
}
type discoveryResult struct {
	conns []providers.DiscoveredConnection
	err   error
	name  string
}
func (p *Provider) DiscoverConnections(ctx context.Context) ([]providers.DiscoveredConnection, error) {
	log.Infof("GCP Provider DiscoverConnections called for id=%s, projectID=%s, region=%s, authMethod=%s",
		p.config.ID, p.config.ProjectID, p.config.Region, p.config.AuthMethod)
	if err := p.initialize(ctx); err != nil {
		log.Errorf("GCP Provider initialize failed: %v", err)
		return nil, err
	}
	log.Infof("GCP Provider initialized successfully")
	taskCount := len(discoveryExtensions)
	if p.config.DiscoverCloudSQL && p.sqladminService != nil {
		taskCount++
	}
	if p.config.DiscoverAlloyDB && p.alloydbClient != nil {
		taskCount++
	}
	if p.config.DiscoverMemorystore && p.memorystoreClient != nil {
		taskCount++
	}
	if p.config.DiscoverMemorystore && p.memcachedClient != nil {
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
	if p.config.DiscoverCloudSQL && p.sqladminService != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverCloudSQL(svcCtx)
			results <- discoveryResult{conns, err, "Cloud SQL"}
			return nil
		})
	}
	if p.config.DiscoverAlloyDB && p.alloydbClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverAlloyDB(svcCtx)
			results <- discoveryResult{conns, err, "AlloyDB"}
			return nil
		})
	}
	if p.config.DiscoverMemorystore && p.memorystoreClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverMemorystore(svcCtx)
			results <- discoveryResult{conns, err, "Memorystore"}
			return nil
		})
	}
	if p.config.DiscoverMemorystore && p.memcachedClient != nil {
		g.Go(func() error {
			svcCtx, cancel := context.WithTimeout(gctx, timeout)
			defer cancel()
			conns, err := p.discoverMemcached(svcCtx)
			results <- discoveryResult{conns, err, "Memcached"}
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
			log.Errorf("GCP Provider: %s discovery failed: %v", r.name, r.err)
			allErrs = append(allErrs, fmt.Errorf("%s: %w", r.name, r.err))
		} else {
			log.Infof("GCP Provider: %s discovery found %d resources", r.name, len(r.conns))
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
	crmService, err := cloudresourcemanager.NewService(ctx, p.clientOpts...)
	if err != nil {
		return gcpinfra.HandleGCPError(err)
	}
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	projectName := "projects/" + p.config.ProjectID
	_, err = crmService.Projects.Get(projectName).Context(validateCtx).Do()
	if err != nil {
		return gcpinfra.HandleGCPError(err)
	}
	return nil
}
func (p *Provider) RefreshConnection(ctx context.Context, connectionID string) (bool, error) {
	return false, nil
}
func (p *Provider) Close(ctx context.Context) error {
	p.initMu.Lock()
	defer p.initMu.Unlock()
	if p.alloydbClient != nil {
		p.alloydbClient.Close()
	}
	if p.memorystoreClient != nil {
		p.memorystoreClient.Close()
	}
	if p.memcachedClient != nil {
		p.memcachedClient.Close()
	}
	return nil
}
func (p *Provider) GetConfig() *Config {
	return p.config
}
func (p *Provider) GetClientOpts() []option.ClientOption {
	return p.clientOpts
}
func (p *Provider) connectionID(resourceID string) string {
	return fmt.Sprintf("%s/%s", p.config.ID, resourceID)
}