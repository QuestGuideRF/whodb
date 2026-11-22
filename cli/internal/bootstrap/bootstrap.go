package bootstrap
import (
	"os"
	"sync"
	"github.com/clidey/whodb/cli/pkg/identity"
	"github.com/clidey/whodb/core/src/log"
	_ "github.com/clidey/whodb/core/src/bamlinit"
	_ "github.com/clidey/whodb/core/src/plugins"
	_ "github.com/clidey/whodb/core/src/plugins/clickhouse"
	_ "github.com/clidey/whodb/core/src/plugins/elasticsearch"
	_ "github.com/clidey/whodb/core/src/plugins/memcached"
	_ "github.com/clidey/whodb/core/src/plugins/mongodb"
	_ "github.com/clidey/whodb/core/src/plugins/mysql"
	_ "github.com/clidey/whodb/core/src/plugins/postgres"
	_ "github.com/clidey/whodb/core/src/plugins/redis"
	_ "github.com/clidey/whodb/core/src/plugins/sqlite3"
	_ "github.com/clidey/whodb/core/src/source/adapters"
	_ "github.com/clidey/whodb/core/src/sources/database"
)
var ensureOnce sync.Once
func Ensure() {
	ensureOnce.Do(func() {
		os.Setenv(identity.Current().RuntimeEnv, "true")
		log.DisableOutput()
		log.SetLogLevel("none")
	})
}