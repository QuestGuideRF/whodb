package mongodb
import (
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *MongoDBPlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	log.Debug("[SSL] MongoDBPlugin.GetSSLStatus: checking configured SSL mode")
	sslConfig := ssl.ParseSSLConfig(engine.DatabaseType_MongoDB, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
	if sslConfig == nil || !sslConfig.IsEnabled() {
		log.Debug("[SSL] MongoDBPlugin.GetSSLStatus: SSL not configured or disabled")
		return &engine.SSLStatus{
			IsEnabled: false,
			Mode:      string(ssl.SSLModeDisabled),
		}, nil
	}
	log.Debugf("[SSL] MongoDBPlugin.GetSSLStatus: SSL enabled, mode=%s", sslConfig.Mode)
	return &engine.SSLStatus{
		IsEnabled: true,
		Mode:      string(sslConfig.Mode),
	}, nil
}