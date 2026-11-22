package postgres
import (
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
)
func (p *PostgresPlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	if cached := plugins.GetCachedSSLStatus(config); cached != nil {
		return cached, nil
	}
	status, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (*engine.SSLStatus, error) {
		var result struct {
			SSL bool `gorm:"column:ssl"`
		}
		query := db.Raw("SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid()").Scan(&result)
		if query.Error != nil {
			return nil, query.Error
		}
		if query.RowsAffected == 0 {
			sslConfig := ssl.ParseSSLConfig(engine.DatabaseType(p.Type), config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
			if sslConfig != nil && sslConfig.IsEnabled() {
				return &engine.SSLStatus{
					IsEnabled: true,
					Mode:      string(sslConfig.Mode),
				}, nil
			}
			return &engine.SSLStatus{
				IsEnabled: false,
				Mode:      string(ssl.SSLModeDisabled),
			}, nil
		}
		if !result.SSL {
			return &engine.SSLStatus{
				IsEnabled: false,
				Mode:      string(ssl.SSLModeDisabled),
			}, nil
		}
		sslConfig := ssl.ParseSSLConfig(engine.DatabaseType(p.Type), config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
		mode := "enabled"
		if sslConfig != nil {
			mode = string(sslConfig.Mode)
		}
		return &engine.SSLStatus{
			IsEnabled: true,
			Mode:      mode,
		}, nil
	})
	if err == nil && status != nil {
		plugins.SetCachedSSLStatus(config, status)
	}
	return status, err
}