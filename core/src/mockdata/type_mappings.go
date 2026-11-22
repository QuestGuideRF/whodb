package mockdata
import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
	"github.com/brianvoe/gofakeit/v7"
)
type TypeGenerator func(dbType string, databaseType string, constraints map[string]any, faker *gofakeit.Faker) (any, bool)
var eeTypeGenerator TypeGenerator
func RegisterExtTypeGenerator(gen TypeGenerator) {
	eeTypeGenerator = gen
}
func unwrapTypeModifiers(dbType string) string {
	for {
		lower := strings.ToLower(dbType)
		trimmed := strings.TrimSpace(lower)
		unwrapped := false
		for _, prefix := range []string{"nullable(", "lowcardinality("} {
			if strings.HasPrefix(trimmed, prefix) && strings.HasSuffix(trimmed, ")") {
				inner := strings.TrimSpace(dbType[len(prefix) : len(dbType)-1])
				dbType = inner
				unwrapped = true
				break
			}
		}
		if !unwrapped {
			return dbType
		}
	}
}
func GenerateByType(dbType string, databaseType string, constraints map[string]any, faker *gofakeit.Faker) any {
	if eeTypeGenerator != nil {
		if value, handled := eeTypeGenerator(dbType, databaseType, constraints, faker); handled {
			return value
		}
	}
	dbType = unwrapTypeModifiers(dbType)
	normalizedType := strings.ToLower(dbType)
	if strings.HasSuffix(normalizedType, "[]") {
		return genArray(dbType, faker)
	}
	if strings.HasPrefix(normalizedType, "array(") {
		return genClickHouseArray(dbType, databaseType, constraints, faker)
	}
	if strings.HasPrefix(normalizedType, "map(") {
		return genClickHouseMap(dbType, databaseType, faker)
	}
	if strings.HasPrefix(normalizedType, "tuple(") {
		return genClickHouseTuple(dbType, databaseType, faker)
	}
	if idx := strings.Index(normalizedType, "("); idx > 0 {
		normalizedType = normalizedType[:idx]
	}
	if strings.HasPrefix(normalizedType, "timestamp") {
		return genDateTime(faker)
	}
	if strings.HasPrefix(normalizedType, "time") && !strings.HasPrefix(normalizedType, "tinyint") {
		return genTime(faker)
	}
	switch normalizedType {
	case "int", "integer", "int2", "int4", "int8", "smallint", "bigint", "tinyint", "mediumint", "serial", "bigserial", "smallserial",
		"int16", "int32", "int64", "int128", "int256":
		return genInt(normalizedType, constraints, faker)
	case "uint", "uint8", "uint16", "uint32", "uint64", "uint128", "uint256",
		"tinyint unsigned", "smallint unsigned", "mediumint unsigned", "int unsigned", "bigint unsigned":
		return genUint(normalizedType, constraints, faker)
	case "float", "float4", "float8", "real", "double", "double precision", "decimal", "numeric", "number", "money",
		"float32", "float64":
		return genDecimal(constraints, faker)
	case "bool", "boolean", "bit":
		return faker.Bool()
	case "date", "date32":
		return genDate(faker)
	case "datetime", "datetime64":
		return genDateTime(faker)
	case "interval":
		return genInterval(faker)
	case "year":
		return genYear(faker)
	case "uuid":
		return faker.UUID()
	case "json", "jsonb":
		return genJSON(faker)
	case "bytea", "blob", "binary", "varbinary", "image", "tinyblob", "mediumblob", "longblob":
		return genBinary(constraints, faker)
	case "text", "string", "varchar", "char", "character", "character varying", "nvarchar", "nchar", "ntext", "fixedstring", "clob", "long",
		"tinytext", "mediumtext", "longtext":
		return genText(constraints, faker)
	case "hstore":
		return genHstore(faker)
	case "enum", "enum8", "enum16":
		return genText(constraints, faker)
	case "set":
		return genText(constraints, faker)
	case "inet", "cidr", "ipv4":
		return genIPv4(faker)
	case "ipv6":
		return genIPv6(faker)
	case "macaddr", "macaddr8":
		return genMACAddr(faker)
	case "point":
		return genPoint(faker)
	case "xml":
		return genXML(faker)
	case "line", "lseg", "box", "path", "polygon", "circle":
		return genGeometry(normalizedType, faker)
	case "geometry", "geography", "linestring", "multipoint", "multilinestring", "multipolygon", "geometrycollection":
		return genSpatial(normalizedType, faker)
	case "decimal32", "decimal64", "decimal128", "decimal256":
		return genClickHouseDecimal(normalizedType, constraints, faker)
	default:
		return genText(constraints, faker)
	}
}
func genInt(typeName string, c map[string]any, f *gofakeit.Faker) any {
	var minVal, maxVal int64
	switch typeName {
	case "tinyint":
		minVal, maxVal = 1, 127
	case "smallint", "int2", "int16", "smallserial":
		minVal, maxVal = 1, 32767
	case "mediumint":
		minVal, maxVal = 1, 8388607
	default:
		minVal, maxVal = 1, 1000
	}
	if c != nil {
		if v, ok := c["check_min"].(float64); ok {
			minVal = int64(v)
		}
		if v, ok := c["check_max"].(float64); ok {
			maxVal = int64(v)
		}
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	return f.IntRange(int(minVal), int(maxVal))
}
func genUint(typeName string, c map[string]any, f *gofakeit.Faker) any {
	var minVal, maxVal uint64
	switch typeName {
	case "uint8", "tinyint unsigned":
		minVal, maxVal = 0, 255
	case "uint16", "smallint unsigned":
		minVal, maxVal = 0, 65535
	case "mediumint unsigned":
		minVal, maxVal = 0, 16777215
	default:
		minVal, maxVal = 0, 1000
	}
	if c != nil {
		if v, ok := c["check_min"].(float64); ok && v >= 0 {
			minVal = uint64(v)
		}
		if v, ok := c["check_max"].(float64); ok && v >= 0 {
			maxVal = uint64(v)
		}
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	return f.UintRange(uint(minVal), uint(maxVal))
}
func genDecimal(c map[string]any, f *gofakeit.Faker) any {
	minVal := 0.0
	maxVal := 1000.0
	scale := 2
	if c != nil {
		if s, ok := c["scale"].(int); ok && s >= 0 {
			scale = s
		}
		var precision int
		if p, ok := c["precision"].(int); ok && p > 0 {
			precision = p
		} else if p, ok := c["precision"].(int64); ok && p > 0 {
			precision = int(p)
		}
		if precision > 0 {
			intDigits := precision - scale
			var precisionMax float64
			if intDigits > 0 {
				precisionMax = math.Pow(10, float64(intDigits)) - math.Pow(10, -float64(scale))
			} else {
				precisionMax = 1 - math.Pow(10, -float64(scale))
			}
			if precisionMax < maxVal {
				maxVal = precisionMax
			}
		}
		if v, ok := c["check_min"].(float64); ok {
			minVal = v
		}
		if v, ok := c["check_max"].(float64); ok {
			maxVal = v
		}
	}
	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}
	val := f.Float64Range(minVal, maxVal)
	multiplier := math.Pow(10, float64(scale))
	rounded := math.Round(val*multiplier) / multiplier
	if rounded > maxVal {
		rounded = maxVal
	}
	if rounded < minVal {
		rounded = minVal
	}
	return rounded
}
func genDate(f *gofakeit.Faker) any {
	start := time.Now().AddDate(-10, 0, 0)
	end := time.Now()
	return f.DateRange(start, end).Format("2006-01-02")
}
func genDateTime(f *gofakeit.Faker) any {
	start := time.Now().AddDate(-10, 0, 0)
	end := time.Now()
	return f.DateRange(start, end).Format("2006-01-02 15:04:05")
}
func genTime(f *gofakeit.Faker) any {
	return f.Date().Format("15:04:05")
}
func genYear(f *gofakeit.Faker) any {
	return f.IntRange(1970, time.Now().Year())
}
func genInterval(f *gofakeit.Faker) any {
	units := []string{"seconds", "minutes", "hours", "days", "weeks", "months", "years"}
	unit := units[f.Number(0, len(units)-1)]
	value := f.Number(1, 30)
	return fmt.Sprintf("%d %s", value, unit)
}
func genIPv4(f *gofakeit.Faker) any {
	return f.IPv4Address()
}
func genIPv6(f *gofakeit.Faker) any {
	return f.IPv6Address()
}
func genMACAddr(f *gofakeit.Faker) any {
	return f.MacAddress()
}
func genPoint(f *gofakeit.Faker) any {
	x := f.Float64Range(-180, 180)
	y := f.Float64Range(-90, 90)
	return fmt.Sprintf("(%f,%f)", x, y)
}
func genGeometry(typeName string, f *gofakeit.Faker) any {
	switch typeName {
	case "line":
		return "{1,2,3}"
	case "lseg":
		return "[(0,0),(1,1)]"
	case "box":
		return "(1,1),(0,0)"
	case "path":
		return "[(0,0),(1,1),(2,0)]"
	case "polygon":
		return "((0,0),(1,0),(1,1),(0,1),(0,0))"
	case "circle":
		return "<(0,0),1>"
	default:
		return "(0,0)"
	}
}
func genSpatial(typeName string, f *gofakeit.Faker) any {
	switch typeName {
	case "geometry", "geography":
		return "POINT(0 0)"
	case "linestring":
		return "LINESTRING(0 0, 1 1, 2 2)"
	case "multipoint":
		return "MULTIPOINT((0 0), (1 1))"
	case "multilinestring":
		return "MULTILINESTRING((0 0, 1 1), (2 2, 3 3))"
	case "multipolygon":
		return "MULTIPOLYGON(((0 0, 1 0, 1 1, 0 0)))"
	case "geometrycollection":
		return "GEOMETRYCOLLECTION(POINT(0 0), LINESTRING(0 0, 1 1))"
	default:
		return "POINT(0 0)"
	}
}
func genClickHouseDecimal(typeName string, c map[string]any, f *gofakeit.Faker) any {
	maxPrecision := 9
	switch typeName {
	case "decimal64":
		maxPrecision = 18
	case "decimal128":
		maxPrecision = 38
	case "decimal256":
		maxPrecision = 76
	}
	scale := 2
	if c != nil {
		if s, ok := c["scale"].(int); ok && s >= 0 {
			scale = s
		}
		if p, ok := c["precision"].(int); ok && p > 0 {
			maxPrecision = p
		}
	}
	intDigits := maxPrecision - scale
	if intDigits < 0 {
		intDigits = 0
	}
	if intDigits > 15 {
		intDigits = 15
	}
	maxVal := math.Pow(10, float64(intDigits)) - math.Pow(10, -float64(scale))
	if maxVal <= 0 {
		maxVal = 1 - math.Pow(10, -float64(scale))
	}
	if maxVal > 1000 {
		maxVal = 1000
	}
	val := f.Float64Range(0, maxVal)
	return fmt.Sprintf("%.*f", scale, val)
}
func genXML(f *gofakeit.Faker) any {
	return fmt.Sprintf("<data><id>%d</id><value>%s</value></data>", f.Number(1, 1000), f.Word())
}
func genJSON(f *gofakeit.Faker) any {
	data := map[string]any{
		"id":     f.Number(1, 1000),
		"name":   f.Name(),
		"email":  f.Email(),
		"active": f.Bool(),
	}
	jsonBytes, _ := json.Marshal(data)
	return string(jsonBytes)
}
func genHstore(f *gofakeit.Faker) any {
	pairs := make([]string, f.Number(1, 3))
	for i := range pairs {
		key := f.Word()
		value := f.Word()
		pairs[i] = fmt.Sprintf("\"%s\"=>\"%s\"", key, value)
	}
	return strings.Join(pairs, ",")
}
func genBinary(c map[string]any, f *gofakeit.Faker) any {
	length := 16
	if c != nil {
		if l, ok := c["length"].(int); ok && l > 0 {
			length = min(l, 256)
		}
	}
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = byte(f.Number(0, 255))
	}
	return fmt.Sprintf("0x%x", bytes)
}
func genArray(dbType string, f *gofakeit.Faker) any {
	count := f.Number(1, 5)
	elements := make([]string, count)
	elementType := strings.TrimSuffix(strings.ToLower(dbType), "[]")
	for i := range count {
		if elementType == "int" || elementType == "integer" || elementType == "int4" {
			elements[i] = f.Numerify("###")
		} else {
			elements[i] = f.Word()
		}
	}
	return "{" + strings.Join(elements, ",") + "}"
}
func genClickHouseArray(dbType string, databaseType string, constraints map[string]any, faker *gofakeit.Faker) any {
	innerType := extractInnerType(dbType)
	count := faker.Number(1, 5)
	elements := make([]string, count)
	for i := range count {
		val := GenerateByType(innerType, databaseType, constraints, faker)
		elements[i] = fmt.Sprintf("%v", val)
	}
	return "[" + strings.Join(elements, ", ") + "]"
}
func genClickHouseMap(dbType string, databaseType string, faker *gofakeit.Faker) any {
	keyType, valueType := extractMapTypes(dbType)
	count := faker.Number(1, 3)
	pairs := make([]string, count)
	for i := range count {
		key := GenerateByType(keyType, databaseType, nil, faker)
		val := GenerateByType(valueType, databaseType, nil, faker)
		keyStr := strings.ReplaceAll(fmt.Sprintf("%v", key), "'", "''")
		valStr := strings.ReplaceAll(fmt.Sprintf("%v", val), "'", "''")
		pairs[i] = "'" + keyStr + "': '" + valStr + "'"
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}
func genClickHouseTuple(dbType string, databaseType string, faker *gofakeit.Faker) any {
	innerTypes := extractTupleTypes(dbType)
	elements := make([]string, len(innerTypes))
	for i, t := range innerTypes {
		val := GenerateByType(t, databaseType, nil, faker)
		if s, isStr := val.(string); isStr {
			elements[i] = "'" + strings.ReplaceAll(s, "'", "''") + "'"
		} else {
			elements[i] = fmt.Sprintf("%v", val)
		}
	}
	return "(" + strings.Join(elements, ", ") + ")"
}
func extractInnerType(dbType string) string {
	start := strings.Index(dbType, "(")
	if start == -1 {
		return "String"
	}
	end := strings.LastIndex(dbType, ")")
	if end == -1 || end <= start {
		return "String"
	}
	return strings.TrimSpace(dbType[start+1 : end])
}
func extractMapTypes(dbType string) (string, string) {
	inner := extractInnerType(dbType)
	depth := 0
	for i, ch := range inner {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				return strings.TrimSpace(inner[:i]), strings.TrimSpace(inner[i+1:])
			}
		}
	}
	return "String", "String"
}
func extractTupleTypes(dbType string) []string {
	inner := extractInnerType(dbType)
	var types []string
	depth := 0
	start := 0
	for i, ch := range inner {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				types = append(types, strings.TrimSpace(inner[start:i]))
				start = i + 1
			}
		}
	}
	if remaining := strings.TrimSpace(inner[start:]); remaining != "" {
		types = append(types, remaining)
	}
	if len(types) == 0 {
		return []string{"String"}
	}
	return types
}
func genText(c map[string]any, f *gofakeit.Faker) any {
	if c != nil {
		if values, ok := c["check_values"].([]string); ok && len(values) > 0 {
			return values[f.Number(0, len(values)-1)]
		}
		if values, ok := c["check_values"].([]any); ok && len(values) > 0 {
			strValues := make([]string, len(values))
			for i, v := range values {
				strValues[i], _ = v.(string)
			}
			return strValues[f.Number(0, len(strValues)-1)]
		}
	}
	maxLen := 255
	if c != nil {
		if l, ok := c["length"].(int); ok && l > 0 {
			maxLen = l
		}
	}
	var text string
	if maxLen <= 10 {
		text = f.LetterN(uint(maxLen))
	} else if maxLen <= 50 {
		text = f.LoremIpsumSentence(3)
	} else {
		text = f.LoremIpsumSentence(f.Number(3, 10))
	}
	if len(text) > maxLen {
		text = text[:maxLen]
	}
	return text
}