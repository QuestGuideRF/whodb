package mysql
import (
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
)
func (p *MySQLPlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	if cached := plugins.GetCachedSSLStatus(config); cached != nil {
		return cached, nil
	}
	status, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (*engine.SSLStatus, error) {
		var result struct {
			VariableName string `gorm:"column:Variable_name"`
			Value        string `gorm:"column:Value"`
		}
		if err := db.Raw("SHOW SESSION STATUS LIKE 'Ssl_cipher'").Scan(&result).Error; err != nil {
			return nil, err
		}
		if result.Value == "" {
			return &engine.SSLStatus{
				IsEnabled: false,
				Mode:      string(ssl.SSLModeDisabled),
			}, nil
		}
		sslConfig := ssl.ParseSSLConfig(p.Type, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
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