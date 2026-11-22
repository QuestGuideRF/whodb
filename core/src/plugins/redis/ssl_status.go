package redis
import (
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *RedisPlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	log.Debug("[SSL] RedisPlugin.GetSSLStatus: checking configured TLS mode")
	sslConfig := ssl.ParseSSLConfig(engine.DatabaseType_Redis, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
	if sslConfig == nil || !sslConfig.IsEnabled() {
		log.Debug("[SSL] RedisPlugin.GetSSLStatus: TLS not configured or disabled")
		return &engine.SSLStatus{
			IsEnabled: false,
			Mode:      string(ssl.SSLModeDisabled),
		}, nil
	}
	log.Debugf("[SSL] RedisPlugin.GetSSLStatus: TLS enabled, mode=%s", sslConfig.Mode)
	return &engine.SSLStatus{
		IsEnabled: true,
		Mode:      string(sslConfig.Mode),
	}, nil
}