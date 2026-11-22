package plugins
import (
	"fmt"
	"time"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)
type SortDirection int
const (
	Up   SortDirection = iota
	Down
)
type Sort struct {
	Column    string
	Direction SortDirection
}
type AtomicCondition struct {
	Key        string
	Operator   string
	Value      any
	ColumnType string
}
type DBOperation[T any] func(*gorm.DB) (T, error)
type DBCreationFunc func(pluginConfig *engine.PluginConfig) (*gorm.DB, error)
func ConfigureConnectionPool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	return nil
}
func GetGormLogConfig() logger.LogLevel {
	switch log.GetLevel() {
	case "warning":
		return logger.Warn
	case "error":
		return logger.Error
	default:
		return logger.Silent
	}
}
func WithConnection[T any](config *engine.PluginConfig, DB DBCreationFunc, operation DBOperation[T]) (T, error) {
	if config != nil && config.Transaction != nil {
		if tx, ok := config.Transaction.(*gorm.DB); ok {
			return operation(tx.WithContext(config.OperationContext()))
		}
	}
	if config != nil && config.MultiStatement {
		db, err := DB(config)
		if err != nil {
			var zero T
			return zero, err
		}
		defer closeGormDB(db)
		return operation(db.WithContext(config.OperationContext()))
	}
	db, err := getOrCreateConnection(config, DB)
	if err != nil {
		log.WithFields(map[string]any{
			"conn_id": connIdentifier(config),
			"error":   err.Error(),
		}).Error("WithConnection FAILED to get connection")
		var zero T
		return zero, err
	}
	if db == nil {
		var zero T
		return zero, fmt.Errorf("internal error: nil database connection")
	}
	return operation(db.WithContext(config.OperationContext()))
}
func GetCachedSSLStatus(config *engine.PluginConfig) *engine.SSLStatus {
	key := getConnectionCacheKey(config)
	secret := getConnectionCacheSecret(config)
	connectionCacheMu.Lock()
	defer connectionCacheMu.Unlock()
	if cached, found := getCachedConnectionLocked(key, secret); found {
		return cached.sslStatus
	}
	return nil
}
func SetCachedSSLStatus(config *engine.PluginConfig, status *engine.SSLStatus) {
	key := getConnectionCacheKey(config)
	secret := getConnectionCacheSecret(config)
	connectionCacheMu.Lock()
	defer connectionCacheMu.Unlock()
	if cached, found := getCachedConnectionLocked(key, secret); found {
		cached.sslStatus = status
	}
}