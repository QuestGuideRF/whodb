package gorm_plugin
import (
	"errors"
	"fmt"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
	"slices"
)
func (p *GormPlugin) UpdateStorageUnit(config *engine.PluginConfig, schema string, storageUnit string, values map[string]string, updatedColumns []string) (bool, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		pkColumns, err := p.GormPluginFunctions.GetPrimaryKeyColumns(db, schema, storageUnit)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to get primary key columns for table %s.%s during update operation", schema, storageUnit))
			pkColumns = []string{}
		}
		columnTypeInfos, err := p.GormPluginFunctions.GetColumnTypes(db, schema, storageUnit)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to get column types for table %s.%s during update operation", schema, storageUnit))
			return false, err
		}
		conditions := make(map[string]any)
		convertedValues := make(map[string]any)
		unchangedValues := make(map[string]any)
		for column, strValue := range values {
			isPK := slices.Contains(pkColumns, column)
			isUpdated := slices.Contains(updatedColumns, column)
			needsConversion := isPK || isUpdated || len(pkColumns) == 0
			if !needsConversion {
				continue
			}
			colInfo, exists := columnTypeInfos[column]
			if !exists {
				return false, fmt.Errorf("column '%s' does not exist in table %s", column, storageUnit)
			}
			convertedValue, err := p.GormPluginFunctions.ConvertStringValue(strValue, colInfo.Type, colInfo.IsNullable)
			if err != nil {
				return false, fmt.Errorf("failed to convert value for column '%s' in table %s.%s: %w", column, schema, storageUnit, err)
			}
			if isPK {
				conditions[column] = convertedValue
			} else if isUpdated {
				convertedValues[column] = convertedValue
			} else {
				unchangedValues[column] = convertedValue
			}
		}
		if len(convertedValues) == 0 {
			return true, nil
		}
		builder := p.GormPluginFunctions.CreateSQLBuilder(db)
		var whereConditions map[string]any
		if len(conditions) == 0 {
			whereConditions = unchangedValues
		} else {
			whereConditions = conditions
		}
		result := builder.UpdateQuery(schema, storageUnit, convertedValues, whereConditions)
		if result.Error != nil {
			log.WithError(result.Error).Error(fmt.Sprintf("Failed to update rows in table %s.%s", schema, storageUnit))
			return false, result.Error
		}
		if p.GormPluginFunctions.ShouldCheckRowsAffected() && result.RowsAffected == 0 {
			return false, errors.New("no rows were updated")
		}
		return true, nil
	})
}