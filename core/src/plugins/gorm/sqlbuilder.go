package gorm_plugin
import (
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
type SQLBuilderInterface interface {
	QuoteIdentifier(identifier string) string
	BuildFullTableName(schema, table string) string
	GetTableQuery(schema, table string) *gorm.DB
	SelectQuery(schema, table string, columns []string, conditions map[string]any) *gorm.DB
	BuildOrderBy(query *gorm.DB, sortList []plugins.Sort) *gorm.DB
	CreateTableQuery(schema, table string, columns []ColumnDef) string
	CreateTableQueryWithSuffix(schema, table string, columns []ColumnDef, suffix string) string
	InsertRow(schema, table string, data map[string]any) error
	UpdateQuery(schema, table string, updates map[string]any, conditions map[string]any) *gorm.DB
	DeleteQuery(schema, table string, conditions map[string]any) *gorm.DB
	CountQuery(schema, table string) (int64, error)
}
type SQLBuilder struct {
	db     *gorm.DB
	plugin GormPluginFunctions
	self   SQLBuilderInterface
}
func NewSQLBuilder(db *gorm.DB, plugin GormPluginFunctions) *SQLBuilder {
	sb := &SQLBuilder{
		db:     db,
		plugin: plugin,
	}
	sb.self = sb
	return sb
}
func (sb *SQLBuilder) SetSelf(self SQLBuilderInterface) {
	sb.self = self
}
func (sb *SQLBuilder) GetDB() *gorm.DB {
	return sb.db
}
func (sb *SQLBuilder) QuoteIdentifier(identifier string) string {
	if sb.db != nil && sb.db.Dialector != nil {
		var b strings.Builder
		sb.db.Dialector.QuoteTo(&b, identifier)
		return b.String()
	}
	return identifier
}
func (sb *SQLBuilder) BuildFullTableName(schema, table string) string {
	if table == "" {
		return ""
	}
	if schema == "" {
		return table
	}
	return schema + "." + table
}
func (sb *SQLBuilder) GetTableQuery(schema, table string) *gorm.DB {
	fullTableName := sb.BuildFullTableName(schema, table)
	return sb.db.Table(fullTableName)
}
func (sb *SQLBuilder) SelectQuery(schema, table string, columns []string, conditions map[string]any) *gorm.DB {
	query := sb.self.GetTableQuery(schema, table)
	if len(columns) > 0 {
		query = query.Select(columns)
	}
	if len(conditions) > 0 {
		query = query.Where(conditions)
	}
	return query
}
func (sb *SQLBuilder) BuildOrderBy(query *gorm.DB, sortList []plugins.Sort) *gorm.DB {
	for _, sort := range sortList {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: sort.Column},
			Desc:   sort.Direction == plugins.Down,
		})
	}
	return query
}
func (sb *SQLBuilder) CreateTableQuery(schema, table string, columns []ColumnDef) string {
	columnDefs := make([]string, len(columns))
	var primaryKeys []string
	for i, col := range columns {
		def := sb.QuoteIdentifier(col.Name) + " " + col.Type
		if col.Primary {
			primaryKeys = append(primaryKeys, sb.QuoteIdentifier(col.Name))
		}
		if col.Extra != "" {
			def += " " + col.Extra
		} else if col.NotNull || (!col.Nullable && !col.Primary) {
			def += " NOT NULL"
		}
		columnDefs[i] = def
	}
	if len(primaryKeys) > 0 {
		columnDefs = append(columnDefs, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}
	var fullTableName string
	if schema == "" {
		fullTableName = sb.QuoteIdentifier(table)
	} else {
		fullTableName = sb.QuoteIdentifier(schema) + "." + sb.QuoteIdentifier(table)
	}
	return fmt.Sprintf("CREATE TABLE %s (%s)", fullTableName, strings.Join(columnDefs, ", "))
}
func (sb *SQLBuilder) CreateTableQueryWithSuffix(schema, table string, columns []ColumnDef, suffix string) string {
	baseQuery := sb.CreateTableQuery(schema, table, columns)
	if suffix != "" {
		return baseQuery + " " + suffix
	}
	return baseQuery
}
func (sb *SQLBuilder) InsertRow(schema, table string, data map[string]any) error {
	if table == "" {
		return fmt.Errorf("table name cannot be empty when inserting row")
	}
	tableName := table
	if schema != "" {
		tableName = schema + "." + table
	}
	result := sb.db.Table(tableName).Create(data)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func (sb *SQLBuilder) UpdateQuery(schema, table string, updates map[string]any, conditions map[string]any) *gorm.DB {
	tableName := table
	if schema != "" {
		tableName = schema + "." + table
	}
	query := sb.db.Table(tableName)
	if len(conditions) > 0 {
		query = query.Where(conditions)
	}
	return query.Updates(updates)
}
func (sb *SQLBuilder) DeleteQuery(schema, table string, conditions map[string]any) *gorm.DB {
	tableName := table
	if schema != "" {
		tableName = schema + "." + table
	}
	query := sb.db.Table(tableName)
	if len(conditions) > 0 {
		query = query.Where(conditions)
	}
	return query.Delete(nil)
}
func (sb *SQLBuilder) CountQuery(schema, table string) (int64, error) {
	var count int64
	tableName := table
	if schema != "" {
		tableName = schema + "." + table
	}
	err := sb.db.Table(tableName).Count(&count).Error
	return count, err
}
type ColumnDef struct {
	Name     string
	Type     string
	Primary  bool
	Nullable bool
	NotNull  bool
	Extra    string
}
type PrimaryKeyDecorator func(def ColumnDef, column engine.Record) ColumnDef
func RecordsToColumnDefs(columns []engine.Record, decorator PrimaryKeyDecorator) []ColumnDef {
	defs := make([]ColumnDef, len(columns))
	for i, column := range columns {
		def := ColumnDef{
			Name: column.Key,
			Type: column.Value,
		}
		if primary, ok := column.Extra["primary"]; ok && primary == "true" {
			def = decorator(def, column)
		} else {
			if nullable, ok := column.Extra["nullable"]; ok && nullable == "false" {
				def.NotNull = true
			}
		}
		defs[i] = def
	}
	return defs
}