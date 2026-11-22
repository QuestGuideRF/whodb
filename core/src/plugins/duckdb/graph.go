package duckdb
import "gorm.io/gorm"
const graphQuery = `
SELECT DISTINCT
    dc.referenced_table AS table1,
    dc.table_name AS table2,
    'OneToMany' AS relation,
    unnest(dc.constraint_column_names) AS source_column,
    unnest(dc.referenced_column_names) AS target_column
FROM duckdb_constraints() dc
WHERE dc.constraint_type = 'FOREIGN KEY'
    AND dc.schema_name = ?
`
func (p *DuckDBPlugin) GetGraphQueryDB(db *gorm.DB, schema string) *gorm.DB {
	return db.Raw(graphQuery, schema)
}