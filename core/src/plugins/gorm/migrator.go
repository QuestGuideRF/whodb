package gorm_plugin
import (
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"gorm.io/gorm"
)
type MigratorHelper struct {
	db       *gorm.DB
	plugin   GormPluginFunctions
	migrator gorm.Migrator
}
func NewMigratorHelper(db *gorm.DB, plugin GormPluginFunctions) *MigratorHelper {
	return &MigratorHelper{
		db:       db,
		plugin:   plugin,
		migrator: db.Migrator(),
	}
}
func (m *MigratorHelper) TableExists(tableName string) bool {
	return m.migrator.HasTable(tableName)
}
func (m *MigratorHelper) GetConstraints(tableName string) (map[string][]gorm.ColumnType, error) {
	columnTypes, err := m.migrator.ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}
	constraints := make(map[string][]gorm.ColumnType)
	for _, col := range columnTypes {
		if primary, ok := col.PrimaryKey(); ok && primary {
			constraints["PRIMARY"] = append(constraints["PRIMARY"], col)
		}
		if unique, ok := col.Unique(); ok && unique {
			constraints["UNIQUE"] = append(constraints["UNIQUE"], col)
		}
		if nullable, ok := col.Nullable(); ok && !nullable {
			constraints["NOT_NULL"] = append(constraints["NOT_NULL"], col)
		}
	}
	return constraints, nil
}
func (m *MigratorHelper) GetColumnTypes(tableName string) (map[string]ColumnTypeInfo, error) {
	columnTypes := make(map[string]ColumnTypeInfo)
	types, err := m.migrator.ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}
	for _, col := range types {
		fullType := m.buildFullTypeName(col)
		normalizedType := m.plugin.NormalizeType(fullType)
		isNullable := true
		if nullable, ok := col.Nullable(); ok {
			isNullable = nullable
		}
		columnTypes[col.Name()] = ColumnTypeInfo{
			Type:       normalizedType,
			IsNullable: isNullable,
		}
	}
	return columnTypes, nil
}
type ColumnTypeInfo struct {
	Type       string
	IsNullable bool
}
var typesWithLength = map[string]bool{
	"VARCHAR": true, "CHAR": true, "CHARACTER": true, "CHARACTER VARYING": true,
	"NVARCHAR": true, "NCHAR": true, "BPCHAR": true,
	"BINARY": true, "VARBINARY": true, "BIT": true, "BIT VARYING": true, "VARBIT": true,
	"FIXEDSTRING": true,
}
var typesWithPrecision = map[string]bool{
	"DECIMAL": true, "NUMERIC": true, "NUMBER": true,
	"FLOAT": true, "DOUBLE": true, "REAL": true,
}
func (m *MigratorHelper) buildFullTypeName(col gorm.ColumnType) string {
	baseName := strings.ToUpper(col.DatabaseTypeName())
	if typesWithLength[baseName] {
		if length, ok := col.Length(); ok && length > 0 {
			return fmt.Sprintf("%s(%d)", baseName, length)
		}
	}
	if typesWithPrecision[baseName] {
		if precision, scale, ok := col.DecimalSize(); ok && precision > 0 {
			return fmt.Sprintf("%s(%d,%d)", baseName, precision, scale)
		}
	}
	return baseName
}
func (m *MigratorHelper) GetOrderedColumns(tableName string) ([]engine.Column, error) {
	types, err := m.migrator.ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}
	columns := make([]engine.Column, 0, len(types))
	for _, col := range types {
		fullType := m.buildFullTypeName(col)
		normalizedType := m.plugin.NormalizeType(fullType)
		baseName := strings.ToUpper(col.DatabaseTypeName())
		isAutoIncr, _ := col.AutoIncrement()
		isNullable := true
		if nullable, ok := col.Nullable(); ok {
			isNullable = nullable
		}
		column := engine.Column{
			Name:            col.Name(),
			Type:            normalizedType,
			IsNullable:      isNullable,
			IsAutoIncrement: isAutoIncr,
		}
		if typesWithLength[baseName] {
			if length, ok := col.Length(); ok && length > 0 {
				l := int(length)
				column.Length = &l
			}
		}
		if typesWithPrecision[baseName] {
			if precision, scale, ok := col.DecimalSize(); ok && precision > 0 {
				p := int(precision)
				column.Precision = &p
				s := int(scale)
				column.Scale = &s
			}
		}
		columns = append(columns, column)
	}
	return columns, nil
}