package clickhouse
import (
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
)
func (p *ClickHousePlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	if cached := plugins.GetCachedSSLStatus(config); cached != nil {
		return cached, nil
	}
	sslConfig := ssl.ParseSSLConfig(engine.DatabaseType_ClickHouse, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
	var status *engine.SSLStatus
	if sslConfig == nil || !sslConfig.IsEnabled() {
		status = &engine.SSLStatus{
			IsEnabled: false,
			Mode:      string(ssl.SSLModeDisabled),
		}
	} else {
		status = &engine.SSLStatus{
			IsEnabled: true,
			Mode:      string(sslConfig.Mode),
		}
	}
	plugins.SetCachedSSLStatus(config, status)
	return status, nil
}