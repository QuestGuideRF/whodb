package engine
import (
	"context"
	"errors"
	"time"
	"github.com/clidey/whodb/core/src/query"
)
var ErrMultiStatementUnsupported = errors.New("multi-statement SQL is not supported by this database")
type Credentials struct {
	Id          *string  `json:"Id,omitempty"`
	Type        string   `json:"Type"`
	Hostname    string   `json:"Hostname"`
	Username    string   `json:"Username"`
	Password    string   `json:"Password"`
	Database    string   `json:"Database"`
	Advanced    []Record `json:"Advanced,omitempty"`
	AccessToken *string  `json:"AccessToken,omitempty"`
	IsProfile   bool     `json:"IsProfile,omitempty"`
}
type PluginConfig struct {
	Credentials           *Credentials
	ExternalModel         *ExternalModel
	Context               context.Context
	Transaction           any
	MultiStatement        bool
	UpsertPKColumns       []string
	SkipConflictPKColumns []string
}
type GetRowsRequest struct {
	Schema      string
	StorageUnit string
	Where       *query.WhereCondition
	Sort        []*query.SortCondition
	PageSize    int
	PageOffset  int
}
type PluginFunctions interface {
	GetDatabases(config *PluginConfig) ([]string, error)
	IsAvailable(ctx context.Context, config *PluginConfig) bool
	GetAllSchemas(config *PluginConfig) ([]string, error)
	GetStorageUnits(config *PluginConfig, schema string) ([]StorageUnit, error)
	StorageUnitExists(config *PluginConfig, schema string, storageUnit string) (bool, error)
	AddStorageUnit(config *PluginConfig, schema string, storageUnit string, fields []Record) (bool, error)
	UpdateStorageUnit(config *PluginConfig, schema string, storageUnit string, values map[string]string, updatedColumns []string) (bool, error)
	AddRow(config *PluginConfig, schema string, storageUnit string, values []Record) (bool, error)
	AddRowReturningID(config *PluginConfig, schema string, storageUnit string, values []Record) (int64, error)
	BulkAddRows(config *PluginConfig, schema string, storageUnit string, rows [][]Record) (bool, error)
	DeleteRow(config *PluginConfig, schema string, storageUnit string, values map[string]string) (bool, error)
	GetRows(config *PluginConfig, req *GetRowsRequest) (*GetRowsResult, error)
	GetRowCount(config *PluginConfig, schema string, storageUnit string, where *query.WhereCondition) (int64, error)
	GetGraph(config *PluginConfig, schema string) ([]GraphUnit, error)
	RawExecute(config *PluginConfig, query string, params ...any) (*GetRowsResult, error)
	Chat(config *PluginConfig, schema string, previousConversation string, query string) ([]*ChatMessage, error)
	ExportData(config *PluginConfig, schema string, storageUnit string, writer func([]string) error, selectedRows []map[string]any) error
	FormatValue(val any) string
	GetColumnsForTable(config *PluginConfig, schema string, storageUnit string) ([]Column, error)
	MarkGeneratedColumns(config *PluginConfig, schema string, storageUnit string, columns []Column) error
	GetColumnConstraints(config *PluginConfig, schema string, storageUnit string) (map[string]map[string]any, error)
	ClearTableData(config *PluginConfig, schema string, storageUnit string) (bool, error)
	NullifyFKColumn(config *PluginConfig, schema string, storageUnit string, column string) error
	GetForeignKeyRelationships(config *PluginConfig, schema string, storageUnit string) (map[string]*ForeignKeyRelationship, error)
	WithTransaction(config *PluginConfig, operation func(tx any) error) error
	GetSSLStatus(config *PluginConfig) (*SSLStatus, error)
}
type Plugin struct {
	PluginFunctions
	Type DatabaseType
}
func NewPluginConfig(credentials *Credentials) *PluginConfig {
	return &PluginConfig{
		Credentials: credentials,
	}
}
func (c *PluginConfig) OperationContext() context.Context {
	if c != nil && c.Context != nil {
		return c.Context
	}
	return context.Background()
}
func (c *PluginConfig) OperationContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.OperationContext(), timeout)
}