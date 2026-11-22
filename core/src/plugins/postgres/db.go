package postgres
import (
	"maps"
	"time"
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)
func (p *PostgresPlugin) DB(config *engine.PluginConfig) (*gorm.DB, error) {
	return p.openDB(config, false)
}
func (p *PostgresPlugin) openDB(config *engine.PluginConfig, multiStatement bool) (*gorm.DB, error) {
	connectionInput, err := p.ParseConnectionConfig(config)
	if err != nil {
		return nil, err
	}
	pgxConfig, err := pgx.ParseConfig("")
	if err != nil {
		log.WithError(err).WithFields(map[string]any{
			"hostname": connectionInput.Hostname,
			"port":     connectionInput.Port,
			"database": connectionInput.Database,
			"username": connectionInput.Username,
		}).Error("Failed to parse PostgreSQL configuration")
		return nil, err
	}
	pgxConfig.Host = connectionInput.Hostname
	pgxConfig.Port = uint16(connectionInput.Port)
	pgxConfig.User = connectionInput.Username
	pgxConfig.Password = connectionInput.Password
	pgxConfig.Database = connectionInput.Database
	pgxConfig.ConnectTimeout = time.Duration(connectionInput.ConnectionTimeout) * time.Second
	if multiStatement {
		pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	}
	sslMode := "disabled"
	if connectionInput.SSLConfig != nil && connectionInput.SSLConfig.IsEnabled() {
		sslMode = string(connectionInput.SSLConfig.Mode)
		tlsConfig, err := ssl.BuildTLSConfig(connectionInput.SSLConfig, connectionInput.Hostname)
		if err != nil {
			log.WithError(err).WithFields(map[string]any{
				"hostname": connectionInput.Hostname,
				"sslMode":  connectionInput.SSLConfig.Mode,
			}).Error("Failed to build TLS configuration for PostgreSQL")
			return nil, err
		}
		pgxConfig.TLSConfig = tlsConfig
	}
	if connectionInput.ExtraOptions != nil || connectionInput.SearchPath != "" {
		if pgxConfig.RuntimeParams == nil {
			pgxConfig.RuntimeParams = make(map[string]string)
		}
		maps.Copy(pgxConfig.RuntimeParams, connectionInput.ExtraOptions)
		if connectionInput.SearchPath != "" {
			pgxConfig.RuntimeParams["search_path"] = connectionInput.SearchPath
		}
	}
	l := log.WithFields(map[string]any{
		"hostname": connectionInput.Hostname,
		"port":     connectionInput.Port,
		"database": connectionInput.Database,
		"username": connectionInput.Username,
		"sslMode":  sslMode,
	})
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: stdlib.OpenDB(*pgxConfig)}), &gorm.Config{Logger: logger.Default.LogMode(plugins.GetGormLogConfig())})
	if err != nil {
		l.WithError(err).Error("Failed to connect to PostgreSQL database")
		return nil, err
	}
	if err := plugins.ConfigureConnectionPool(db); err != nil {
		l.WithError(err).Warn("Failed to configure connection pool")
	}
	return db, nil
}