package sqlite3
import (
	"gorm.io/gorm"
)
const graphQuery = `
SELECT DISTINCT
    f."table" AS table1,
    m.name AS table2,
    'OneToMany' AS relation,
    f."from" AS source_column,
    f."to" AS target_column
FROM
    sqlite_master m,
    pragma_foreign_key_list(m.name) f
WHERE
    m.type = 'table'
`
func (p *Sqlite3Plugin) GetGraphQueryDB(db *gorm.DB, schema string) *gorm.DB {
	return db.Raw(graphQuery)
}