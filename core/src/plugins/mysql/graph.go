package mysql
import "gorm.io/gorm"
const graphQuery = `
SELECT DISTINCT
    rc.REFERENCED_TABLE_NAME AS Table1,
    rc.TABLE_NAME AS Table2,
    'OneToMany' AS Relation,
    kcu.COLUMN_NAME AS SourceColumn,
    kcu.REFERENCED_COLUMN_NAME AS TargetColumn
FROM
    INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS rc
    JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
        ON rc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
        AND rc.CONSTRAINT_SCHEMA = kcu.CONSTRAINT_SCHEMA
WHERE
    rc.CONSTRAINT_SCHEMA = ?
`
func (p *MySQLPlugin) GetGraphQueryDB(db *gorm.DB, schema string) *gorm.DB {
	return db.Raw(graphQuery, schema)
}