package gorm_plugin
import (
	"database/sql"
	"strings"
	"sync/atomic"
	"testing"
	"github.com/clidey/whodb/core/src/engine"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)
type chatContextTestPlugin struct {
	GormPlugin
	db          *gorm.DB
	columnsRead atomic.Int32
}
func newChatContextTestPlugin(t *testing.T) *chatContextTestPlugin {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:chat-context-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test database: %v", err)
	}
	if err := db.Exec(`CREATE TABLE orders (id INTEGER PRIMARY KEY, customer_name TEXT NOT NULL)`).Error; err != nil {
		t.Fatalf("failed to create orders table: %v", err)
	}
	plugin := &chatContextTestPlugin{db: db}
	plugin.Type = engine.DatabaseType_QuestDB
	plugin.PluginFunctions = plugin
	plugin.GormPluginFunctions = plugin
	return plugin
}
func (p *chatContextTestPlugin) DB(config *engine.PluginConfig) (*gorm.DB, error) {
	return p.db, nil
}
func (p *chatContextTestPlugin) GetTableInfoQuery() string {
	return `
		SELECT name, type
		FROM sqlite_master
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
		ORDER BY name
	`
}
func (p *chatContextTestPlugin) GetTableNameAndAttributes(rows *sql.Rows) (string, []engine.Record) {
	var tableName string
	var tableType string
	if err := rows.Scan(&tableName, &tableType); err != nil {
		return "", nil
	}
	return tableName, []engine.Record{{Key: "Type", Value: tableType}}
}
func (p *chatContextTestPlugin) GetColumnsForTable(config *engine.PluginConfig, schema string, storageUnit string) ([]engine.Column, error) {
	p.columnsRead.Add(1)
	return []engine.Column{
		{Name: "questdb_only_column", Type: "STRING"},
	}, nil
}
func TestBuildChatTableContextUsesPluginColumnLookup(t *testing.T) {
	plugin := newChatContextTestPlugin(t)
	config := engine.NewPluginConfig(&engine.Credentials{Type: string(engine.DatabaseType_QuestDB)})
	tableContext, err := plugin.buildChatTableContext(config, plugin.db, "")
	if err != nil {
		t.Fatalf("buildChatTableContext returned error: %v", err)
	}
	if plugin.columnsRead.Load() != 1 {
		t.Fatalf("expected chat context to use plugin GetColumnsForTable exactly once, got %d", plugin.columnsRead.Load())
	}
	if !strings.Contains(tableContext, "table: orders\n") {
		t.Fatalf("expected chat context to include orders table, got:\n%s", tableContext)
	}
	if !strings.Contains(tableContext, "- questdb_only_column (STRING)\n") {
		t.Fatalf("expected chat context to use plugin-provided column metadata, got:\n%s", tableContext)
	}
}