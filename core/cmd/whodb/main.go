package main
import (
	_ "github.com/clidey/whodb/core/src/bamlinit"
	"github.com/clidey/whodb/core/graph"
	"github.com/clidey/whodb/core/src/app"
	_ "github.com/clidey/whodb/core/src/plugins/clickhouse"
	_ "github.com/clidey/whodb/core/src/plugins/elasticsearch"
	_ "github.com/clidey/whodb/core/src/plugins/memcached"
	_ "github.com/clidey/whodb/core/src/plugins/mongodb"
	_ "github.com/clidey/whodb/core/src/plugins/mysql"
	_ "github.com/clidey/whodb/core/src/plugins/postgres"
	_ "github.com/clidey/whodb/core/src/plugins/redis"
	_ "github.com/clidey/whodb/core/src/plugins/sqlite3"
)
func main() {
	app.Run(app.AppConfig{
		Schema: graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}),
	}, staticFiles)
}