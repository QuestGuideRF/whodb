package gorm_plugin
import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/dromara/carbon/v2"
	"gorm.io/gorm"
)
type ColumnCodec interface {
	Scanner() any
	Format(scanner any) (string, error)
}
type ColumnCodecFuncs struct {
	NewScanner  func() any
	FormatValue func(scanner any) (string, error)
}
func (c ColumnCodecFuncs) Scanner() any {
	if c.NewScanner == nil {
		return nil
	}
	return c.NewScanner()
}
func (c ColumnCodecFuncs) Format(scanner any) (string, error) {
	if c.FormatValue == nil {
		return "", nil
	}
	return c.FormatValue(scanner)
}
func (p *GormPlugin) ConvertRawToRows(rows *sql.Rows) (*engine.GetRowsResult, error) {
	columns, typeMap, resultColumns, err := p.describeRawColumns(rows)
	if err != nil {
		return nil, err
	}
	result := &engine.GetRowsResult{
		Columns: resultColumns,
		Rows:    make([][]string, 0, 100),
	}
	for rows.Next() {
		row, err := p.scanRawRow(rows, columns, typeMap)
		if err != nil {
			return nil, err
		}
		result.Rows = append(result.Rows, row)
	}
	if err := rows.Err(); err != nil {
		log.WithError(err).Error("Failed while iterating result rows")
		return nil, err
	}
	result.TotalCount = int64(len(result.Rows))
	return result, nil
}
func (p *GormPlugin) describeRawColumns(rows *sql.Rows) ([]string, map[string]*sql.ColumnType, []engine.Column, error) {
	columns, err := rows.Columns()
	if err != nil {
		log.WithError(err).Error("Failed to get column names from result set")
		return nil, nil, nil, err
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.WithError(err).Error("Failed to get column types from result set")
		return nil, nil, nil, err
	}
	typeMap := make(map[string]*sql.ColumnType, len(columnTypes))
	for _, colType := range columnTypes {
		typeMap[colType.Name()] = colType
	}
	resultColumns := make([]engine.Column, 0, len(columns))
	for _, col := range columns {
		if colType, exists := typeMap[col]; exists {
			colTypeName := colType.DatabaseTypeName()
			if p.GormPluginFunctions.IsArrayType(colTypeName) {
				colTypeName = strings.Replace(colTypeName, "_", "[]", 1)
			}
			if customName := p.GormPluginFunctions.GetCustomColumnTypeName(col, colTypeName); customName != "" {
				colTypeName = customName
			}
			column := engine.Column{Name: col, Type: colTypeName}
			baseTypeName := strings.ToUpper(colType.DatabaseTypeName())
			if typesWithLength[baseTypeName] {
				if length, ok := colType.Length(); ok && length > 0 {
					l := int(length)
					column.Length = &l
					colTypeName = fmt.Sprintf("%s(%d)", colTypeName, length)
					column.Type = colTypeName
				}
			}
			if typesWithPrecision[baseTypeName] {
				if precision, scale, ok := colType.DecimalSize(); ok && precision > 0 {
					prec := int(precision)
					column.Precision = &prec
					if scale > 0 {
						s := int(scale)
						column.Scale = &s
						colTypeName = fmt.Sprintf("%s(%d,%d)", colType.DatabaseTypeName(), precision, scale)
					} else {
						colTypeName = fmt.Sprintf("%s(%d)", colType.DatabaseTypeName(), precision)
					}
					column.Type = colTypeName
				}
			}
			resultColumns = append(resultColumns, column)
		}
	}
	return columns, typeMap, resultColumns, nil
}
func (p *GormPlugin) scanRawRow(rows *sql.Rows, columns []string, typeMap map[string]*sql.ColumnType) ([]string, error) {
	columnPointers := make([]any, len(columns))
	row := make([]string, len(columns))
	for i, col := range columns {
		colType := typeMap[col]
		typeName := ""
		if colType != nil {
			typeName = colType.DatabaseTypeName()
		}
		columnPointers[i] = CreateRawColumnScanner(p.GormPluginFunctions, typeName)
	}
	if err := rows.Scan(columnPointers...); err != nil {
		log.WithError(err).Error("Failed to scan row data")
		return nil, err
	}
	for i, colPtr := range columnPointers {
		colType := typeMap[columns[i]]
		typeName := ""
		if colType != nil {
			typeName = colType.DatabaseTypeName()
		}
		value, err := FormatScannedRawColumnValue(p.GormPluginFunctions, typeName, colPtr)
		if err != nil {
			row[i] = "ERROR: " + err.Error()
			continue
		}
		row[i] = value
	}
	return row, nil
}
func CreateRawColumnScanner(plugin GormPluginFunctions, columnType string) any {
	if codec := plugin.GetColumnCodec(columnType); codec != nil {
		return codec.Scanner()
	}
	switch columnType {
	case "VARBINARY", "BINARY", "IMAGE", "BYTEA", "BLOB", "HIERARCHYID":
		return new(sql.RawBytes)
	default:
		if plugin.IsGeometryType(columnType) {
			return new(sql.RawBytes)
		}
		return new(sql.NullString)
	}
}
func FormatScannedRawColumnValue(plugin GormPluginFunctions, columnType string, scanner any) (string, error) {
	if codec := plugin.GetColumnCodec(columnType); codec != nil {
		return codec.Format(scanner)
	}
	switch columnType {
	case "VARBINARY", "BINARY", "IMAGE", "BYTEA", "BLOB":
		rawBytes := scanner.(*sql.RawBytes)
		if rawBytes == nil || len(*rawBytes) == 0 {
			return "", nil
		}
		return "0x" + hex.EncodeToString(*rawBytes), nil
	case "TIME":
		val := scanner.(*sql.NullString)
		if !val.Valid {
			return "", nil
		}
		return formatTimeOnly(val.String), nil
	default:
		if plugin.IsGeometryType(columnType) {
			rawBytes := scanner.(*sql.RawBytes)
			if rawBytes == nil || len(*rawBytes) == 0 {
				return "", nil
			}
			if formatted := plugin.FormatGeometryValue(*rawBytes, columnType); formatted != "" {
				return formatted, nil
			}
			return "0x" + hex.EncodeToString(*rawBytes), nil
		}
		val := scanner.(*sql.NullString)
		if !val.Valid {
			return "", nil
		}
		return val.String, nil
	}
}
func (p *GormPlugin) FindMissingDataType(db *gorm.DB, columnType string) string {
	if p.Type == engine.DatabaseType_Postgres {
		var typname string
		if err := db.Table("pg_type").
			Select("typname").
			Where("oid = ?", columnType).
			Scan(&typname).Error; err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to find PostgreSQL type name for OID: %s", columnType))
			typname = columnType
		}
		return strings.ToUpper(typname)
	}
	return columnType
}
func (p *GormPlugin) GetRowsOrderBy(db *gorm.DB, schema string, storageUnit string) string {
	return ""
}
func (p *GormPlugin) GetColumnCodec(columnType string) ColumnCodec {
	return nil
}
func (p *GormPlugin) GetCustomColumnTypeName(columnName string, defaultTypeName string) string {
	return ""
}
func (p *GormPlugin) IsGeometryType(columnType string) bool {
	switch columnType {
	case "GEOMETRY", "POINT", "LINESTRING", "POLYGON", "GEOGRAPHY",
		"MULTIPOINT", "MULTILINESTRING", "MULTIPOLYGON", "GEOMETRYCOLLECTION":
		return true
	default:
		return false
	}
}
func (p *GormPlugin) FormatGeometryValue(rawBytes []byte, columnType string) string {
	return ""
}
func formatTimeOnly(value string) string {
	c := carbon.Parse(value)
	if c.Error != nil || c.IsInvalid() {
		return value
	}
	if c.Nanosecond() > 0 {
		return c.ToTimeMilliString()
	}
	return c.ToTimeString()
}
func (p *GormPlugin) HandleCustomDataType(value string, columnType string, isNullable bool) (any, bool, error) {
	return nil, false, nil
}
func (p *GormPlugin) IsArrayType(columnType string) bool {
	return false
}
func (p *GormPlugin) ResolveGraphSchema(config *engine.PluginConfig, schema string) string {
	return schema
}
func (p *GormPlugin) ShouldCheckRowsAffected() bool {
	return true
}