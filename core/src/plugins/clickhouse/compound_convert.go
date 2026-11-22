package clickhouse
import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"gorm.io/gorm"
)
func formatLiteral(v any) string {
	switch v := v.(type) {
	case string:
		return "'" + v + "'"
	default:
		return fmt.Sprintf("%v", v)
	}
}
func formatMap(m map[string]any) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(m))
	for _, k := range keys {
		parts = append(parts, "'"+k+"': "+formatLiteral(m[k]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
func formatTuple(items []any) string {
	parts := make([]string, len(items))
	for i, v := range items {
		parts[i] = formatLiteral(v)
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
func formatSlice(items []any) string {
	parts := make([]string, len(items))
	for i, v := range items {
		parts[i] = formatLiteral(v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
func formatReflectMap(rv reflect.Value, upperType string) string {
	keys := rv.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
	})
	parts := make([]string, 0, rv.Len())
	for _, key := range keys {
		k := formatLiteral(key.Interface())
		v := formatLiteral(rv.MapIndex(key).Interface())
		parts = append(parts, k+": "+v)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
func formatReflectTuple(rv reflect.Value) string {
	parts := make([]string, rv.Len())
	for i := range rv.Len() {
		parts[i] = formatLiteral(rv.Index(i).Interface())
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
func formatReflectSlice(rv reflect.Value, upperType string) string {
	parts := make([]string, rv.Len())
	for i := range rv.Len() {
		parts[i] = formatLiteral(rv.Index(i).Interface())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
func convertArrayLiteral(value string, columnType string) (any, error) {
	elemType := extractInner(columnType)
	if elemType == "" {
		elemType = "String"
	}
	goType := chTypeToReflect(elemType)
	sliceType := reflect.SliceOf(goType)
	result := reflect.MakeSlice(sliceType, 0, 0)
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "["), "]"))
	if inner == "" {
		return result.Interface(), nil
	}
	for _, elem := range splitTopLevel(inner, ',') {
		elem = strings.Trim(strings.TrimSpace(elem), "'")
		v, err := parseValue(elem, elemType)
		if err != nil {
			return nil, fmt.Errorf("array element: %w", err)
		}
		result = reflect.Append(result, reflect.ValueOf(v))
	}
	return result.Interface(), nil
}
func arrayToExpr(v any) any {
	rv := reflect.ValueOf(v)
	parts := make([]string, rv.Len())
	for i := range rv.Len() {
		parts[i] = formatLiteral(rv.Index(i).Interface())
	}
	return gorm.Expr("[" + strings.Join(parts, ", ") + "]")
}
func convertMapLiteral(value string, columnType string) (any, error) {
	keyType, valType := extractTypeParams2(columnType)
	keyGoType := chTypeToReflect(keyType)
	valGoType := chTypeToReflect(valType)
	mapType := reflect.MapOf(keyGoType, valGoType)
	result := reflect.MakeMap(mapType)
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "{"), "}"))
	if inner == "" {
		return result.Interface(), nil
	}
	for _, entry := range splitTopLevel(inner, ',') {
		parts := splitTopLevel(strings.TrimSpace(entry), ':')
		if len(parts) != 2 {
			continue
		}
		k, err := parseValue(strings.Trim(strings.TrimSpace(parts[0]), "'"), keyType)
		if err != nil {
			return nil, fmt.Errorf("map key: %w", err)
		}
		v, err := parseValue(strings.Trim(strings.TrimSpace(parts[1]), "'"), valType)
		if err != nil {
			return nil, fmt.Errorf("map value: %w", err)
		}
		result.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
	return result.Interface(), nil
}
func convertTupleLiteral(value string, columnType string) (any, error) {
	elemTypes := splitTopLevel(extractInner(columnType), ',')
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "("), ")"))
	if inner == "" {
		return []any{}, nil
	}
	elements := splitTopLevel(inner, ',')
	result := make([]any, 0, len(elements))
	for i, elem := range elements {
		chType := "String"
		if i < len(elemTypes) {
			chType = strings.TrimSpace(elemTypes[i])
		}
		v, err := parseValue(strings.Trim(strings.TrimSpace(elem), "'"), chType)
		if err != nil {
			return nil, fmt.Errorf("tuple element %d: %w", i, err)
		}
		result = append(result, v)
	}
	return result, nil
}
func parseValue(s string, chType string) (any, error) {
	switch strings.ToUpper(strings.TrimSpace(chType)) {
	case "INT8":
		v, err := strconv.ParseInt(s, 10, 8)
		return int8(v), err
	case "INT16":
		v, err := strconv.ParseInt(s, 10, 16)
		return int16(v), err
	case "INT32":
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	case "INT64":
		v, err := strconv.ParseInt(s, 10, 64)
		return v, err
	case "UINT8":
		v, err := strconv.ParseUint(s, 10, 8)
		return uint8(v), err
	case "UINT16":
		v, err := strconv.ParseUint(s, 10, 16)
		return uint16(v), err
	case "UINT32":
		v, err := strconv.ParseUint(s, 10, 32)
		return uint32(v), err
	case "UINT64":
		v, err := strconv.ParseUint(s, 10, 64)
		return v, err
	case "FLOAT32":
		v, err := strconv.ParseFloat(s, 32)
		return float32(v), err
	case "FLOAT64":
		v, err := strconv.ParseFloat(s, 64)
		return v, err
	case "BOOL", "BOOLEAN":
		v, err := strconv.ParseBool(s)
		return v, err
	default:
		return s, nil
	}
}
func chTypeToReflect(chType string) reflect.Type {
	switch strings.ToUpper(strings.TrimSpace(chType)) {
	case "INT8":
		return reflect.TypeOf(int8(0))
	case "INT16":
		return reflect.TypeOf(int16(0))
	case "INT32":
		return reflect.TypeOf(int32(0))
	case "INT64":
		return reflect.TypeOf(int64(0))
	case "UINT8":
		return reflect.TypeOf(uint8(0))
	case "UINT16":
		return reflect.TypeOf(uint16(0))
	case "UINT32":
		return reflect.TypeOf(uint32(0))
	case "UINT64":
		return reflect.TypeOf(uint64(0))
	case "FLOAT32":
		return reflect.TypeOf(float32(0))
	case "FLOAT64":
		return reflect.TypeOf(float64(0))
	case "BOOL", "BOOLEAN":
		return reflect.TypeOf(false)
	default:
		return reflect.TypeOf("")
	}
}
func splitTopLevel(s string, sep byte) []string {
	var result []string
	depth := 0
	inQ := false
	start := 0
	for i := 0; i < len(s); i++ {
		switch ch := s[i]; {
		case ch == '\'':
			inQ = !inQ
		case !inQ:
			switch ch {
			case '(', '[', '{':
				depth++
			case ')', ']', '}':
				depth--
			default:
				if ch == sep && depth == 0 {
					result = append(result, s[start:i])
					start = i + 1
				}
			}
		}
	}
	if start <= len(s) {
		result = append(result, s[start:])
	}
	return result
}
func extractInner(s string) string {
	i := strings.Index(s, "(")
	j := strings.LastIndex(s, ")")
	if i == -1 || j <= i {
		return ""
	}
	return s[i+1 : j]
}
func extractTypeParams2(columnType string) (string, string) {
	parts := splitTopLevel(extractInner(columnType), ',')
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "String", "String"
}