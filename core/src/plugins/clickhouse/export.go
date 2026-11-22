package clickhouse
import (
	"fmt"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins/gorm"
)
func (p *ClickHousePlugin) ExportData(config *engine.PluginConfig, schema string, storageUnit string, writer func([]string) error, selectedRows []map[string]any) error {
	if len(selectedRows) > 0 {
		return p.GormPlugin.ExportData(config, schema, storageUnit, writer, selectedRows)
	}
	db, err := p.DB(config)
	if err != nil {
		return err
	}
	rows, err := db.Table("system.columns").
		Select("name, type").
		Where("database = ? AND table = ?", schema, storageUnit).
		Order("position").
		Rows()
	if err != nil {
		return fmt.Errorf("failed to get columns: %v", err)
	}
	defer rows.Close()
	var columns []string
	var types []string
	for rows.Next() {
		var col, typ string
		if err := rows.Scan(&col, &typ); err != nil {
			return err
		}
		columns = append(columns, col)
		types = append(types, typ)
	}
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = common.FormatCSVHeader(col, types[i])
	}
	if err := writer(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}
	builder := gorm_plugin.NewSQLBuilder(db, p)
	fullTable := builder.BuildFullTableName(schema, storageUnit)
	exportQuery := db.Table(fullTable)
	if len(columns) > 0 {
		exportQuery = exportQuery.Select(columns)
	}
	dataRows, err := exportQuery.Rows()
	if err != nil {
		return fmt.Errorf("failed to query data: %v", err)
	}
	defer dataRows.Close()
	rowCount := 0
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}
	for dataRows.Next() {
		if err := dataRows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}
		row := make([]string, len(columns))
		for i, val := range values {
			row[i] = p.formatValue(val)
		}
		if err := writer(row); err != nil {
			return fmt.Errorf("failed to write row: %v", err)
		}
		rowCount++
	}
	return dataRows.Err()
}
func (p *ClickHousePlugin) formatValue(val any) string {
	if val == nil {
		return ""
	}
	var strVal string
	switch v := val.(type) {
	case []byte:
		strVal = string(v)
	case string:
		strVal = v
	default:
		strVal = fmt.Sprintf("%v", v)
	}
	return common.EscapeFormula(strVal)
}