package bamlconfig
import (
	"context"
	"errors"
	"github.com/clidey/whodb/core/src/source"
)
type ChatQueryExecutor interface {
	RunQuery(ctx context.Context, query string, params ...any) (*source.RowsResult, error)
}
type ChatQueryExecutorFunc func(ctx context.Context, query string, params ...any) (*source.RowsResult, error)
func (fn ChatQueryExecutorFunc) RunQuery(ctx context.Context, query string, params ...any) (*source.RowsResult, error) {
	return fn(ctx, query, params...)
}
var ErrBAMLNotSupported = errors.New("AI features are not supported on this platform (arm/riscv64). BAML requires amd64 or arm64 architecture")
func ExecuteChatQuery(
	ctx context.Context,
	databaseType string,
	schema string,
	tableDetails string,
	previousConversation string,
	userQuery string,
	model *source.ExternalModel,
	executor ChatQueryExecutor,
) ([]*source.ChatMessage, error) {
	return nil, ErrBAMLNotSupported
}
func SetupAIClient(externalModel *source.ExternalModel) []any {
	return nil
}
func CreateDynamicBAMLClient(externalModel *source.ExternalModel) any {
	return nil
}
type BAMLConfigResolver func(providerType, apiKey, endpoint, model string) (string, map[string]any, error)
func RegisterBAMLConfigResolver(resolver BAMLConfigResolver) {}