package src
import (
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins/duckdb"
)
func registerDuckDB(e *engine.Engine) {
	e.RegistryPlugin(duckdb.NewDuckDBPlugin())
}