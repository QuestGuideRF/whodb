package redis
import (
	"net"
	"strconv"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/go-redis/redis/v8"
)
func DB(config *engine.PluginConfig) (*redis.Client, error) {
	ctx := config.OperationContext()
	port, err := strconv.Atoi(common.GetRecordValueOrDefault(config.Credentials.Advanced, "Port", "6379"))
	if err != nil {
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to parse Redis port number")
		return nil, err
	}
	database := 0
	if config.Credentials.Database != "" {
		var err error
		database, err = strconv.Atoi(config.Credentials.Database)
		if err != nil {
			log.WithError(err).WithField("database", config.Credentials.Database).WithField("hostname", config.Credentials.Hostname).Error("Failed to parse Redis database number")
			return nil, err
		}
	}
	addr := net.JoinHostPort(config.Credentials.Hostname, strconv.Itoa(port))
	opts := &redis.Options{
		Addr:     addr,
		Username: config.Credentials.Username,
		Password: config.Credentials.Password,
		DB:       database,
	}
	sslMode := "disabled"
	sslConfig := ssl.ParseSSLConfig(engine.DatabaseType_Redis, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
	if sslConfig != nil && sslConfig.IsEnabled() {
		sslMode = string(sslConfig.Mode)
		tlsConfig, err := ssl.BuildTLSConfig(sslConfig, config.Credentials.Hostname)
		if err != nil {
			log.WithError(err).WithFields(map[string]any{
				"hostname": config.Credentials.Hostname,
				"sslMode":  sslConfig.Mode,
			}).Error("Failed to build TLS configuration for Redis")
			return nil, err
		}
		opts.TLSConfig = tlsConfig
	}
	client := redis.NewClient(opts)
	if _, err := client.Ping(ctx).Result(); err != nil {
		log.WithError(err).WithFields(map[string]any{
			"hostname": config.Credentials.Hostname,
			"database": database,
			"sslMode":  sslMode,
		}).Error("Failed to ping Redis server")
		return nil, err
	}
	return client, nil
}