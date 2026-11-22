package types
import (
	"encoding/json"
	"fmt"
	baml "github.com/boundaryml/baml/engine/language_client_go/pkg"
	"github.com/boundaryml/baml/engine/language_client_go/pkg/cffi"
)
type ChatMessageType string
const (
	ChatMessageTypeSQL     ChatMessageType = "SQL"
	ChatMessageTypeMESSAGE ChatMessageType = "MESSAGE"
	ChatMessageTypeERROR   ChatMessageType = "ERROR"
)
func (ChatMessageType) Values() []ChatMessageType {
	return []ChatMessageType{
		ChatMessageTypeSQL,
		ChatMessageTypeMESSAGE,
		ChatMessageTypeERROR,
	}
}
func (e ChatMessageType) IsValid() bool {
	for _, v := range e.Values() {
		if e == v {
			return true
		}
	}
	return false
}
func (e ChatMessageType) MarshalJSON() ([]byte, error) {
	if !e.IsValid() {
		return nil, fmt.Errorf("invalid ChatMessageType: %q", e)
	}
	return json.Marshal(string(e))
}
func (e *ChatMessageType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*e = ChatMessageType(s)
	if !e.IsValid() {
		return fmt.Errorf("invalid ChatMessageType: %q", s)
	}
	return nil
}
func (e *ChatMessageType) Decode(holder *cffi.CFFIValueEnum, typeMap baml.TypeMap) {
	name := holder.Name
	if name.Name != "ChatMessageType" && name.Namespace != cffi.CFFITypeNamespace_TYPES {
		panic(fmt.Sprintf("expected types.ChatMessageType, got %s.%s", string(name.Namespace.String()), string(name.Name)))
	}
	value := holder.Value
	*e = ChatMessageType(value)
}
func (e ChatMessageType) Encode() (*cffi.HostValue, error) {
	return baml.EncodeEnum("ChatMessageType", string(e), false)
}
func (e ChatMessageType) BamlTypeName() string {
	return "ChatMessageType"
}
type OperationType string
const (
	OperationTypeGET    OperationType = "GET"
	OperationTypeINSERT OperationType = "INSERT"
	OperationTypeUPDATE OperationType = "UPDATE"
	OperationTypeDELETE OperationType = "DELETE"
	OperationTypeCREATE OperationType = "CREATE"
	OperationTypeALTER  OperationType = "ALTER"
	OperationTypeDROP   OperationType = "DROP"
	OperationTypeTEXT   OperationType = "TEXT"
)
func (OperationType) Values() []OperationType {
	return []OperationType{
		OperationTypeGET,
		OperationTypeINSERT,
		OperationTypeUPDATE,
		OperationTypeDELETE,
		OperationTypeCREATE,
		OperationTypeALTER,
		OperationTypeDROP,
		OperationTypeTEXT,
	}
}
func (e OperationType) IsValid() bool {
	for _, v := range e.Values() {
		if e == v {
			return true
		}
	}
	return false
}
func (e OperationType) MarshalJSON() ([]byte, error) {
	if !e.IsValid() {
		return nil, fmt.Errorf("invalid OperationType: %q", e)
	}
	return json.Marshal(string(e))
}
func (e *OperationType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*e = OperationType(s)
	if !e.IsValid() {
		return fmt.Errorf("invalid OperationType: %q", s)
	}
	return nil
}
func (e *OperationType) Decode(holder *cffi.CFFIValueEnum, typeMap baml.TypeMap) {
	name := holder.Name
	if name.Name != "OperationType" && name.Namespace != cffi.CFFITypeNamespace_TYPES {
		panic(fmt.Sprintf("expected types.OperationType, got %s.%s", string(name.Namespace.String()), string(name.Name)))
	}
	value := holder.Value
	*e = OperationType(value)
}
func (e OperationType) Encode() (*cffi.HostValue, error) {
	return baml.EncodeEnum("OperationType", string(e), false)
}
func (e OperationType) BamlTypeName() string {
	return "OperationType"
}