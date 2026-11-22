package duckdb
import (
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
func (p *DuckDBPlugin) GetColumnConstraints(config *engine.PluginConfig, schema string, storageUnit string) (map[string]map[string]any, error) {
	constraints := make(map[string]map[string]any)
	_, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		rows, err := db.Raw(`
			SELECT column_name, is_nullable
			FROM information_schema.columns
			WHERE table_schema = ? AND table_name = ?
		`, schema, storageUnit).Rows()
		if err != nil {
			return false, err
		}
		defer rows.Close()
		for rows.Next() {
			var columnName, isNullable string
			if err := rows.Scan(&columnName, &isNullable); err != nil {
				continue
			}
			gorm_plugin.EnsureConstraintEntry(constraints, columnName)["nullable"] = isNullable == "YES"
		}
		constraintRows, err := db.Raw(`
			SELECT unnest(dc.constraint_column_names) AS column_name, dc.constraint_type
			FROM duckdb_constraints() dc
			WHERE dc.constraint_type IN ('PRIMARY KEY', 'UNIQUE')
				AND dc.schema_name = ? AND dc.table_name = ?
		`, schema, storageUnit).Rows()
		if err == nil {
			defer constraintRows.Close()
			for constraintRows.Next() {
				var columnName, constraintType string
				if err := constraintRows.Scan(&columnName, &constraintType); err != nil {
					continue
				}
				switch constraintType {
				case "PRIMARY KEY":
					gorm_plugin.EnsureConstraintEntry(constraints, columnName)["primary"] = true
				case "UNIQUE":
					gorm_plugin.EnsureConstraintEntry(constraints, columnName)["unique"] = true
				}
			}
		}
		return true, nil
	})
	if err != nil {
		return constraints, err
	}
	return constraints, nil
}