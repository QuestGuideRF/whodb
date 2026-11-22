package baml_client
import (
	"reflect"
	"github.com/clidey/whodb/core/baml_client/stream_types"
	"github.com/clidey/whodb/core/baml_client/types"
)
var typeMap = map[string]reflect.Type{
	"TYPES.ChatResponse":           reflect.TypeOf(types.ChatResponse{}),
	"STREAM_TYPES.ChatResponse":    reflect.TypeOf(stream_types.ChatResponse{}),
	"TYPES.DatabaseContext":        reflect.TypeOf(types.DatabaseContext{}),
	"STREAM_TYPES.DatabaseContext": reflect.TypeOf(stream_types.DatabaseContext{}),
	"TYPES.ChatMessageType": reflect.TypeOf(types.ChatMessageType("")),
	"TYPES.OperationType":   reflect.TypeOf(types.OperationType("")),
}