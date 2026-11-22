package database
import (
	"context"
	"fmt"
	"strings"
	"github.com/clidey/whodb/cli/internal/baml"
	"github.com/clidey/whodb/core/baml_client"
	"github.com/clidey/whodb/core/baml_client/types"
	"github.com/clidey/whodb/core/src/bamlconfig"
	"github.com/clidey/whodb/core/src/source"
)
func (m *Manager) SendAIChatStream(ctx context.Context, providerID, modelType, token, schema, model, previousConversation, query string) (<-chan StreamChunk, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	baml.Ensure()
	spec, session, err := m.currentSourceSession(ctx)
	if err != nil {
		return nil, err
	}
	reader, ok := session.(source.TabularReader)
	if !ok {
		return nil, fmt.Errorf("chat is not supported for %s", spec.Label)
	}
	runner, ok := session.(source.QueryRunner)
	if !ok {
		return nil, fmt.Errorf("querying is not supported for %s", spec.Label)
	}
	externalModel := resolveExternalModel(providerID, modelType, token, model)
	tableDetails, err := m.buildSourceChatTableDetails(ctx, spec, session, reader, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}
	dbContext := types.DatabaseContext{
		Database_type:         spec.ID,
		Schema:                schema,
		Tables_and_fields:     tableDetails,
		Previous_conversation: previousConversation,
	}
	callOpts := bamlconfig.SetupAIClient(externalModel)
	bamlStream, err := baml_client.Stream.GenerateSQLQuery(ctx, dbContext, query, callOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to start stream: %w", err)
	}
	out := make(chan StreamChunk, 1)
	go func() {
		defer close(out)
		var lastText string
		for chunk := range bamlStream {
			if chunk.IsError {
				out <- StreamChunk{Err: chunk.Error}
				return
			}
			if chunk.IsFinal {
				final := chunk.Final()
				if final == nil {
					out <- StreamChunk{IsFinal: true, Final: []*ChatMessage{}}
					return
				}
				messages := convertSourceFinalResponses(ctx, *final, runner)
				out <- StreamChunk{IsFinal: true, Final: messages}
				return
			}
			if stream := chunk.Stream(); stream != nil {
				for _, resp := range *stream {
					if resp.Text != nil {
						lastText = *resp.Text
					}
				}
				if lastText != "" {
					out <- StreamChunk{Text: lastText}
				}
			}
		}
		if lastText != "" {
			out <- StreamChunk{
				IsFinal: true,
				Final: []*ChatMessage{{
					Type: "message",
					Text: lastText,
				}},
			}
		} else {
			out <- StreamChunk{IsFinal: true, Final: []*ChatMessage{}}
		}
	}()
	return out, nil
}
func (m *Manager) buildSourceChatTableDetails(ctx context.Context, spec source.TypeSpec, session source.SourceSession, reader source.TabularReader, schema string) (string, error) {
	objects, err := m.listStorageUnitObjects(ctx, spec, session, schema)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, object := range objects {
		b.WriteString(fmt.Sprintf("table: %s\n", object.Name))
		columns, err := reader.Columns(ctx, object.Ref)
		if err != nil {
			continue
		}
		for _, col := range columns {
			b.WriteString(fmt.Sprintf("- %s (%s)\n", col.Name, col.Type))
		}
	}
	return b.String(), nil
}
type sourceChatQueryExecutor struct {
	ctx    context.Context
	runner source.QueryRunner
}
func (e *sourceChatQueryExecutor) RunQuery(_ context.Context, query string, params ...any) (*source.RowsResult, error) {
	return e.runner.RunQuery(e.ctx, query, params...)
}
func convertSourceFinalResponses(ctx context.Context, responses []types.ChatResponse, runner source.QueryRunner) []*ChatMessage {
	var messages []*ChatMessage
	executor := &sourceChatQueryExecutor{ctx: ctx, runner: runner}
	for _, resp := range responses {
		resp.Text = strings.TrimRight(strings.TrimSpace(resp.Text), ";")
		chatMsg := bamlconfig.ProcessChatResponse(ctx, &resp, executor)
		messages = append(messages, &ChatMessage{
			Type:                 chatMsg.Type,
			Text:                 chatMsg.Text,
			RequiresConfirmation: chatMsg.RequiresConfirmation,
			Result:               chatMsg.Result,
		})
	}
	return messages
}