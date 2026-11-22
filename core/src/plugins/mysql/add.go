package mysql
import (
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
func (p *MySQLPlugin) GetCreateTableQuery(db *gorm.DB, schema string, storageUnit string, columns []engine.Record) string {
	builder := p.GormPluginFunctions.CreateSQLBuilder(db)
	columnDefs := gorm_plugin.RecordsToColumnDefs(columns, func(def gorm_plugin.ColumnDef, column engine.Record) gorm_plugin.ColumnDef {
		def.Primary = true
		if strings.Contains(strings.ToLower(column.Value), "int") {
			def.Extra = "AUTO_INCREMENT"
		}
		return def
	})
	return builder.CreateTableQuery(schema, storageUnit, columnDefs)
}