package gorm_plugin
import (
	"database/sql"
	"fmt"
	"time"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
)
func (p *GormPlugin) ExportData(config *engine.PluginConfig, schema string, storageUnit string, writer func([]string) error, selectedRows []map[string]any) error {
	if len(selectedRows) > 0 {
		columns := make([]string, 0, len(selectedRows[0]))
		for col := range selectedRows[0] {
			columns = append(columns, col)
		}
		if err := writer(columns); err != nil {
			log.WithError(err).Error("Failed to write CSV headers for selected rows export")
			return fmt.Errorf("failed to write headers: %v", err)
		}
		for i, row := range selectedRows {
			rowData := make([]string, len(columns))
			for j, col := range columns {
				if val, ok := row[col]; ok {
					rowData[j] = p.GormPluginFunctions.FormatValue(val)
				} else {
					rowData[j] = ""
				}
			}
			if err := writer(rowData); err != nil {
				log.WithError(err).Error(fmt.Sprintf("Failed to write selected row %d during export", i+1))
				return fmt.Errorf("failed to write row %d: %v", i+1, err)
			}
		}
		return nil
	}
	_, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (struct{}, error) {
		columnConfig := config
		if config != nil {
			clonedConfig := *config
			clonedConfig.Transaction = db
			columnConfig = &clonedConfig
		} else {
			columnConfig = &engine.PluginConfig{Transaction: db}
		}
		orderedColumns, err := p.PluginFunctions.GetColumnsForTable(columnConfig, schema, storageUnit)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to get columns for export of table %s.%s", schema, storageUnit))
			return struct{}{}, fmt.Errorf("failed to get columns: %v", err)
		}
		if len(orderedColumns) == 0 {
			return struct{}{}, fmt.Errorf("no columns found for table %s.%s", schema, storageUnit)
		}
		columns := make([]string, len(orderedColumns))
		columnTypes := make([]string, len(orderedColumns))
		for i, col := range orderedColumns {
			columns[i] = col.Name
			columnTypes[i] = col.Type
		}
		headers := make([]string, len(columns))
		for i, col := range columns {
			headers[i] = common.FormatCSVHeader(col, columnTypes[i])
		}
		if err := writer(headers); err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to write CSV headers for table %s.%s export", schema, storageUnit))
			return struct{}{}, fmt.Errorf("failed to write headers: %v", err)
		}
		processor := NewBatchProcessor(p, p.Type, &BatchConfig{
			BatchSize:   10000,
			LogProgress: true,
		})
		totalRows := 0
		err = processor.ExportInBatches(db, schema, storageUnit, columns, func(batch []map[string]any) error {
			for _, record := range batch {
				row := make([]string, len(columns))
				for i, col := range columns {
					if val, exists := record[col]; exists {
						row[i] = p.GormPluginFunctions.FormatValue(val)
					} else {
						row[i] = ""
					}
				}
				if err := writer(row); err != nil {
					log.WithError(err).Error(fmt.Sprintf("Failed to write row %d during export of table %s.%s", totalRows+1, schema, storageUnit))
					return fmt.Errorf("failed to write row %d: %v", totalRows+1, err)
				}
				totalRows++
			}
			return nil
		})
		if err != nil {
			return struct{}{}, fmt.Errorf("export failed after %d rows: %v", totalRows, err)
		}
		log.WithField("totalRows", totalRows).
			WithField("table", fmt.Sprintf("%s.%s", schema, storageUnit)).
			Info("Export completed successfully")
		return struct{}{}, nil
	})
	return err
}
func formatExportValue(plugin GormPluginFunctions, val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case []byte:
		return common.EscapeFormula(string(v))
	case string:
		return common.EscapeFormula(v)
	case time.Time:
		return common.EscapeFormula(plugin.FormatTimeForExport(v))
	case *time.Time:
		if v == nil {
			return ""
		}
		return common.EscapeFormula(plugin.FormatTimeForExport(*v))
	case sql.NullTime:
		if !v.Valid {
			return ""
		}
		return common.EscapeFormula(plugin.FormatTimeForExport(v.Time))
	default:
		return common.EscapeFormula(fmt.Sprintf("%v", v))
	}
}
func (p *GormPlugin) FormatValue(val any) string {
	return formatExportValue(p.GormPluginFunctions, val)
}
func (p *GormPlugin) FormatTimeForExport(value time.Time) string {
	if value.Hour() == 0 && value.Minute() == 0 && value.Second() == 0 && value.Nanosecond() == 0 {
		return value.Format("2006-01-02")
	}
	return value.Format("2006-01-02T15:04:05")
}
func (p *GormPlugin) GetPlaceholder(index int) string {
	return "?"
}