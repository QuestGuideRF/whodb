package source
import (
	"context"
	"github.com/clidey/whodb/core/src/query"
)
type SourceSession interface {
	Metadata(ctx context.Context) (*SessionMetadata, error)
}
type SourceConnector interface {
	Open(ctx context.Context, spec TypeSpec, credentials *Credentials) (SourceSession, error)
}
type SessionInvalidator interface {
	Invalidate(ctx context.Context, spec TypeSpec, credentials *Credentials) error
}
type DriverShutdowner interface {
	Shutdown(ctx context.Context) error
}
type SourceBrowser interface {
	ListObjects(ctx context.Context, parent *ObjectRef, kinds []ObjectKind) ([]Object, error)
	GetObject(ctx context.Context, ref ObjectRef) (*Object, error)
}
type TabularReader interface {
	ReadRows(ctx context.Context, ref ObjectRef, where *query.WhereCondition, sort []*query.SortCondition, pageSize int, pageOffset int) (*RowsResult, error)
	Columns(ctx context.Context, ref ObjectRef) ([]Column, error)
	ColumnsBatch(ctx context.Context, refs []ObjectRef) ([]ObjectColumns, error)
}
type ColumnConstraintReader interface {
	ColumnConstraints(ctx context.Context, ref ObjectRef) (map[string]map[string]any, error)
}
type ContentReader interface {
	ReadContent(ctx context.Context, ref ObjectRef) (*ContentResult, error)
}
type AvailabilityChecker interface {
	IsAvailable(ctx context.Context) bool
}
type QueryRunner interface {
	RunQuery(ctx context.Context, query string, params ...any) (*RowsResult, error)
}
type QueryStreamWriter interface {
	WriteColumns(columns []Column) error
	WriteRow(row []string) error
}
type StreamQueryRunner interface {
	RunQueryStream(ctx context.Context, query string, writer QueryStreamWriter, params ...any) error
}
type ScriptRunner interface {
	RunScript(ctx context.Context, script string, multiStatement bool, params ...any) (*RowsResult, error)
}
type GraphReader interface {
	ReadGraph(ctx context.Context, ref *ObjectRef) ([]GraphUnit, error)
}
type SourceAssistant interface {
	Reply(ctx context.Context, ref *ObjectRef, previousConversation string, query string) ([]*ChatMessage, error)
}
type ModelAwareSourceAssistant interface {
	ReplyWithModel(ctx context.Context, ref *ObjectRef, previousConversation string, query string, model *ExternalModel) ([]*ChatMessage, error)
}
type ObjectManager interface {
	CreateObject(ctx context.Context, parent *ObjectRef, name string, fields []Record) (bool, error)
	UpdateObject(ctx context.Context, ref ObjectRef, values map[string]string, updatedColumns []string) (bool, error)
	AddRow(ctx context.Context, ref ObjectRef, values []Record) (bool, error)
	DeleteRow(ctx context.Context, ref ObjectRef, values map[string]string) (bool, error)
}
type ConnectionFieldOptionsReader interface {
	ConnectionFieldOptions(ctx context.Context, fieldKey string, values map[string]string) ([]string, error)
}
type TabularExporter interface {
	ExportRows(ctx context.Context, ref ObjectRef, writer func([]string) error, selectedRows []map[string]any) error
}
type NDJSONExporter interface {
	ExportRowsNDJSON(ctx context.Context, ref ObjectRef, writer func(string) error, selectedRows []map[string]any) error
}
type SecurityReader interface {
	SSLStatus(ctx context.Context) (*SSLStatus, error)
}
type DataImporter interface {
	ImportData(ctx context.Context, ref ObjectRef, request ImportRequest) (*ImportResult, error)
}
type MockDataManager interface {
	GenerateMockData(ctx context.Context, ref ObjectRef, rowCount int, fkDensityRatio int, overwriteExisting bool) (*MockDataGenerationResult, error)
	AnalyzeMockDataDependencies(ctx context.Context, ref ObjectRef, rowCount int, fkDensityRatio int) (*MockDataDependencyAnalysis, error)
}
type QuerySuggester interface {
	QuerySuggestions(ctx context.Context, ref *ObjectRef) ([]QuerySuggestion, error)
}