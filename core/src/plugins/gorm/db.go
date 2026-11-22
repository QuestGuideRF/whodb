package gorm_plugin
import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
)
const (
	portKey                    = "Port"
	parseTimeKey               = "Parse Time"
	locKey                     = "Loc"
	allowClearTextPasswordsKey = "Allow clear text passwords"
	httpProtocolKey            = "HTTP Protocol"
	readOnlyKey                = "Readonly"
	debugKey                   = "Debug"
	connectionTimeoutKey       = "Connection Timeout"
	searchPathKey              = "Search Path"
)
type ConnectionInput struct {
	Username string `validate:"required"`
	Password string `validate:"required"`
	Database string `validate:"required"`
	Hostname string `validate:"required"`
	Port     int    `validate:"required"`
	ParseTime               bool           `validate:"boolean"`
	Loc                     *time.Location `validate:"required"`
	AllowClearTextPasswords bool           `validate:"boolean"`
	SearchPath string
	HTTPProtocol string
	ReadOnly     string
	Debug        string
	ConnectionTimeout int
	SSLConfig *ssl.SSLConfig
	ExtraOptions map[string]string `validate:"omitnil"`
}
func (p *GormPlugin) ParseConnectionConfig(config *engine.PluginConfig) (*ConnectionInput, error) {
	port := 0
	if rawPort := common.GetRecordValueOrDefault(config.Credentials.Advanced, portKey, ""); rawPort != "" {
		parsedPort, err := strconv.Atoi(rawPort)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to parse port for database type %s", p.Type))
			return nil, err
		}
		port = parsedPort
	}
	parseTime, err := strconv.ParseBool(common.GetRecordValueOrDefault(config.Credentials.Advanced, parseTimeKey, "True"))
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Failed to parse parseTime setting for database type %s", p.Type))
		return nil, err
	}
	loc, err := time.LoadLocation(common.GetRecordValueOrDefault(config.Credentials.Advanced, locKey, "UTC"))
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Failed to load time location for database type %s", p.Type))
		return nil, err
	}
	allowClearTextPasswords, err := strconv.ParseBool(common.GetRecordValueOrDefault(config.Credentials.Advanced, allowClearTextPasswordsKey, "0"))
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Failed to parse allowClearTextPasswords setting for database type %s", p.Type))
		return nil, err
	}
	httpProtocol := common.GetRecordValueOrDefault(config.Credentials.Advanced, httpProtocolKey, "disable")
	readOnly := common.GetRecordValueOrDefault(config.Credentials.Advanced, readOnlyKey, "disable")
	debug := common.GetRecordValueOrDefault(config.Credentials.Advanced, debugKey, "disable")
	searchPath := common.GetRecordValueOrDefault(config.Credentials.Advanced, searchPathKey, "")
	sslConfig := ssl.ParseSSLConfig(p.Type, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
	connectionTimeout, err := strconv.Atoi(common.GetRecordValueOrDefault(config.Credentials.Advanced, connectionTimeoutKey, "90"))
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Failed to parse connection timeout for database type %s", p.Type))
		return nil, err
	}
	database := config.Credentials.Database
	username := config.Credentials.Username
	password := config.Credentials.Password
	hostname := config.Credentials.Hostname
	input := &ConnectionInput{
		Username:                username,
		Password:                password,
		Database:                database,
		Hostname:                hostname,
		Port:                    port,
		ParseTime:               parseTime,
		Loc:                     loc,
		AllowClearTextPasswords: allowClearTextPasswords,
		SearchPath:              searchPath,
		HTTPProtocol:            httpProtocol,
		ReadOnly:                readOnly,
		Debug:                   debug,
		ConnectionTimeout:       connectionTimeout,
		SSLConfig:               sslConfig,
	}
	if config.Credentials.IsProfile {
		params := make(map[string]string)
		for _, record := range config.Credentials.Advanced {
			switch record.Key {
			case portKey, parseTimeKey, locKey, allowClearTextPasswordsKey, httpProtocolKey, readOnlyKey, debugKey, connectionTimeoutKey, searchPathKey,
				ssl.KeySSLMode, ssl.KeySSLCACertContent, ssl.KeySSLClientCertContent, ssl.KeySSLClientKeyContent,
				ssl.KeySSLCACertPath, ssl.KeySSLClientCertPath, ssl.KeySSLClientKeyPath, ssl.KeySSLServerName:
				continue
			default:
				if p.Type == engine.DatabaseType_Postgres {
					params[record.Key] = record.Value
				} else {
					params[record.Key] = url.QueryEscape(record.Value)
				}
			}
		}
		input.ExtraOptions = params
	}
	return input, nil
}
func (p *GormPlugin) IsAvailable(ctx context.Context, config *engine.PluginConfig) bool {
	available, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		return true, nil
	})
	if err != nil {
		return false
	}
	return available
}