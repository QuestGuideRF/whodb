package gorm_plugin
import (
	ctx "context"
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/bamlconfig"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"github.com/clidey/whodb/core/src/source"
	"gorm.io/gorm"
)
func (p *GormPlugin) buildChatTableContext(config *engine.PluginConfig, db *gorm.DB, schema string) (string, error) {
	rows, err := db.Raw(p.GetTableInfoQuery(), schema).Rows()
	if err != nil {
		log.WithError(err).Error(fmt.Sprintf("Failed to get tables for chat operation in schema: %s", schema))
		return "", err
	}
	defer rows.Close()
	tableDetails := strings.Builder{}
	for rows.Next() {
		tableName, _ := p.GetTableNameAndAttributes(rows)
		if tableName == "" {
			continue
		}
		tableDetails.WriteString(fmt.Sprintf("table: %v\n", tableName))
		orderedColumns, err := p.PluginFunctions.GetColumnsForTable(config, schema, tableName)
		if err != nil {
			log.WithError(err).Warnf("Failed to get column types for table %s in chat", p.FormTableName(schema, tableName))
			continue
		}
		for _, col := range orderedColumns {
			tableDetails.WriteString(fmt.Sprintf("- %v (%v)\n", col.Name, col.Type))
		}
	}
	return tableDetails.String(), nil
}
func (p *GormPlugin) Chat(config *engine.PluginConfig, schema string, previousConversation string, query string) ([]*engine.ChatMessage, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) ([]*engine.ChatMessage, error) {
		tableContext, err := p.buildChatTableContext(config, db, schema)
		if err != nil {
			return nil, err
		}
		callCtx := ctx.Background()
		return bamlconfig.ExecuteChatQuery(
			callCtx,
			string(p.Type),
			schema,
			tableContext,
			previousConversation,
			query,
			config.ExternalModel,
			bamlconfig.ChatQueryExecutorFunc(func(_ ctx.Context, query string, params ...any) (*source.RowsResult, error) {
				return p.RawExecute(config, query, params...)
			}),
		)
	})
}