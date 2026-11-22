package plugins
import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"gorm.io/gorm"
)
type cachedConnection struct {
	db        *gorm.DB
	lastUsed  time.Time
	sslStatus *engine.SSLStatus
}
type connectionCacheBucket map[string]*cachedConnection
const connectionCacheTTL = 5 * time.Minute
const maxCachedConnections = 50
var (
	connectionCache   = make(map[string]connectionCacheBucket)
	connectionCacheMu sync.Mutex
	stopCleanup       = make(chan struct{})
)
func init() {
	startConnectionCleanup()
}
func startConnectionCleanup() {
	go func() {
		ticker := time.NewTicker(connectionCacheTTL / 5)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleanupStaleConnections()
			case <-stopCleanup:
				return
			}
		}
	}()
}
func cleanupStaleConnections() {
	log.Debug("cleaning up stale connections")
	staleThreshold := time.Now().Add(-connectionCacheTTL)
	connectionCacheMu.Lock()
	defer connectionCacheMu.Unlock()
	for key, bucket := range connectionCache {
		for secret, cached := range bucket {
			if cached.lastUsed.Before(staleThreshold) {
				delete(bucket, secret)
				closeGormDB(cached.db)
				log.Debug("Closed stale database connection")
			}
		}
		if len(bucket) == 0 {
			delete(connectionCache, key)
		}
	}
}
func closeGormDB(db *gorm.DB) {
	if db == nil {
		return
	}
	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			log.WithError(err).Error("failed to close db connection")
		}
	}
}
func connIdentifier(config *engine.PluginConfig) string {
	return fmt.Sprintf("%s:%s:%s", config.Credentials.Type, config.Credentials.Hostname, config.Credentials.Database)
}
func shortKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:4])
}
func getConnectionCacheKey(config *engine.PluginConfig) string {
	parts := []string{
		config.Credentials.Type,
		config.Credentials.Hostname,
		config.Credentials.Username,
		config.Credentials.Database,
		strconv.FormatBool(config.Credentials.IsProfile),
	}
	if config.Credentials.Id != nil {
		parts = append(parts, *config.Credentials.Id)
	}
	for _, adv := range config.Credentials.Advanced {
		parts = append(parts, adv.Key, adv.Value)
	}
	return strings.Join(parts, "\x00")
}
func getConnectionCacheSecret(config *engine.PluginConfig) string {
	return config.Credentials.Password
}
func getCachedConnectionLocked(key string, secret string) (*cachedConnection, bool) {
	bucket, found := connectionCache[key]
	if !found {
		return nil, false
	}
	cached, found := bucket[secret]
	return cached, found
}
func setCachedConnectionLocked(key string, secret string, cached *cachedConnection) {
	bucket, found := connectionCache[key]
	if !found {
		bucket = make(connectionCacheBucket)
		connectionCache[key] = bucket
	}
	bucket[secret] = cached
}
func deleteCachedConnectionLocked(key string, secret string) *cachedConnection {
	bucket, found := connectionCache[key]
	if !found {
		return nil
	}
	cached, found := bucket[secret]
	if !found {
		return nil
	}
	delete(bucket, secret)
	if len(bucket) == 0 {
		delete(connectionCache, key)
	}
	return cached
}
func connectionCacheEntryCountLocked() int {
	count := 0
	for _, bucket := range connectionCache {
		count += len(bucket)
	}
	return count
}
func RemoveConnection(config *engine.PluginConfig) {
	connID := connIdentifier(config)
	key := getConnectionCacheKey(config)
	secret := getConnectionCacheSecret(config)
	l := log.WithFields(map[string]any{"conn_id": connID, "cache_key": shortKey(key)})
	l.Debug("RemoveConnection called")
	connectionCacheMu.Lock()
	cached := deleteCachedConnectionLocked(key, secret)
	if cached != nil {
		connectionCacheMu.Unlock()
		closeGormDB(cached.db)
		l.Debug("Connection removed and closed")
		return
	}
	connectionCacheMu.Unlock()
	l.Debug("Connection not found in cache")
}
func CloseAllConnections(_ context.Context) {
	l := log.WithField("phase", "shutdown")
	l.Info("CloseAllConnections called, stopping cleanup goroutine")
	close(stopCleanup)
	connectionCacheMu.Lock()
	connCount := connectionCacheEntryCountLocked()
	l.WithField("conn_count", connCount).Info("Closing cached connections")
	for key, bucket := range connectionCache {
		for secret, cached := range bucket {
			l.WithField("cache_key", shortKey(key)).Debug("Closing connection")
			closeGormDB(cached.db)
			delete(bucket, secret)
		}
		delete(connectionCache, key)
	}
	connectionCacheMu.Unlock()
	l.Info("All connections closed")
}
func getOrCreateConnection(config *engine.PluginConfig, createDB DBCreationFunc) (*gorm.DB, error) {
	connID := connIdentifier(config)
	key := getConnectionCacheKey(config)
	secret := getConnectionCacheSecret(config)
	l := log.WithFields(map[string]any{"conn_id": connID, "cache_key": shortKey(key)})
	connectionCacheMu.Lock()
	if cached, found := getCachedConnectionLocked(key, secret); found && cached != nil && cached.db != nil {
		cached.lastUsed = time.Now()
		db := cached.db
		connectionCacheMu.Unlock()
		l.Debug("Cache HIT - reusing cached connection")
		return db, nil
	} else {
		connectionCacheMu.Unlock()
	}
	l.Debug("Creating NEW database connection")
	createStart := time.Now()
	db, err := createDB(config)
	if err != nil {
		l.WithFields(map[string]any{"duration_ms": time.Since(createStart).Milliseconds(), "error": err.Error()}).Error("Failed to create connection")
		return nil, err
	}
	l.WithField("duration_ms", time.Since(createStart).Milliseconds()).Info("Connection created successfully")
	connectionCacheMu.Lock()
	defer connectionCacheMu.Unlock()
	if cached, found := getCachedConnectionLocked(key, secret); found && cached != nil && cached.db != nil {
		cached.lastUsed = time.Now()
		l.Debug("using connection created by another goroutine")
		closeGormDB(db)
		return cached.db, nil
	}
	setCachedConnectionLocked(key, secret, &cachedConnection{
		db:       db,
		lastUsed: time.Now(),
	})
	if connectionCacheEntryCountLocked() > maxCachedConnections {
		evictOldestConnection(key, secret)
	}
	return db, nil
}
func evictOldestConnection(excludeKey string, excludeSecret string) {
	var oldestKey string
	var oldestSecret string
	var oldestTime = time.Now().Add(time.Hour)
	for key, bucket := range connectionCache {
		for secret, cached := range bucket {
			if key == excludeKey && secret == excludeSecret {
				continue
			}
			if cached.lastUsed.Before(oldestTime) {
				oldestTime = cached.lastUsed
				oldestKey = key
				oldestSecret = secret
			}
		}
	}
	if oldestKey != "" {
		cached := deleteCachedConnectionLocked(oldestKey, oldestSecret)
		closeGormDB(cached.db)
		log.WithField("cache_key", shortKey(oldestKey)).Debug("Evicted oldest connection to stay under limit")
	}
}