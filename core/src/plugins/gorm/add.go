package gorm_plugin
import (
	"errors"
	"fmt"
	"strconv"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"github.com/clidey/whodb/core/src/sourcecatalog"
	"gorm.io/gorm"
)
func (p *GormPlugin) AddStorageUnit(config *engine.PluginConfig, schema string, storageUnit string, fields []engine.Record) (bool, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		if len(fields) == 0 {
			return false, errors.New("no fields provided for table creation")
		}
		migrator := NewMigratorHelper(db, p.GormPluginFunctions)
		fullTableName := p.FormTableName(schema, storageUnit)
		if migrator.TableExists(fullTableName) {
			return false, fmt.Errorf("table %s already exists", fullTableName)
		}
		var columns []engine.Record
		metadata, _ := sourcecatalog.ResolveSessionMetadata(string(p.Type))
		for _, fieldType := range fields {
			if err := engine.ValidateColumnType(fieldType.Value, string(p.Type), metadata); err != nil {
				return false, err
			}
			fieldName := fieldType.Key
			primaryKey, err := strconv.ParseBool(fieldType.Extra["Primary"])
			if err != nil {
				log.WithError(err).Error(fmt.Sprintf("Failed to parse Primary key flag for field %s in table %s.%s", fieldType.Key, schema, storageUnit))
				primaryKey = false
			}
			nullable, err := strconv.ParseBool(fieldType.Extra["Nullable"])
			if err != nil {
				log.WithError(err).Error(fmt.Sprintf("Failed to parse Nullable flag for field %s in table %s.%s", fieldType.Key, schema, storageUnit))
				nullable = false
			}
			if primaryKey && nullable {
				return false, fmt.Errorf("column %s cannot be both primary key and nullable", fieldType.Key)
			}
			columns = append(columns, engine.Record{
				Key:   fieldName,
				Value: fieldType.Value,
				Extra: map[string]string{
					"primary":  strconv.FormatBool(primaryKey),
					"nullable": strconv.FormatBool(nullable),
				},
			})
		}
		createTableQuery := p.GetCreateTableQuery(db, schema, storageUnit, columns)
		if err := db.Exec(createTableQuery).Error; err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to create table %s.%s with query: %s", schema, storageUnit, createTableQuery))
			return false, err
		}
		return true, nil
	})
}
func (p *GormPlugin) addRowWithDB(db *gorm.DB, schema string, storageUnit string, values []engine.Record) error {
	if len(values) == 0 {
		return errors.New("no values provided to insert into the table")
	}
	if storageUnit == "" {
		return fmt.Errorf("storage unit name cannot be empty")
	}
	columnTypeInfos, err := p.GormPluginFunctions.GetColumnTypes(db, schema, storageUnit)
	if err != nil {
		log.WithError(err).WithField("schema", schema).WithField("storageUnit", storageUnit).
			Warn("Failed to fetch column types, continuing without type information")
		columnTypeInfos = make(map[string]ColumnTypeInfo)
	}
	for i, value := range values {
		if values[i].Extra == nil {
			values[i].Extra = make(map[string]string)
		}
		if values[i].Extra["Type"] == "" {
			if colInfo, ok := columnTypeInfos[value.Key]; ok {
				values[i].Extra["Type"] = colInfo.Type
				if colInfo.IsNullable {
					values[i].Extra["IsNullable"] = "true"
				}
			}
		}
	}
	builder := p.GormPluginFunctions.CreateSQLBuilder(db)
	valuesToAdd, err := p.ConvertRecordValuesToMap(values)
	if err != nil {
		return err
	}
	return builder.InsertRow(schema, storageUnit, valuesToAdd)
}
func (p *GormPlugin) AddRow(config *engine.PluginConfig, schema string, storageUnit string, values []engine.Record) (bool, error) {
	if p.errorHandler == nil {
		p.InitPlugin()
	}
	if storageUnit == "" {
		log.Error("AddRow called with empty storageUnit name")
		return false, fmt.Errorf("storage unit name cannot be empty")
	}
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		err := p.addRowWithDB(db, schema, storageUnit, values)
		if err != nil {
			err = p.errorHandler.HandleError(err, "AddRow", map[string]any{
				"schema":      schema,
				"storageUnit": storageUnit,
				"valueCount":  len(values),
			})
		}
		return err == nil, err
	})
}
func (p *GormPlugin) AddRowReturningID(config *engine.PluginConfig, schema string, storageUnit string, values []engine.Record) (int64, error) {
	if p.errorHandler == nil {
		p.InitPlugin()
	}
	if storageUnit == "" {
		log.Error("AddRowReturningID called with empty storageUnit name")
		return 0, fmt.Errorf("storage unit name cannot be empty")
	}
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (int64, error) {
		var lastID int64
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := p.addRowWithDB(tx, schema, storageUnit, values); err != nil {
				return err
			}
			var err error
			lastID, err = p.GormPluginFunctions.GetLastInsertID(tx)
			if err != nil {
				log.WithError(err).Warn("Failed to get last insert ID")
				return nil
			}
			return nil
		})
		if err != nil {
			err = p.errorHandler.HandleError(err, "AddRowReturningID", map[string]any{
				"schema":      schema,
				"storageUnit": storageUnit,
				"valueCount":  len(values),
			})
			return 0, err
		}
		return lastID, nil
	})
}