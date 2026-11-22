package postgres
import (
	"database/sql"
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	sourcecatalogspecs "github.com/clidey/whodb/core/src/sourcecatalog/specs"
	"gorm.io/gorm"
)
var (
	supportedOperators = sourcecatalogspecs.PostgreSQLSupportedOperators
)
type PostgresPlugin struct {
	gorm_plugin.GormPlugin
}
func (p *PostgresPlugin) GetSupportedOperators() map[string]string {
	return supportedOperators
}
func (p *PostgresPlugin) GetLastInsertID(db *gorm.DB) (int64, error) {
	var id int64
	if err := db.Raw("SELECT lastval()").Scan(&id).Error; err != nil {
		if strings.Contains(err.Error(), "lastval is not yet defined") {
			return 0, nil
		}
		return 0, err
	}
	return id, nil
}
func (p *PostgresPlugin) GetAllSchemasQuery() string {
	return "SELECT schema_name AS schemaname FROM information_schema.schemata WHERE has_schema_privilege(schema_name, 'USAGE')"
}
func (p *PostgresPlugin) GetTableInfoQuery() string {
	return `
		SELECT
			t.table_name,
			t.table_type,
			pg_size_pretty(pg_total_relation_size(quote_ident(t.table_schema) || '.' || quote_ident(t.table_name))) AS total_size,
			pg_size_pretty(pg_relation_size(quote_ident(t.table_schema) || '.' || quote_ident(t.table_name))) AS data_size
		FROM
			information_schema.tables t
		WHERE
			t.table_schema = ?;
	`
}
func (p *PostgresPlugin) GetStorageUnitExistsQuery() string {
	return `SELECT to_regclass($1 || '.' || $2) IS NOT NULL`
}
func (p *PostgresPlugin) GetPlaceholder(index int) string {
	return fmt.Sprintf("$%d", index)
}
func (p *PostgresPlugin) GetTableNameAndAttributes(rows *sql.Rows) (string, []engine.Record) {
	var tableName, tableType, totalSize, dataSize string
	if err := rows.Scan(&tableName, &tableType, &totalSize, &dataSize); err != nil {
		log.WithError(err).Error("Failed to scan table info row data")
		return "", nil
	}
	attributes := []engine.Record{
		{Key: "Type", Value: tableType},
		{Key: "Total Size", Value: totalSize},
		{Key: "Data Size", Value: dataSize},
	}
	return tableName, attributes
}
func (p *PostgresPlugin) GetDatabases(config *engine.PluginConfig) ([]string, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) ([]string, error) {
		var databases []struct {
			Datname string `gorm:"column:datname"`
		}
		if err := db.Raw("SELECT datname FROM pg_database WHERE datistemplate = false AND datallowconn AND has_database_privilege(datname, 'CONNECT')").
			Scan(&databases).Error; err != nil {
			return nil, err
		}
		databaseNames := []string{}
		for _, database := range databases {
			databaseNames = append(databaseNames, database.Datname)
		}
		return databaseNames, nil
	})
}
func (p *PostgresPlugin) RawExecute(config *engine.PluginConfig, query string, params ...any) (*engine.GetRowsResult, error) {
	return p.ExecuteRawSQL(config, func(cfg *engine.PluginConfig) (*gorm.DB, error) {
		return p.openDB(cfg, true)
	}, query, params...)
}
func (p *PostgresPlugin) GetForeignKeyRelationships(config *engine.PluginConfig, schema string, storageUnit string) (map[string]*engine.ForeignKeyRelationship, error) {
	query := `
		SELECT
			kcu.column_name,
			ccu.table_name AS referenced_table,
			ccu.column_name AS referenced_column
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = ?
			AND tc.table_name = ?
	`
	return p.QueryForeignKeyRelationships(config, query, schema, storageUnit)
}
func (p *PostgresPlugin) NormalizeType(typeName string) string {
	return common.NormalizeTypeWithMap(typeName, sourcecatalogspecs.PostgresAliasMap)
}
func (p *PostgresPlugin) MarkGeneratedColumns(config *engine.PluginConfig, schema string, storageUnit string, columns []engine.Column) error {
	computed, err := p.QueryComputedColumns(config, `
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ? AND is_generated = 'ALWAYS'
	`, schema, storageUnit)
	if err != nil {
		return err
	}
	for i := range columns {
		if computed[columns[i].Name] {
			columns[i].IsComputed = true
		}
	}
	return nil
}
func (p *PostgresPlugin) IsArrayType(columnType string) bool {
	return strings.HasPrefix(columnType, "_")
}
func init() {
	engine.RegisterPlugin(NewPostgresPlugin())
}
func NewPostgresPlugin() *engine.Plugin {
	plugin := &PostgresPlugin{}
	plugin.Type = engine.DatabaseType_Postgres
	plugin.PluginFunctions = plugin
	plugin.GormPluginFunctions = plugin
	return &plugin.Plugin
}