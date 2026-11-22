package mysql
import (
	"fmt"
	"net"
	"strconv"
	"time"
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)
func (p *MySQLPlugin) DB(config *engine.PluginConfig) (*gorm.DB, error) {
	return p.openDB(config, false)
}
func (p *MySQLPlugin) openDB(config *engine.PluginConfig, multiStatements bool) (*gorm.DB, error) {
	connectionInput, err := p.ParseConnectionConfig(config)
	if err != nil {
		return nil, err
	}
	mysqlConfig := mysqldriver.NewConfig()
	mysqlConfig.User = connectionInput.Username
	mysqlConfig.Passwd = connectionInput.Password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = net.JoinHostPort(connectionInput.Hostname, strconv.Itoa(connectionInput.Port))
	mysqlConfig.DBName = connectionInput.Database
	mysqlConfig.AllowCleartextPasswords = connectionInput.AllowClearTextPasswords
	mysqlConfig.ParseTime = connectionInput.ParseTime
	mysqlConfig.Loc = connectionInput.Loc
	mysqlConfig.Timeout = time.Duration(connectionInput.ConnectionTimeout) * time.Second
	mysqlConfig.Params = connectionInput.ExtraOptions
	mysqlConfig.MultiStatements = multiStatements
	sslMode := "disabled"
	if connectionInput.SSLConfig != nil && connectionInput.SSLConfig.IsEnabled() {
		sslMode = string(connectionInput.SSLConfig.Mode)
		if connectionInput.SSLConfig.Mode == ssl.SSLModePreferred {
			mysqlConfig.TLSConfig = "preferred"
		} else {
			tlsConfig, err := ssl.BuildTLSConfig(connectionInput.SSLConfig, connectionInput.Hostname)
			if err != nil {
				log.WithError(err).WithFields(map[string]any{
					"hostname": connectionInput.Hostname,
					"sslMode":  connectionInput.SSLConfig.Mode,
				}).Error("Failed to build TLS configuration for MySQL")
				return nil, err
			}
			configName := fmt.Sprintf("whodb_%s_%d", connectionInput.Database, time.Now().UnixNano())
			if err := mysqldriver.RegisterTLSConfig(configName, tlsConfig); err != nil {
				log.WithError(err).WithField("configName", configName).Error("Failed to register TLS config for MySQL")
				return nil, err
			}
			mysqlConfig.TLSConfig = configName
		}
	}
	l := log.WithFields(map[string]any{
		"hostname": connectionInput.Hostname,
		"port":     connectionInput.Port,
		"database": connectionInput.Database,
		"username": connectionInput.Username,
		"sslMode":  sslMode,
	})
	db, err := gorm.Open(mysql.Open(mysqlConfig.FormatDSN()), &gorm.Config{Logger: logger.Default.LogMode(plugins.GetGormLogConfig())})
	if err != nil {
		l.WithError(err).Error("Failed to connect to MySQL database")
		return nil, err
	}
	if err := plugins.ConfigureConnectionPool(db); err != nil {
		l.WithError(err).Warn("Failed to configure connection pool")
	}
	return db, nil
}