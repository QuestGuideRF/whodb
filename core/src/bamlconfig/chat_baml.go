package bamlconfig
import (
	"context"
	baml "github.com/boundaryml/baml/engine/language_client_go/pkg"
	"github.com/clidey/whodb/core/baml_client"
	"github.com/clidey/whodb/core/baml_client/types"
	"github.com/clidey/whodb/core/src/source"
)
type ChatQueryExecutor interface {
	RunQuery(ctx context.Context, query string, params ...any) (*source.RowsResult, error)
}
type ChatQueryExecutorFunc func(ctx context.Context, query string, params ...any) (*source.RowsResult, error)
func (fn ChatQueryExecutorFunc) RunQuery(ctx context.Context, query string, params ...any) (*source.RowsResult, error) {
	return fn(ctx, query, params...)
}
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
	dbContext := types.DatabaseContext{
		Database_type:         databaseType,
		Schema:                schema,
		Tables_and_fields:     tableDetails,
		Previous_conversation: previousConversation,
	}
	callOpts := SetupAIClient(model)
	responses, err := baml_client.GenerateSQLQuery(ctx, dbContext, userQuery, callOpts...)
	if err != nil {
		return nil, err
	}
	var chatMessages []*source.ChatMessage
	for _, bamlResp := range responses {
		chatMessages = append(chatMessages, ProcessChatResponse(ctx, &bamlResp, executor))
	}
	return chatMessages, nil
}
func ProcessChatResponse(ctx context.Context, bamlResp *types.ChatResponse, executor ChatQueryExecutor) *source.ChatMessage {
	message := &source.ChatMessage{
		Type: ConvertBAMLTypeToWhoDB(bamlResp.Type),
		Text: bamlResp.Text,
	}
	if bamlResp.Type != types.ChatMessageTypeSQL || bamlResp.Operation == nil {
		return message
	}
	isMutation := *bamlResp.Operation == types.OperationTypeINSERT ||
		*bamlResp.Operation == types.OperationTypeUPDATE ||
		*bamlResp.Operation == types.OperationTypeDELETE ||
		*bamlResp.Operation == types.OperationTypeCREATE ||
		*bamlResp.Operation == types.OperationTypeALTER ||
		*bamlResp.Operation == types.OperationTypeDROP
	if isMutation {
		message.Type = ConvertOperationType(*bamlResp.Operation)
		message.RequiresConfirmation = true
		return message
	}
	result, err := executor.RunQuery(ctx, bamlResp.Text)
	if err != nil {
		message.Type = "error"
		message.Text = err.Error()
	} else {
		message.Type = ConvertOperationType(*bamlResp.Operation)
		message.Result = result
	}
	return message
}
func ConvertBAMLTypeToWhoDB(bamlType types.ChatMessageType) string {
	switch bamlType {
	case types.ChatMessageTypeSQL:
		return "sql"
	case types.ChatMessageTypeMESSAGE:
		return "message"
	case types.ChatMessageTypeERROR:
		return "error"
	default:
		return "message"
	}
}
func OperationToString(op types.OperationType) string {
	switch op {
	case types.OperationTypeGET:
		return "get"
	case types.OperationTypeINSERT:
		return "insert"
	case types.OperationTypeUPDATE:
		return "update"
	case types.OperationTypeDELETE:
		return "delete"
	case types.OperationTypeCREATE:
		return "create"
	case types.OperationTypeALTER:
		return "alter"
	case types.OperationTypeDROP:
		return "drop"
	case types.OperationTypeTEXT:
		return "text"
	default:
		return string(op)
	}
}
func ConvertOperationType(op types.OperationType) string {
	return "sql:" + OperationToString(op)
}
type BAMLConfigResolver func(providerType, apiKey, endpoint, model string) (string, map[string]any, error)
var bamlConfigResolver BAMLConfigResolver
func RegisterBAMLConfigResolver(resolver BAMLConfigResolver) {
	bamlConfigResolver = resolver
}
func SetupAIClient(externalModel *source.ExternalModel) []baml_client.CallOptionFunc {
	var callOpts []baml_client.CallOptionFunc
	if externalModel == nil {
		return callOpts
	}
	if externalModel.Model == "" {
		return callOpts
	}
	if registry := CreateDynamicBAMLClient(externalModel); registry != nil {
		callOpts = append(callOpts, baml_client.WithClientRegistry(registry))
	}
	return callOpts
}
func CreateDynamicBAMLClient(externalModel *source.ExternalModel) *baml.ClientRegistry {
	if externalModel == nil {
		return nil
	}
	registry := baml.NewClientRegistry()
	provider, opts := getBAMLProviderAndOptions(externalModel)
	registry.AddLlmClient("DefaultClient", provider, opts)
	registry.SetPrimaryClient("DefaultClient")
	return registry
}
func getBAMLProviderAndOptions(m *source.ExternalModel) (string, map[string]any) {
	if bamlConfigResolver != nil {
		provider, opts, err := bamlConfigResolver(m.Type, m.Token, m.Endpoint, m.Model)
		if err == nil {
			return provider, opts
		}
	}
	opts := map[string]any{"model": m.Model}
	if m.Endpoint != "" {
		opts["base_url"] = m.Endpoint
	}
	if m.Token != "" {
		opts["api_key"] = m.Token
	}
	return "openai-generic", opts
}