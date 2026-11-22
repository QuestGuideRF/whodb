package elasticsearch
import (
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *ElasticSearchPlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	log.Debug("[SSL] ElasticSearchPlugin.GetSSLStatus: checking configured SSL mode")
	sslConfig := ssl.ParseSSLConfig(engine.DatabaseType_ElasticSearch, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
	if sslConfig == nil || !sslConfig.IsEnabled() {
		log.Debug("[SSL] ElasticSearchPlugin.GetSSLStatus: SSL not configured or disabled")
		return &engine.SSLStatus{
			IsEnabled: false,
			Mode:      string(ssl.SSLModeDisabled),
		}, nil
	}
	log.Debugf("[SSL] ElasticSearchPlugin.GetSSLStatus: SSL enabled, mode=%s", sslConfig.Mode)
	return &engine.SSLStatus{
		IsEnabled: true,
		Mode:      string(sslConfig.Mode),
	}, nil
}