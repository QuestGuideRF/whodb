package sqlite3
import (
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
func (p *Sqlite3Plugin) GetCreateTableQuery(db *gorm.DB, schema string, storageUnit string, columns []engine.Record) string {
	builder := gorm_plugin.NewSQLBuilder(db, p)
	columnDefs := gorm_plugin.RecordsToColumnDefs(columns, func(def gorm_plugin.ColumnDef, column engine.Record) gorm_plugin.ColumnDef {
		def.Primary = true
		if strings.Contains(strings.ToLower(column.Value), "int") {
			def.Type = "INTEGER"
		}
		return def
	})
	return builder.CreateTableQuery("", storageUnit, columnDefs)
}