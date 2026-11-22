package clickhouse
import (
	"errors"
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
var ErrUpsertNotSupported = errors.New("import.error.upsert_not_supported")
func (p *ClickHousePlugin) GetCreateTableQuery(db *gorm.DB, schema string, storageUnit string, columns []engine.Record) string {
	builder := gorm_plugin.NewSQLBuilder(db, p)
	var primaryKeys []string
	columnDefs := gorm_plugin.RecordsToColumnDefs(columns, func(def gorm_plugin.ColumnDef, column engine.Record) gorm_plugin.ColumnDef {
		primaryKeys = append(primaryKeys, column.Key)
		if nullable, ok := column.Extra["nullable"]; ok && nullable == "false" {
			def.NotNull = true
		}
		return def
	})
	orderByClause := ""
	if len(primaryKeys) > 0 {
		quotedKeys := make([]string, len(primaryKeys))
		for i, key := range primaryKeys {
			quotedKeys[i] = builder.QuoteIdentifier(key)
		}
		orderByClause = strings.Join(quotedKeys, ", ")
	} else if len(columns) > 0 {
		orderByClause = builder.QuoteIdentifier(columns[0].Key)
	}
	suffix := fmt.Sprintf("ENGINE = MergeTree() ORDER BY (%s)", orderByClause)
	return builder.CreateTableQueryWithSuffix(schema, storageUnit, columnDefs, suffix)
}
func (p *ClickHousePlugin) BulkAddRows(config *engine.PluginConfig, schema string, storageUnit string, rows [][]engine.Record) (bool, error) {
	if len(config.UpsertPKColumns) > 0 {
		return false, ErrUpsertNotSupported
	}
	return p.GormPlugin.BulkAddRows(config, schema, storageUnit, rows)
}