package sqlite3
import (
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *Sqlite3Plugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	log.Debug("[SSL] Sqlite3Plugin.GetSSLStatus: SQLite does not support SSL (local file-based)")
	return nil, nil
}