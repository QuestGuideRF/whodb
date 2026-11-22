package providers
import (
	"context"
	"errors"
	"sync"
	"time"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
type cachedDiscovery struct {
	connections  []DiscoveredConnection
	discoveredAt time.Time
}
var (
	ErrProviderNotFound = errors.New("provider not found")
	ErrProviderAlreadyExists = errors.New("provider already exists")
	ErrConnectionNotFound = errors.New("connection not found")
)
var DefaultDiscoveryCacheTTL = 5 * time.Minute
type Registry struct {
	mu        sync.RWMutex
	providers map[string]ConnectionProvider
	cacheMu   sync.RWMutex
	connCache map[string]*cachedDiscovery
}
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]ConnectionProvider),
		connCache: make(map[string]*cachedDiscovery),
	}
}
func (r *Registry) Register(provider ConnectionProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := provider.ID()
	if _, exists := r.providers[id]; exists {
		return ErrProviderAlreadyExists
	}
	r.providers[id] = provider
	return nil
}
func (r *Registry) Unregister(providerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	provider, exists := r.providers[providerID]
	if !exists {
		return ErrProviderNotFound
	}
	if err := provider.Close(context.Background()); err != nil {
		log.Warnf("Provider %s close failed during unregister: %v", providerID, err)
	}
	delete(r.providers, providerID)
	r.cacheMu.Lock()
	delete(r.connCache, providerID)
	r.cacheMu.Unlock()
	return nil
}
func (r *Registry) Get(providerID string) (ConnectionProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	provider, exists := r.providers[providerID]
	if !exists {
		return nil, ErrProviderNotFound
	}
	return provider, nil
}
func (r *Registry) GetByType(providerType ProviderType) []ConnectionProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []ConnectionProvider
	for _, p := range r.providers {
		if p.Type() == providerType {
			result = append(result, p)
		}
	}
	return result
}
func (r *Registry) List() []ConnectionProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ConnectionProvider, 0, len(r.providers))
	for _, p := range r.providers {
		result = append(result, p)
	}
	return result
}
func (r *Registry) DiscoverAll(ctx context.Context) ([]DiscoveredConnection, error) {
	r.mu.RLock()
	providers := make([]ConnectionProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	r.mu.RUnlock()
	log.Infof("DiscoverAll: found %d registered providers", len(providers))
	var allConns []DiscoveredConnection
	var errs []error
	for _, p := range providers {
		providerID := p.ID()
		r.cacheMu.RLock()
		cached, hasCached := r.connCache[providerID]
		r.cacheMu.RUnlock()
		if hasCached && time.Since(cached.discoveredAt) < DefaultDiscoveryCacheTTL {
			age := time.Since(cached.discoveredAt).Round(time.Second)
			log.Infof("DiscoverAll: using %d cached connections for provider %s (age=%s, ttl=%s)", len(cached.connections), providerID, age, DefaultDiscoveryCacheTTL)
			allConns = append(allConns, cached.connections...)
			continue
		}
		if hasCached {
			age := time.Since(cached.discoveredAt).Round(time.Second)
			log.Infof("DiscoverAll: cache expired for provider %s (age=%s, ttl=%s), re-discovering", providerID, age, DefaultDiscoveryCacheTTL)
		} else {
			log.Infof("DiscoverAll: no cache for provider %s, discovering", providerID)
		}
		conns, err := p.DiscoverConnections(ctx)
		if len(conns) > 0 {
			log.Infof("DiscoverAll: got %d connections from provider %s", len(conns), providerID)
			r.cacheMu.Lock()
			r.connCache[providerID] = &cachedDiscovery{connections: conns, discoveredAt: time.Now()}
			r.cacheMu.Unlock()
			allConns = append(allConns, conns...)
		}
		if err != nil {
			log.Warnf("DiscoverAll: error from provider %s: %v", providerID, err)
			errs = append(errs, err)
		}
	}
	log.Infof("DiscoverAll: returning %d total connections", len(allConns))
	if len(errs) > 0 {
		return allConns, errors.Join(errs...)
	}
	return allConns, nil
}
func (r *Registry) RefreshDiscovery(ctx context.Context, providerID string) ([]DiscoveredConnection, error) {
	if providerID == "" {
		r.cacheMu.Lock()
		r.connCache = make(map[string]*cachedDiscovery)
		r.cacheMu.Unlock()
		return r.DiscoverAll(ctx)
	}
	r.cacheMu.Lock()
	delete(r.connCache, providerID)
	r.cacheMu.Unlock()
	provider, err := r.Get(providerID)
	if err != nil {
		return nil, err
	}
	conns, err := provider.DiscoverConnections(ctx)
	if err != nil {
		if len(conns) > 0 {
			r.cacheMu.Lock()
			r.connCache[providerID] = &cachedDiscovery{connections: conns, discoveredAt: time.Now()}
			r.cacheMu.Unlock()
		}
		return conns, err
	}
	r.cacheMu.Lock()
	r.connCache[providerID] = &cachedDiscovery{connections: conns, discoveredAt: time.Now()}
	r.cacheMu.Unlock()
	return conns, nil
}
func (r *Registry) FilterByDatabaseType(ctx context.Context, dbType engine.DatabaseType) ([]DiscoveredConnection, error) {
	all, err := r.DiscoverAll(ctx)
	if err != nil {
		return nil, err
	}
	var result []DiscoveredConnection
	for _, c := range all {
		if c.DatabaseType == dbType {
			result = append(result, c)
		}
	}
	return result, nil
}
func (r *Registry) FilterByProvider(ctx context.Context, providerID string) ([]DiscoveredConnection, error) {
	all, err := r.DiscoverAll(ctx)
	if err != nil {
		return nil, err
	}
	var result []DiscoveredConnection
	for _, c := range all {
		if c.ProviderID == providerID {
			result = append(result, c)
		}
	}
	return result, nil
}
func (r *Registry) FilterAvailable(ctx context.Context) ([]DiscoveredConnection, error) {
	all, err := r.DiscoverAll(ctx)
	if err != nil {
		return nil, err
	}
	var result []DiscoveredConnection
	for _, c := range all {
		if c.Status.IsAvailable() {
			result = append(result, c)
		}
	}
	return result, nil
}
func (r *Registry) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var errs []error
	for _, p := range r.providers {
		if err := p.Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	r.providers = make(map[string]ConnectionProvider)
	r.cacheMu.Lock()
	r.connCache = make(map[string]*cachedDiscovery)
	r.cacheMu.Unlock()
	return errors.Join(errs...)
}
var (
	defaultRegistry     *Registry
	defaultRegistryOnce sync.Once
)
func GetDefaultRegistry() *Registry {
	defaultRegistryOnce.Do(func() {
		defaultRegistry = NewRegistry()
	})
	return defaultRegistry
}