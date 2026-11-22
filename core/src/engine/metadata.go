package engine
import (
	"strings"
	"github.com/clidey/whodb/core/src/source"
)
func IntPtr(i int) *int {
	return &i
}
func ValidateColumnType(typeName string, sourceType string, metadata *source.TypeSessionMetadata) error {
	if metadata == nil || len(metadata.TypeDefinitions) == 0 {
		return nil
	}
	baseType, hasParams := parseTypeForValidation(typeName)
	baseTypeUpper := strings.ToUpper(baseType)
	if canonical, ok := metadata.AliasMap[baseTypeUpper]; ok {
		baseTypeUpper = strings.ToUpper(canonical)
	}
	var typeDef *TypeDefinition
	for i := range metadata.TypeDefinitions {
		if strings.ToUpper(metadata.TypeDefinitions[i].ID) == baseTypeUpper {
			typeDef = &metadata.TypeDefinitions[i]
			break
		}
	}
	if typeDef == nil {
		return &UnsupportedTypeError{TypeName: typeName, DatabaseType: sourceType}
	}
	if (typeDef.HasLength || typeDef.HasPrecision) && !hasParams {
		return nil
	}
	return nil
}
type UnsupportedTypeError struct {
	TypeName     string
	DatabaseType string
}
func (e *UnsupportedTypeError) Error() string {
	return "data type: " + e.TypeName + " not supported by: " + e.DatabaseType
}
func parseTypeForValidation(typeName string) (baseType string, hasParams bool) {
	typeName = strings.TrimSpace(typeName)
	parenIdx := strings.IndexByte(typeName, '(')
	if parenIdx == -1 {
		return typeName, false
	}
	return typeName[:parenIdx], true
}