package postgres
import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
type QuestDBPlugin struct {
	PostgresPlugin
}
type questDBColumnMetadata struct {
	name       string
	dataType   string
	isNullable bool
}
func (p *QuestDBPlugin) GetTableInfoQuery() string {
	return `
		SELECT
			t.table_name,
			t.table_type
		FROM
			information_schema.tables t
		WHERE
			($1 = '' OR t.table_schema = $1)
			AND t.table_schema NOT IN ('information_schema', 'pg_catalog');
	`
}
func (p *QuestDBPlugin) GetTableNameAndAttributes(rows *sql.Rows) (string, []engine.Record) {
	var tableName, tableType string
	if err := rows.Scan(&tableName, &tableType); err != nil {
		log.WithError(err).Error("Failed to scan QuestDB table info row data")
		return "", nil
	}
	return tableName, []engine.Record{
		{Key: "Type", Value: tableType},
	}
}
func (p *QuestDBPlugin) GetStorageUnitExistsQuery() string {
	return `
		SELECT CASE
			WHEN COUNT(1) > 0 THEN TRUE
			ELSE FALSE
		END
		FROM information_schema.tables
		WHERE ($1 = '' OR table_schema = $1)
			AND table_name = $2
			AND table_schema NOT IN ('information_schema', 'pg_catalog')
	`
}
func (p *QuestDBPlugin) getColumnsQuery() string {
	return `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE ($1 = '' OR table_schema = $1)
			AND table_name = $2
		ORDER BY ordinal_position
	`
}
func (p *QuestDBPlugin) normalizeQuestDBColumnMetadata(columnName string, dataType string, isNullable string) questDBColumnMetadata {
	return questDBColumnMetadata{
		name:       columnName,
		dataType:   p.NormalizeType(dataType),
		isNullable: strings.EqualFold(isNullable, "yes"),
	}
}
func (p *QuestDBPlugin) readQuestDBColumns(db *gorm.DB, schema string, tableName string) ([]questDBColumnMetadata, error) {
	rows, err := db.Raw(p.getColumnsQuery(), schema, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns := make([]questDBColumnMetadata, 0)
	for rows.Next() {
		var columnName string
		var dataType string
		var isNullable string
		if err := rows.Scan(&columnName, &dataType, &isNullable); err != nil {
			return nil, err
		}
		columns = append(columns, p.normalizeQuestDBColumnMetadata(columnName, dataType, isNullable))
	}
	return columns, nil
}
func (p *QuestDBPlugin) GetColumnTypes(db *gorm.DB, schema, tableName string) (map[string]gorm_plugin.ColumnTypeInfo, error) {
	columns, err := p.readQuestDBColumns(db, schema, tableName)
	if err != nil {
		return nil, err
	}
	columnTypes := make(map[string]gorm_plugin.ColumnTypeInfo, len(columns))
	for _, column := range columns {
		columnTypes[column.name] = gorm_plugin.ColumnTypeInfo{
			Type:       column.dataType,
			IsNullable: column.isNullable,
		}
	}
	return columnTypes, nil
}
func (p *QuestDBPlugin) GetColumnsForTable(config *engine.PluginConfig, schema string, storageUnit string) ([]engine.Column, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) ([]engine.Column, error) {
		columnMetadata, err := p.readQuestDBColumns(db, schema, storageUnit)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to get columns for table %s.%s", schema, storageUnit))
			return nil, err
		}
		primaryKeys, err := p.GetPrimaryKeyColumns(db, schema, storageUnit)
		if err != nil {
			log.WithError(err).Warn(fmt.Sprintf("Failed to get primary keys for table %s.%s", schema, storageUnit))
			primaryKeys = []string{}
		}
		columns := make([]engine.Column, 0, len(columnMetadata))
		for _, column := range columnMetadata {
			columns = append(columns, engine.Column{
				Name:       column.name,
				Type:       column.dataType,
				IsNullable: column.isNullable,
				IsPrimary:  slices.Contains(primaryKeys, column.name),
			})
		}
		return columns, nil
	})
}
func (p *QuestDBPlugin) GetPrimaryKeyColQuery() string {
	return `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		JOIN pg_class c ON c.oid = i.indrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE ($1 = '' OR n.nspname = $1) AND c.relname = $2 AND i.indisprimary;
	`
}
func (p *QuestDBPlugin) GetForeignKeyRelationships(_ *engine.PluginConfig, _, _ string) (map[string]*engine.ForeignKeyRelationship, error) {
	return map[string]*engine.ForeignKeyRelationship{}, nil
}
func (p *QuestDBPlugin) MarkGeneratedColumns(config *engine.PluginConfig, schema string, storageUnit string, columns []engine.Column) error {
	return nil
}
func (p *QuestDBPlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	if cached := plugins.GetCachedSSLStatus(config); cached != nil {
		return cached, nil
	}
	sslConfig := ssl.ParseSSLConfig(engine.DatabaseType_QuestDB, config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
	var status *engine.SSLStatus
	if sslConfig == nil || !sslConfig.IsEnabled() {
		status = &engine.SSLStatus{
			IsEnabled: false,
			Mode:      string(ssl.SSLModeDisabled),
		}
	} else {
		status = &engine.SSLStatus{
			IsEnabled: true,
			Mode:      string(sslConfig.Mode),
		}
	}
	plugins.SetCachedSSLStatus(config, status)
	return status, nil
}
func NewQuestDBPlugin() *engine.Plugin {
	questDBPlugin := &QuestDBPlugin{}
	questDBPlugin.Type = engine.DatabaseType_QuestDB
	questDBPlugin.PluginFunctions = questDBPlugin
	questDBPlugin.GormPluginFunctions = questDBPlugin
	return &questDBPlugin.Plugin
}
func init() {
	engine.RegisterPlugin(NewQuestDBPlugin())
}