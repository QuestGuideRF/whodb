package duckdb
import (
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
func (p *DuckDBPlugin) GetCreateTableQuery(db *gorm.DB, schema string, storageUnit string, columns []engine.Record) string {
	builder := gorm_plugin.NewSQLBuilder(db, p)
	var seqSQL string
	columnDefs := gorm_plugin.RecordsToColumnDefs(columns, func(def gorm_plugin.ColumnDef, column engine.Record) gorm_plugin.ColumnDef {
		lowerType := strings.ToLower(column.Value)
		if strings.Contains(lowerType, "int") || strings.Contains(lowerType, "integer") {
			seqName := fmt.Sprintf("%s_%s_seq", storageUnit, column.Key)
			seqSQL = fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s; ", builder.QuoteIdentifier(seqName))
			def.Extra = fmt.Sprintf("PRIMARY KEY DEFAULT nextval('%s')", strings.ReplaceAll(seqName, "'", "''"))
		} else {
			def.Primary = true
		}
		return def
	})
	return seqSQL + builder.CreateTableQuery(schema, storageUnit, columnDefs)
}