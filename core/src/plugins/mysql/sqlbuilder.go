package mysql
import (
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
type MySQLSQLBuilder struct {
	*gorm_plugin.SQLBuilder
}
func NewMySQLSQLBuilder(db *gorm.DB, plugin gorm_plugin.GormPluginFunctions) gorm_plugin.SQLBuilderInterface {
	msb := &MySQLSQLBuilder{
		SQLBuilder: gorm_plugin.NewSQLBuilder(db, plugin),
	}
	msb.SQLBuilder.SetSelf(msb)
	return msb
}
func (msb *MySQLSQLBuilder) GetTableQuery(schema, table string) *gorm.DB {
	return msb.SQLBuilder.GetDB().Table(table).Clauses(clause.From{
		Tables: []clause.Table{
			{Name: table},
		},
	})
}