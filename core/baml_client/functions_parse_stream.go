package baml_client
import (
	"context"
	"fmt"
	baml "github.com/boundaryml/baml/engine/language_client_go/pkg"
	"github.com/clidey/whodb/core/baml_client/stream_types"
)
type parse_stream struct{}
var ParseStream = &parse_stream{}
func (*parse_stream) GenerateChatTitle(text string, opts ...CallOptionFunc) (string, error) {
	var callOpts callOption
	for _, opt := range opts {
		opt(&callOpts)
	}
	args := baml.BamlFunctionArguments{
		Kwargs: map[string]any{"text": text, "stream": true},
		Env:    getEnvVars(callOpts.env),
	}
	if callOpts.clientRegistry != nil {
		args.ClientRegistry = callOpts.clientRegistry
	}
	if callOpts.collectors != nil {
		args.Collectors = callOpts.collectors
	}
	if callOpts.typeBuilder != nil {
		args.TypeBuilder = callOpts.typeBuilder
	}
	if callOpts.tags != nil {
		args.Tags = callOpts.tags
	}
	encoded, err := args.Encode()
	if err != nil {
		wrapped_err := fmt.Errorf("BAML INTERNAL ERROR: GenerateChatTitle: %w", err)
		panic(wrapped_err)
	}
	result, err := bamlRuntime.CallFunctionParse(context.Background(), "GenerateChatTitle", encoded)
	if err != nil {
		return "", err
	}
	casted := (result).(string)
	return casted, nil
}
func (*parse_stream) GenerateSQLQuery(text string, opts ...CallOptionFunc) ([]stream_types.ChatResponse, error) {
	var callOpts callOption
	for _, opt := range opts {
		opt(&callOpts)
	}
	args := baml.BamlFunctionArguments{
		Kwargs: map[string]any{"text": text, "stream": true},
		Env:    getEnvVars(callOpts.env),
	}
	if callOpts.clientRegistry != nil {
		args.ClientRegistry = callOpts.clientRegistry
	}
	if callOpts.collectors != nil {
		args.Collectors = callOpts.collectors
	}
	if callOpts.typeBuilder != nil {
		args.TypeBuilder = callOpts.typeBuilder
	}
	if callOpts.tags != nil {
		args.Tags = callOpts.tags
	}
	encoded, err := args.Encode()
	if err != nil {
		wrapped_err := fmt.Errorf("BAML INTERNAL ERROR: GenerateSQLQuery: %w", err)
		panic(wrapped_err)
	}
	result, err := bamlRuntime.CallFunctionParse(context.Background(), "GenerateSQLQuery", encoded)
	if err != nil {
		return nil, err
	}
	casted := (result).([]stream_types.ChatResponse)
	return casted, nil
}