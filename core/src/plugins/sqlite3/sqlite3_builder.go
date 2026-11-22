package sqlite3
import (
	"fmt"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
type SQLiteSQLBuilder struct {
	*gorm_plugin.SQLBuilder
}
func NewSQLiteSQLBuilder(db *gorm.DB, plugin gorm_plugin.GormPluginFunctions) *SQLiteSQLBuilder {
	baseBuilder := gorm_plugin.NewSQLBuilder(db, plugin)
	builder := &SQLiteSQLBuilder{
		SQLBuilder: baseBuilder,
	}
	baseBuilder.SetSelf(builder)
	return builder
}
func (sb *SQLiteSQLBuilder) PragmaQuery(pragma, table string) (string, error) {
	allowedPragmas := map[string]bool{
		"table_info":       true,
		"table_xinfo":      true,
		"index_list":       true,
		"index_info":       true,
		"foreign_key_list": true,
		"table_list":       true,
	}
	if !allowedPragmas[pragma] {
		return "", fmt.Errorf("disallowed pragma: %s", pragma)
	}
	return fmt.Sprintf("PRAGMA %s(%s)", pragma, sb.QuoteIdentifier(table)), nil
}