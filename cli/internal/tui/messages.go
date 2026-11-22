package tui
import (
	"time"
	"github.com/clidey/whodb/cli/internal/database"
	"github.com/clidey/whodb/cli/internal/schemadiff"
	"github.com/clidey/whodb/core/src/engine"
)
type OperationState int
const (
	OperationIdle OperationState = iota
	OperationRunning
	OperationCancelling
)
type QueryExecutedMsg struct {
	Result *engine.GetRowsResult
	Query  string
	Err    error
}
type QueryCancelledMsg struct {
	Query string
}
type QueryTimeoutMsg struct {
	Query   string
	Timeout time.Duration
}
type HistoryQueryMsg struct {
	Result *engine.GetRowsResult
	Query  string
	Err    error
}
type PageLoadedMsg struct {
	Results   *engine.GetRowsResult
	Err       error
	Schema    string
	TableName string
}
type AutocompleteDebounceMsg struct {
	SeqID int
	Text  string
	Pos   int
}
type tablesLoadedMsg struct {
	tables  []engine.StorageUnit
	schemas []string
	schema  string
	err     error
}
type chatResponseMsg struct {
	messages []*database.ChatMessage
	query    string
	err      error
}
type chatStreamChunkMsg struct {
	text string
}
type chatStreamDoneMsg struct {
	messages []*database.ChatMessage
	query    string
	err      error
}
type modelsLoadedMsg struct {
	models []string
	err    error
}
type connectionResultMsg struct {
	err           error
	statusMessage string
}
type dockerConnectionsLoadedMsg struct {
	items []connectionItem
}
type cloudConnectionsLoadedMsg struct {
	items []connectionItem
}
type escTimeoutTickMsg struct{}
type exportResultMsg struct {
	success       bool
	err           error
	savedFilePath string
}
type schemaLoadedMsg struct {
	tables []tableWithColumns
	err    error
	schema string
}
type schemaTableColumnsLoadedMsg struct {
	tableName string
	columns   []engine.Column
	err       error
}
type statusMessageTimeoutMsg struct{}
type themeChangedMsg struct {
	themeName string
}
type escConfirmTimeoutMsg struct{}
type explainResultMsg struct {
	query string
	plan  string
	err   error
}
type schemaDiffResultMsg struct {
	result *schemadiff.Result
	err    error
}
type tableWithColumns struct {
	StorageUnit    engine.StorageUnit
	Columns        []engine.Column
	ColumnsLoaded  bool
	ColumnsLoading bool
}