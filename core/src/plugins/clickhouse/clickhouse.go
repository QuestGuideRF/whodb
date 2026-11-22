package clickhouse
import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	sourcecatalogspecs "github.com/clidey/whodb/core/src/sourcecatalog/specs"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)
var (
	supportedOperators = sourcecatalogspecs.ClickHouseSupportedOperators
)
type ClickHousePlugin struct {
	gorm_plugin.GormPlugin
}
func (p *ClickHousePlugin) GetDatabases(config *engine.PluginConfig) ([]string, error) {
	return plugins.WithConnection[[]string](config, p.DB, func(db *gorm.DB) ([]string, error) {
		var databases []struct {
			Datname string `gorm:"column:datname"`
		}
		if err := db.Table("system.databases").
			Select("name AS datname").
			Scan(&databases).Error; err != nil {
			return nil, err
		}
		databaseNames := []string{}
		for _, database := range databases {
			databaseNames = append(databaseNames, database.Datname)
		}
		return databaseNames, nil
	})
}
func (p *ClickHousePlugin) GetSupportedOperators() map[string]string {
	return supportedOperators
}
func (p *ClickHousePlugin) HandleCustomDataType(value string, columnType string, isNullable bool) (any, bool, error) {
	upper := strings.ToUpper(columnType)
	var convert func(string, string) (any, error)
	switch {
	case strings.HasPrefix(upper, "ARRAY("):
		convert = convertArrayLiteral
	case strings.HasPrefix(upper, "MAP("):
		convert = convertMapLiteral
	case strings.HasPrefix(upper, "TUPLE("):
		convert = convertTupleLiteral
	default:
		return nil, false, nil
	}
	if isNullable && (value == "" || strings.EqualFold(value, "NULL")) {
		return nil, true, nil
	}
	result, err := convert(value, columnType)
	return result, true, err
}
func (p *ClickHousePlugin) ConvertStringValue(value, columnType string, isNullable bool) (any, error) {
	upperCheck := strings.ToUpper(columnType)
	if strings.HasPrefix(upperCheck, "NULLABLE(") {
		isNullable = true
		columnType = columnType[9:]
		if strings.HasSuffix(columnType, ")") {
			columnType = columnType[:len(columnType)-1]
		}
	}
	if strings.HasPrefix(strings.ToUpper(columnType), "LOWCARDINALITY(") {
		columnType = columnType[15:]
		if strings.HasSuffix(columnType, ")") {
			columnType = columnType[:len(columnType)-1]
		}
	}
	normalized := strings.ToUpper(p.NormalizeType(columnType))
	if strings.HasPrefix(normalized, "ARRAY") {
		if !strings.Contains(columnType, "(") {
			return []any{value}, nil
		}
		return p.ConvertArrayValue(value, columnType)
	}
	if strings.Contains(normalized, "JSON") {
		if !json.Valid([]byte(value)) {
			return nil, fmt.Errorf("invalid JSON format")
		}
		return value, nil
	}
	return p.GormPlugin.ConvertStringValue(value, columnType, isNullable)
}
func (p *ClickHousePlugin) GetAllSchemas(config *engine.PluginConfig) ([]string, error) {
	return nil, errors.ErrUnsupported
}
func (p *ClickHousePlugin) GetTableInfoQuery() string {
	return `
		SELECT
			name,
			engine,
			total_rows,
			formatReadableSize(total_bytes) as total_size
		FROM system.tables
		WHERE database = ?
		AND name NOT LIKE '.inner%'
	`
}
func (p *ClickHousePlugin) GetStorageUnitExistsQuery() string {
	return `SELECT EXISTS(SELECT 1 FROM system.tables WHERE database = ? AND name = ?)`
}
func (p *ClickHousePlugin) GetTableNameAndAttributes(rows *sql.Rows) (string, []engine.Record) {
	var tableName, tableType string
	var totalRows *uint64
	var totalSize *string
	if err := rows.Scan(&tableName, &tableType, &totalRows, &totalSize); err != nil {
		log.WithError(err).Error("Failed to scan table name and attributes from ClickHouse system.tables query")
		return "", nil
	}
	rowCount := "0"
	if totalRows != nil {
		rowCount = strconv.FormatUint(*totalRows, 10)
	}
	size := "unknown"
	if totalSize != nil {
		size = *totalSize
	}
	attributes := []engine.Record{
		{Key: "Type", Value: tableType},
		{Key: "Total Size", Value: size},
		{Key: "Count", Value: rowCount},
	}
	return tableName, attributes
}
func (p *ClickHousePlugin) RawExecute(config *engine.PluginConfig, query string, params ...any) (*engine.GetRowsResult, error) {
	return p.executeRawSQL(config, query, params...)
}
func (p *ClickHousePlugin) ClearTableData(config *engine.PluginConfig, schema string, storageUnit string) (bool, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		builder := p.CreateSQLBuilder(db)
		tableName := builder.BuildFullTableName(schema, storageUnit)
		query := fmt.Sprintf("ALTER TABLE %s DELETE WHERE 1=1", tableName)
		err := db.Exec(query).Error
		if err != nil {
			if err.Error() == "driver: bad connection" {
				var result int
				if db.Raw("SELECT 1").Scan(&result).Error == nil {
					log.WithField("table", tableName).Debug("ClickHouse DELETE mutation accepted")
					return true, nil
				}
			}
			return false, err
		}
		log.WithField("table", tableName).Debug("ClickHouse DELETE executed")
		return true, nil
	})
}
func (p *ClickHousePlugin) UpdateStorageUnit(config *engine.PluginConfig, schema string, storageUnit string, values map[string]string, updatedColumns []string) (bool, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		columnTypeInfos, err := p.GetColumnTypes(db, schema, storageUnit)
		if err != nil {
			return false, err
		}
		arrayColumns := map[string]bool{}
		for _, col := range updatedColumns {
			if ct, ok := columnTypeInfos[col]; ok && strings.HasPrefix(strings.ToUpper(ct.Type), "ARRAY(") {
				arrayColumns[col] = true
			}
		}
		if len(arrayColumns) == 0 {
			result, err := p.GormPlugin.UpdateStorageUnit(config, schema, storageUnit, values, updatedColumns)
			if err != nil && err.Error() == "driver: bad connection" {
				var check int
				if db.Raw("SELECT 1").Scan(&check).Error == nil {
					return true, nil
				}
			}
			return result, err
		}
		pkColumns, err := p.GetPrimaryKeyColumns(db, schema, storageUnit)
		if err != nil {
			pkColumns = []string{}
		}
		conditions := make(map[string]any)
		convertedValues := make(map[string]any)
		for column, strValue := range values {
			isPK := slices.Contains(pkColumns, column)
			isUpdated := slices.Contains(updatedColumns, column)
			if !isPK && !isUpdated && len(pkColumns) > 0 {
				continue
			}
			colInfo := columnTypeInfos[column]
			converted, convErr := p.ConvertStringValue(strValue, colInfo.Type, colInfo.IsNullable)
			if convErr != nil {
				return false, fmt.Errorf("failed to convert value for column '%s' in table %s.%s: %w", column, schema, storageUnit, convErr)
			}
			if isUpdated && arrayColumns[column] && reflect.ValueOf(converted).Kind() == reflect.Slice {
				converted = arrayToExpr(converted)
			}
			if isPK {
				conditions[column] = converted
			} else if isUpdated {
				convertedValues[column] = converted
			} else if len(pkColumns) == 0 {
				conditions[column] = converted
			}
		}
		if len(convertedValues) == 0 {
			return true, nil
		}
		result := p.CreateSQLBuilder(db).UpdateQuery(schema, storageUnit, convertedValues, conditions)
		if result.Error != nil {
			if result.Error.Error() == "driver: bad connection" {
				var check int
				if db.Raw("SELECT 1").Scan(&check).Error == nil {
					return true, nil
				}
			}
			return false, result.Error
		}
		return true, nil
	})
}
func (p *ClickHousePlugin) executeRawSQL(config *engine.PluginConfig, query string, params ...any) (*engine.GetRowsResult, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (*engine.GetRowsResult, error) {
		if config != nil && config.MultiStatement && isMultiStatement(query) {
			return nil, engine.ErrMultiStatementUnsupported
		}
		rows, err := db.Raw(query, params...).Rows()
		if err != nil {
			if err.Error() == "driver: bad connection" {
				var result int
				if db.Raw("SELECT 1").Scan(&result).Error == nil {
					return &engine.GetRowsResult{
						Columns: []engine.Column{},
						Rows:    [][]string{},
					}, nil
				}
			}
			return nil, err
		}
		defer rows.Close()
		return p.ConvertRawToRows(rows)
	})
}
func (p *ClickHousePlugin) GetColumnCodec(typeName string) gorm_plugin.ColumnCodec {
	upper := strings.ToUpper(typeName)
	handled := strings.HasPrefix(upper, "NULLABLE(") ||
		strings.HasPrefix(upper, "LOWCARDINALITY(") ||
		strings.HasPrefix(upper, "ARRAY(") ||
		strings.HasPrefix(upper, "TUPLE(") ||
		strings.HasPrefix(upper, "MAP(") ||
		strings.HasPrefix(upper, "NESTED(") ||
		strings.HasPrefix(upper, "INT128") ||
		strings.HasPrefix(upper, "INT256") ||
		strings.HasPrefix(upper, "UINT128") ||
		strings.HasPrefix(upper, "UINT256") ||
		strings.HasPrefix(upper, "DECIMAL") ||
		strings.HasPrefix(upper, "FIXEDSTRING") ||
		strings.HasPrefix(upper, "ENUM") ||
		strings.Contains(upper, "DATETIME64") ||
		upper == "IPV4" ||
		upper == "IPV6" ||
		upper == "UUID" ||
		upper == "DATE32" ||
		upper == "JSON" ||
		upper == "POINT" ||
		upper == "RING" ||
		upper == "POLYGON" ||
		upper == "MULTIPOLYGON"
	if !handled {
		return nil
	}
	return gorm_plugin.ColumnCodecFuncs{
		NewScanner: func() any {
			if strings.HasPrefix(upper, "INT128") || strings.HasPrefix(upper, "INT256") || strings.HasPrefix(upper, "UINT128") || strings.HasPrefix(upper, "UINT256") {
				return new(big.Int)
			}
			var value any
			return &value
		},
		FormatValue: func(value any) (string, error) {
			if ptr, ok := value.(*any); ok && ptr != nil {
				actualValue := *ptr
				if actualValue == nil {
					return "", nil
				}
				upperType := upper
				switch v := actualValue.(type) {
				case *big.Int:
					if v == nil {
						return "", nil
					}
					return v.String(), nil
				case []string:
					quoted := make([]string, len(v))
					for i, s := range v {
						quoted[i] = "'" + s + "'"
					}
					return "[" + strings.Join(quoted, ", ") + "]", nil
				case []any:
					if strings.HasPrefix(upperType, "TUPLE") {
						return formatTuple(v), nil
					}
					return formatSlice(v), nil
				case map[string]any:
					return formatMap(v), nil
				case net.IP:
					return v.String(), nil
				case *net.IP:
					if v != nil {
						return v.String(), nil
					}
					return "", nil
				case uuid.UUID:
					return v.String(), nil
				case time.Time:
					if strings.Contains(upperType, "DATE") && !strings.Contains(upperType, "DATETIME") {
						return v.Format("2006-01-02"), nil
					}
					if strings.Contains(upperType, "DATETIME64") {
						return v.Format("2006-01-02 15:04:05.999999999"), nil
					}
					return v.Format("2006-01-02 15:04:05"), nil
				case decimal.Decimal:
					return v.String(), nil
				case []byte:
					if strings.HasPrefix(upperType, "FIXEDSTRING") {
						trimmed := strings.TrimRight(string(v), "\x00")
						return trimmed, nil
					}
					return fmt.Sprintf("0x%x", v), nil
				case string:
					if strings.HasPrefix(upperType, "FIXEDSTRING") {
						return strings.TrimRight(v, "\x00"), nil
					}
					return v, nil
				default:
					rv := reflect.ValueOf(actualValue)
					if rv.Kind() == reflect.Map {
						return formatReflectMap(rv, upperType), nil
					}
					if rv.Kind() == reflect.Slice {
						if strings.HasPrefix(upperType, "TUPLE") {
							return formatReflectTuple(rv), nil
						}
						return formatReflectSlice(rv, upperType), nil
					}
					if stringer, ok := actualValue.(fmt.Stringer); ok {
						return stringer.String(), nil
					}
					if marshaled, err := json.Marshal(actualValue); err == nil && json.Valid(marshaled) {
						return string(marshaled), nil
					}
					return fmt.Sprintf("%v", actualValue), nil
				}
			}
			return fmt.Sprintf("%v", value), nil
		},
	}
}
func (p *ClickHousePlugin) NormalizeType(typeName string) string {
	return common.NormalizeTypeWithMap(typeName, sourcecatalogspecs.ClickHouseAliasMap)
}
func (p *ClickHousePlugin) GetColumnTypes(db *gorm.DB, schema, tableName string) (map[string]gorm_plugin.ColumnTypeInfo, error) {
	migrator := gorm_plugin.NewMigratorHelper(db, p)
	return migrator.GetColumnTypes(tableName)
}
func (p *ClickHousePlugin) GetColumnsForTable(config *engine.PluginConfig, schema string, storageUnit string) ([]engine.Column, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) ([]engine.Column, error) {
		migrator := gorm_plugin.NewMigratorHelper(db, p)
		columns, err := migrator.GetOrderedColumns(storageUnit)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to get columns for table %s.%s", schema, storageUnit))
			return nil, err
		}
		primaryKeys, err := p.GetPrimaryKeyColumns(db, schema, storageUnit)
		if err != nil {
			log.WithError(err).Warn(fmt.Sprintf("Failed to get primary keys for table %s.%s", schema, storageUnit))
			primaryKeys = []string{}
		}
		for i := range columns {
			columns[i].IsAutoIncrement = false
			if columns[i].Length == nil {
				typeSpec := common.ParseTypeSpec(columns[i].Type)
				if typeSpec.Length > 0 {
					columns[i].Length = &typeSpec.Length
				}
			}
			for _, pk := range primaryKeys {
				if columns[i].Name == pk {
					columns[i].IsPrimary = true
					break
				}
			}
		}
		return columns, nil
	})
}
func (p *ClickHousePlugin) WithTransaction(config *engine.PluginConfig, operation func(tx any) error) error {
	return operation(nil)
}
func (p *ClickHousePlugin) ResolveGraphSchema(config *engine.PluginConfig, schema string) string {
	return config.Credentials.Database
}
func (p *ClickHousePlugin) ShouldCheckRowsAffected() bool {
	return false
}
func init() {
	engine.RegisterPlugin(NewClickHousePlugin())
}
func NewClickHousePlugin() *engine.Plugin {
	clickhousePlugin := &ClickHousePlugin{}
	clickhousePlugin.Type = engine.DatabaseType_ClickHouse
	clickhousePlugin.PluginFunctions = clickhousePlugin
	clickhousePlugin.GormPluginFunctions = clickhousePlugin
	return &clickhousePlugin.Plugin
}