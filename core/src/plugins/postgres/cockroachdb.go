package postgres
import (
	"database/sql"
	"github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	sourcecatalogspecs "github.com/clidey/whodb/core/src/sourcecatalog/specs"
	"gorm.io/gorm"
)
type CockroachDBPlugin struct {
	PostgresPlugin
}
func (p *CockroachDBPlugin) GetTableInfoQuery() string {
	return `
		SELECT
			t.table_name,
			t.table_type
		FROM
			information_schema.tables t
		WHERE
			t.table_schema = ?;
	`
}
func (p *CockroachDBPlugin) GetTableNameAndAttributes(rows *sql.Rows) (string, []engine.Record) {
	var tableName, tableType string
	if err := rows.Scan(&tableName, &tableType); err != nil {
		log.WithError(err).Error("Failed to scan CockroachDB table info row data")
		return "", nil
	}
	attributes := []engine.Record{
		{Key: "Type", Value: tableType},
	}
	return tableName, attributes
}
func (p *CockroachDBPlugin) GetStorageUnitExistsQuery() string {
	return `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2)`
}
func (p *CockroachDBPlugin) IsGeometryType(columnType string) bool {
	return false
}
func (p *CockroachDBPlugin) GetSSLStatus(config *engine.PluginConfig) (*engine.SSLStatus, error) {
	if cached := plugins.GetCachedSSLStatus(config); cached != nil {
		return cached, nil
	}
	status, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (*engine.SSLStatus, error) {
		var result struct {
			SSL string `gorm:"column:ssl"`
		}
		query := db.Raw("SHOW ssl").Scan(&result)
		if query.Error != nil {
			return nil, query.Error
		}
		if result.SSL != "on" {
			return &engine.SSLStatus{
				IsEnabled: false,
				Mode:      string(ssl.SSLModeDisabled),
			}, nil
		}
		sslConfig := ssl.ParseSSLConfig(engine.DatabaseType(p.Type), config.Credentials.Advanced, config.Credentials.Hostname, config.Credentials.IsProfile)
		mode := "enabled"
		if sslConfig != nil {
			mode = string(sslConfig.Mode)
		}
		return &engine.SSLStatus{
			IsEnabled: true,
			Mode:      mode,
		}, nil
	})
	if err == nil && status != nil {
		plugins.SetCachedSSLStatus(config, status)
	}
	return status, err
}
var cockroachDBTypeDefinitions = sourcecatalogspecs.CockroachDBTypeDefinitions
func NewCockroachDBPlugin() *engine.Plugin {
	crdbPlugin := &CockroachDBPlugin{}
	crdbPlugin.Type = engine.DatabaseType_CockroachDB
	crdbPlugin.PluginFunctions = crdbPlugin
	crdbPlugin.GormPluginFunctions = crdbPlugin
	return &crdbPlugin.Plugin
}
func init() {
	engine.RegisterPlugin(NewCockroachDBPlugin())
}