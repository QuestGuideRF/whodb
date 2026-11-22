package postgres
import (
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
func (p *PostgresPlugin) GetCreateTableQuery(db *gorm.DB, schema string, storageUnit string, columns []engine.Record) string {
	builder := gorm_plugin.NewSQLBuilder(db, p)
	columnDefs := gorm_plugin.RecordsToColumnDefs(columns, func(def gorm_plugin.ColumnDef, column engine.Record) gorm_plugin.ColumnDef {
		lowerType := strings.ToLower(column.Value)
		if strings.Contains(lowerType, "int") || strings.Contains(lowerType, "integer") {
			def.Extra = "PRIMARY KEY GENERATED ALWAYS AS IDENTITY"
		} else {
			def.Primary = true
		}
		return def
	})
	return builder.CreateTableQuery(schema, storageUnit, columnDefs)
}