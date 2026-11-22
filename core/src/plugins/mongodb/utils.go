package mongodb
import (
	"strconv"
	"go.mongodb.org/mongo-driver/v2/bson"
)
func normalizeMongoID(value any) (any, error) {
	switch v := value.(type) {
	case bson.ObjectID:
		return v, nil
	case string:
		oid, err := bson.ObjectIDFromHex(v)
		if err != nil {
			return v, nil
		}
		return oid, nil
	default:
		return value, nil
	}
}
func coerceMongoValue(key string, raw string, typeHint string) any {
	if key == "_id" {
		id, err := normalizeMongoID(raw)
		if err == nil {
			return id
		}
	}
	switch typeHint {
	case "double":
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			return f
		}
		return raw
	case "int":
		if i, err := strconv.ParseInt(raw, 10, 32); err == nil {
			return int32(i)
		}
		return raw
	case "long":
		if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return i
		}
		return raw
	case "bool":
		if b, err := strconv.ParseBool(raw); err == nil {
			return b
		}
		return raw
	case "string":
		return raw
	}
	if raw == "true" || raw == "false" {
		b, _ := strconv.ParseBool(raw)
		return b
	}
	if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return f
	}
	return raw
}
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
func getDocumentFieldNames(doc bson.M) []string {
	fields := make([]string, 0, len(doc))
	for k := range doc {
		fields = append(fields, k)
	}
	return fields
}