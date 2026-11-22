package mockdata
import (
	"fmt"
	"strconv"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
)
func valueToRecord(col engine.Column, value any, constraintType string) engine.Record {
	typeHint := col.Type
	if constraintType != "" {
		typeHint = constraintType
	}
	extra := map[string]string{
		"Type": typeHint,
	}
	var valueStr string
	if value == nil {
		valueStr = ""
		extra["IsNull"] = "true"
	} else if f, ok := value.(float64); ok {
		valueStr = strconv.FormatFloat(f, 'f', -1, 64)
	} else {
		valueStr = fmt.Sprintf("%v", value)
	}
	return engine.Record{
		Key:   col.Name,
		Value: valueStr,
		Extra: extra,
	}
}
func detectDatabaseType(columnType string) string {
	columnType = unwrapTypeModifiers(columnType)
	upperType := strings.ToUpper(columnType)
	if strings.Contains(upperType, "[]") {
		return "array"
	}
	if strings.HasPrefix(upperType, "TIMESTAMP") {
		return "datetime"
	}
	if strings.HasPrefix(upperType, "TIME") && !strings.HasPrefix(upperType, "TINYINT") {
		return "datetime"
	}
	if idx := strings.Index(upperType, "("); idx > 0 {
		upperType = upperType[:idx]
	}
	upperType = strings.TrimSpace(upperType)
	switch {
	case intTypes.Contains(upperType):
		return "int"
	case uintTypes.Contains(upperType):
		return "uint"
	case floatTypes.Contains(upperType):
		return "float"
	case boolTypes.Contains(upperType):
		return "bool"
	case dateTypes.Contains(upperType):
		return "date"
	case dateTimeTypes.Contains(upperType):
		return "datetime"
	case uuidTypes.Contains(upperType):
		return "uuid"
	case jsonTypes.Contains(upperType):
		return "json"
	case textTypes.Contains(upperType):
		return "text"
	default:
		return "text"
	}
}
func getConstraintsForColumn(constraints map[string]map[string]any, columnName string) map[string]any {
	if c, ok := constraints[columnName]; ok {
		return c
	}
	lowerName := strings.ToLower(columnName)
	for key, value := range constraints {
		if strings.ToLower(key) == lowerName {
			return value
		}
	}
	return nil
}