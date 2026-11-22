package gorm_plugin
import (
	"fmt"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
)
func (p *GormPlugin) StreamRawExecute(config *engine.PluginConfig, query string, writer engine.QueryStreamWriter, params ...any) error {
	if config != nil && config.MultiStatement {
		return engine.ErrMultiStatementUnsupported
	}
	_, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		rows, err := db.Raw(query, params...).Rows()
		if err != nil {
			return false, err
		}
		defer rows.Close()
		columns, typeMap, resultColumns, err := p.describeRawColumns(rows)
		if err != nil {
			return false, err
		}
		if err := writer.WriteColumns(resultColumns); err != nil {
			return false, fmt.Errorf("failed to write streamed columns: %w", err)
		}
		for rows.Next() {
			row, err := p.scanRawRow(rows, columns, typeMap)
			if err != nil {
				return false, err
			}
			if err := writer.WriteRow(row); err != nil {
				return false, fmt.Errorf("failed to write streamed row: %w", err)
			}
		}
		if err := rows.Err(); err != nil {
			return false, err
		}
		return true, nil
	})
	return err
}