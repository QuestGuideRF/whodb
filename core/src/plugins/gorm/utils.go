package gorm_plugin
import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dromara/carbon/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)
var (
	intTypes      = common.IntTypes
	uintTypes     = common.UintTypes
	bigIntTypes   = common.BigIntTypes
	decimalTypes  = common.DecimalTypes
	floatTypes    = common.FloatTypes
	boolTypes     = common.BoolTypes
	dateTypes     = common.DateTypes
	dateTimeTypes = common.DateTimeTypes
	uuidTypes     = common.UuidTypes
	binaryTypes   = common.BinaryTypes
	jsonTypes     = common.JsonTypes
	networkTypes  = common.NetworkTypes
	geometryTypes = common.GeometryTypes
	xmlTypes      = common.XMLTypes
)
func (p *GormPlugin) ConvertRecordValuesToMap(values []engine.Record) (map[string]any, error) {
	data := make(map[string]any, len(values))
	for _, value := range values {
		if value.Extra != nil && value.Extra["IsNull"] == "true" {
			data[value.Key] = nil
		} else {
			isNullable := value.Extra != nil && value.Extra["IsNullable"] == "true"
			val, err := p.GormPluginFunctions.ConvertStringValue(value.Value, value.Extra["Type"], isNullable)
			if err != nil {
				return nil, err
			}
			data[value.Key] = val
		}
	}
	return data, nil
}
func (p *GormPlugin) GetPrimaryKeyColumns(db *gorm.DB, schema string, tableName string) ([]string, error) {
	var primaryKeys []string
	query := p.GetPrimaryKeyColQuery()
	if query == "" {
		return primaryKeys, nil
	}
	rows, err := db.Raw(query, schema, tableName).Rows()
	if err != nil {
		log.Debug(fmt.Sprintf("No primary keys found for table %s.%s: %v", schema, tableName, err))
		return primaryKeys, nil
	}
	defer rows.Close()
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to scan primary key column name for table %s.%s", schema, tableName))
			continue
		}
		primaryKeys = append(primaryKeys, columnName)
	}
	return primaryKeys, nil
}
func (p *GormPlugin) GetColumnTypes(db *gorm.DB, schema, tableName string) (map[string]ColumnTypeInfo, error) {
	migrator := NewMigratorHelper(db, p.GormPluginFunctions)
	fullTableName := p.FormTableName(schema, tableName)
	return migrator.GetColumnTypes(fullTableName)
}
type typeConverter struct {
	types mapset.Set[string]
	isNumeric bool
	nullValue func() any
	convert func(p *GormPlugin, value, baseType, columnType string, isNullable bool) (any, error)
}
var typeConverters = []typeConverter{
	{
		types: intTypes, isNumeric: true,
		nullValue: func() any { return sql.NullInt64{Valid: false} },
		convert:   convertInt,
	},
	{
		types: uintTypes, isNumeric: true,
		nullValue: func() any { return nil },
		convert:   convertUint,
	},
	{
		types: bigIntTypes, isNumeric: true,
		nullValue: func() any { return nil },
		convert:   convertBigInt,
	},
	{
		types: decimalTypes, isNumeric: true,
		nullValue: func() any { return nil },
		convert:   convertDecimal,
	},
	{
		types: floatTypes, isNumeric: true,
		nullValue: func() any { return sql.NullFloat64{Valid: false} },
		convert:   convertFloat,
	},
	{
		types:     boolTypes,
		nullValue: func() any { return sql.NullBool{Valid: false} },
		convert:   convertBool,
	},
	{
		types:     dateTypes,
		nullValue: func() any { return sql.NullTime{Valid: false} },
		convert:   convertDate,
	},
	{
		types:     dateTimeTypes,
		nullValue: func() any { return sql.NullTime{Valid: false} },
		convert:   convertDateTime,
	},
	{
		types:     binaryTypes,
		nullValue: func() any { return sql.NullString{Valid: false} },
		convert:   convertBinary,
	},
	{
		types:     uuidTypes,
		nullValue: func() any { return sql.NullString{Valid: false} },
		convert:   convertUUID,
	},
	{
		types:     jsonTypes,
		nullValue: func() any { return sql.NullString{Valid: false} },
		convert:   convertJSON,
	},
	{
		types:     networkTypes,
		nullValue: func() any { return sql.NullString{Valid: false} },
		convert:   convertNetwork,
	},
	{
		types:     geometryTypes,
		nullValue: func() any { return sql.NullString{Valid: false} },
		convert:   convertString,
	},
	{
		types:     xmlTypes,
		nullValue: func() any { return sql.NullString{Valid: false} },
		convert:   convertString,
	},
}
func (p *GormPlugin) ConvertStringValue(value, columnType string, isNullable bool) (any, error) {
	columnType = strings.ToUpper(columnType)
	columnType = p.NormalizeType(columnType)
	baseType := common.ParseTypeSpec(columnType).BaseType
	if customValue, handled, err := p.GormPluginFunctions.HandleCustomDataType(value, columnType, isNullable); handled {
		return customValue, err
	}
	for i := range typeConverters {
		tc := &typeConverters[i]
		if !tc.types.Contains(baseType) {
			continue
		}
		if isNullable && (value == "" || strings.EqualFold(value, "NULL")) {
			return tc.nullValue(), nil
		}
		if value == "" && tc.isNumeric {
			return nil, fmt.Errorf("cannot convert empty string to %s (column is not nullable)", columnType)
		}
		return tc.convert(p, value, baseType, columnType, isNullable)
	}
	if isNullable && (value == "" || strings.EqualFold(value, "NULL")) {
		return sql.NullString{Valid: false}, nil
	}
	if isNullable {
		return sql.NullString{String: value, Valid: true}, nil
	}
	return value, nil
}
func convertInt(_ *GormPlugin, value, _, columnType string, isNullable bool) (any, error) {
	parsedValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	if isNullable {
		return sql.NullInt64{Int64: parsedValue, Valid: true}, nil
	}
	return parsedValue, nil
}
func convertUint(_ *GormPlugin, value, baseType, _ string, isNullable bool) (any, error) {
	bitSize := 64
	if len(baseType) > 4 {
		if size, err := strconv.Atoi(baseType[4:]); err == nil {
			bitSize = size
		}
	}
	parsedValue, err := strconv.ParseUint(value, 10, bitSize)
	if err != nil {
		return nil, err
	}
	if isNullable {
		return &parsedValue, nil
	}
	return parsedValue, nil
}
func convertBigInt(_ *GormPlugin, value, _, _ string, isNullable bool) (any, error) {
	parsedValue := new(big.Int)
	if _, ok := parsedValue.SetString(value, 10); !ok {
		return nil, fmt.Errorf("invalid big integer value: %s", value)
	}
	return parsedValue, nil
}
func convertDecimal(_ *GormPlugin, value, _, _ string, isNullable bool) (any, error) {
	parsedValue, err := decimal.NewFromString(value)
	if err != nil {
		return nil, err
	}
	if isNullable {
		return &parsedValue, nil
	}
	return parsedValue, nil
}
func convertFloat(_ *GormPlugin, value, _, _ string, isNullable bool) (any, error) {
	parsedValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	if isNullable {
		return sql.NullFloat64{Float64: parsedValue, Valid: true}, nil
	}
	return parsedValue, nil
}
func convertBool(_ *GormPlugin, value, _, _ string, isNullable bool) (any, error) {
	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	if isNullable {
		return sql.NullBool{Bool: parsedValue, Valid: true}, nil
	}
	return parsedValue, nil
}
func convertDate(p *GormPlugin, value, _, _ string, isNullable bool) (any, error) {
	date, err := p.parseDate(value)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %v", err)
	}
	if isNullable {
		return sql.NullTime{Time: date, Valid: true}, nil
	}
	return date, nil
}
func convertDateTime(p *GormPlugin, value, baseType, _ string, isNullable bool) (any, error) {
	if baseType == "TIME" || baseType == "TIME WITH TIME ZONE" || baseType == "YEAR" {
		if isNullable {
			return sql.NullString{String: value, Valid: true}, nil
		}
		return value, nil
	}
	datetime, err := p.parseDateTime(value)
	if err != nil {
		return nil, fmt.Errorf("invalid datetime format: %v", err)
	}
	if isNullable {
		return sql.NullTime{Time: datetime, Valid: true}, nil
	}
	return datetime, nil
}
func DecodeHexLiteral(value string) ([]byte, bool, error) {
	if !strings.HasPrefix(value, "0x") && !strings.HasPrefix(value, "0X") {
		return nil, false, nil
	}
	data, err := hex.DecodeString(value[2:])
	return data, true, err
}
func NullableStringValue(value string, isNullable bool) any {
	if isNullable && (value == "" || strings.EqualFold(value, "NULL")) {
		return sql.NullString{Valid: false}
	}
	if isNullable {
		return sql.NullString{String: value, Valid: true}
	}
	return value
}
func convertBinary(_ *GormPlugin, value, _, _ string, isNullable bool) (any, error) {
	var blobData []byte
	if decoded, hasHexPrefix, err := DecodeHexLiteral(value); hasHexPrefix {
		if err != nil {
			return nil, fmt.Errorf("invalid hex binary format: %v", err)
		}
		blobData = decoded
	} else {
		blobData = []byte(value)
	}
	if isNullable && len(blobData) == 0 {
		return sql.NullString{Valid: false}, nil
	}
	return blobData, nil
}
func convertUUID(_ *GormPlugin, value, _, _ string, isNullable bool) (any, error) {
	if _, err := uuid.Parse(value); err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}
	if isNullable {
		return sql.NullString{String: value, Valid: true}, nil
	}
	return value, nil
}
func convertJSON(_ *GormPlugin, value, _, _ string, _ bool) (any, error) {
	if !json.Valid([]byte(value)) {
		return nil, fmt.Errorf("invalid JSON format")
	}
	return value, nil
}
func convertNetwork(_ *GormPlugin, value, baseType, _ string, isNullable bool) (any, error) {
	if baseType == "IPV4" || baseType == "IPV6" || baseType == "INET" || baseType == "CIDR" {
		ipStr := value
		if strings.Contains(value, "/") {
			ipStr = strings.Split(value, "/")[0]
		}
		if ip := net.ParseIP(ipStr); ip == nil {
			return nil, fmt.Errorf("invalid IP address format: %s", value)
		}
	}
	if isNullable {
		return sql.NullString{String: value, Valid: true}, nil
	}
	return value, nil
}
func convertString(_ *GormPlugin, value, _, _ string, isNullable bool) (any, error) {
	if isNullable {
		return sql.NullString{String: value, Valid: true}, nil
	}
	return value, nil
}
func (p *GormPlugin) parseDateTime(value string) (time.Time, error) {
	c := carbon.Parse(value)
	if c.Error != nil {
		return time.Time{}, fmt.Errorf("could not parse datetime '%s': %v", value, c.Error)
	}
	if c.IsInvalid() {
		return time.Time{}, fmt.Errorf("could not parse datetime '%s': invalid date", value)
	}
	return c.StdTime(), nil
}
func (p *GormPlugin) parseDate(value string) (time.Time, error) {
	c := carbon.Parse(value)
	if c.Error != nil {
		return time.Time{}, fmt.Errorf("could not parse date '%s': %v", value, c.Error)
	}
	if c.IsInvalid() {
		return time.Time{}, fmt.Errorf("could not parse date '%s': invalid date", value)
	}
	return c.StartOfDay().StdTime(), nil
}
func (p *GormPlugin) GetLastInsertID(db *gorm.DB) (int64, error) {
	return 0, nil
}
func (p *GormPlugin) ConvertArrayValue(value string, columnType string) (any, error) {
	upperType := strings.ToUpper(columnType)
	if strings.HasPrefix(upperType, "ARRAY(") {
		columnType = columnType[6:]
	}
	if strings.HasSuffix(columnType, ")") {
		columnType = columnType[:len(columnType)-1]
	}
	elementType := columnType
	value = strings.Trim(value, "[]")
	if value == "" {
		return []any{}, nil
	}
	elements := strings.Split(value, ",")
	result := make([]any, 0, len(elements))
	for _, element := range elements {
		element = strings.TrimSpace(element)
		if element == "" {
			continue
		}
		converted, err := p.GormPluginFunctions.ConvertStringValue(element, elementType, false)
		if err != nil {
			log.WithError(err).WithField("element", element).WithField("elementType", elementType).Error("Failed to convert array element")
			return nil, fmt.Errorf("converting array element: %w", err)
		}
		result = append(result, converted)
	}
	return result, nil
}