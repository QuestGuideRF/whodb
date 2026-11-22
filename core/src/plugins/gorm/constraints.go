package gorm_plugin
import (
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
)
func (p *GormPlugin) GetColumnConstraints(config *engine.PluginConfig, schema string, storageUnit string) (map[string]map[string]any, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (map[string]map[string]any, error) {
		fullTableName := p.FormTableName(schema, storageUnit)
		migrator := NewMigratorHelper(db, p.GormPluginFunctions)
		migratorConstraints, err := migrator.GetConstraints(fullTableName)
		if err != nil {
			return make(map[string]map[string]any), nil
		}
		constraints := make(map[string]map[string]any)
		for constraintType, columns := range migratorConstraints {
			for _, col := range columns {
				columnName := col.Name()
				if constraints[columnName] == nil {
					constraints[columnName] = make(map[string]any)
				}
				switch constraintType {
				case "PRIMARY":
					constraints[columnName]["primary"] = true
				case "UNIQUE":
					constraints[columnName]["unique"] = true
				case "NOT_NULL":
					constraints[columnName]["nullable"] = false
				}
				constraints[columnName]["type"] = col.DatabaseTypeName()
				if nullable, ok := col.Nullable(); ok {
					constraints[columnName]["nullable"] = nullable
				}
				if unique, ok := col.Unique(); ok {
					constraints[columnName]["unique"] = unique
				}
				if autoIncrement, ok := col.AutoIncrement(); ok && autoIncrement {
					constraints[columnName]["auto_increment"] = true
				}
				if defaultValue, ok := col.DefaultValue(); ok && defaultValue != "" {
					constraints[columnName]["default"] = defaultValue
				}
				if comment, ok := col.Comment(); ok && comment != "" {
					constraints[columnName]["comment"] = comment
				}
				if length, ok := col.Length(); ok && length > 0 {
					constraints[columnName]["length"] = length
				}
				if precision, scale, ok := col.DecimalSize(); ok {
					constraints[columnName]["precision"] = precision
					if scale > 0 {
						constraints[columnName]["scale"] = scale
					}
				}
			}
		}
		if len(constraints) == 0 {
			return p.getColumnConstraintsRaw(db, schema, storageUnit)
		}
		return constraints, nil
	})
}
func (p *GormPlugin) getColumnConstraintsRaw(db *gorm.DB, schema string, storageUnit string) (map[string]map[string]any, error) {
	return make(map[string]map[string]any), nil
}
func (p *GormPlugin) clearTableDataWithDB(db *gorm.DB, schema string, storageUnit string) error {
	builder := p.GormPluginFunctions.CreateSQLBuilder(db)
	tableName := builder.BuildFullTableName(schema, storageUnit)
	result := db.Table(tableName).Where("1=1").Delete(nil)
	return result.Error
}
func (p *GormPlugin) NullifyFKColumn(config *engine.PluginConfig, schema string, storageUnit string, column string) error {
	_, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		builder := p.GormPluginFunctions.CreateSQLBuilder(db)
		tableName := builder.BuildFullTableName(schema, storageUnit)
		result := db.Table(tableName).Where(column+" IS NOT NULL").Update(column, gorm.Expr("NULL"))
		return result.Error == nil, result.Error
	})
	return err
}
func (p *GormPlugin) ClearTableData(config *engine.PluginConfig, schema string, storageUnit string) (bool, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		err := p.clearTableDataWithDB(db, schema, storageUnit)
		return err == nil, err
	})
}