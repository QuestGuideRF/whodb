package duckdb
import (
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *DuckDBPlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	log.Debug("[SSL] DuckDBPlugin.GetSSLStatus: DuckDB does not support SSL (local file-based)")
	return nil, nil
}