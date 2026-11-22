package clickhouse
import (
	"regexp"
	"strconv"
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
var enumValuePattern = regexp.MustCompile(`'([^']*)'`)
func (p *ClickHousePlugin) GetColumnConstraints(config *engine.PluginConfig, schema string, storageUnit string) (map[string]map[string]any, error) {
	constraints, err := p.GormPlugin.GetColumnConstraints(config, schema, storageUnit)
	if err != nil {
		constraints = make(map[string]map[string]any)
	}
	_, err = plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		var columns []struct {
			Name string `gorm:"column:name"`
			Type string `gorm:"column:type"`
		}
		if err := db.Table("system.columns").
			Select("name, type").
			Where("database = ? AND table = ?", schema, storageUnit).
			Scan(&columns).Error; err != nil {
			return false, err
		}
		for _, col := range columns {
			enrichConstraintsFromType(constraints, col.Name, col.Type)
		}
		return true, nil
	})
	if err != nil {
		return constraints, nil
	}
	return constraints, nil
}
func enrichConstraintsFromType(constraints map[string]map[string]any, columnName string, fullType string) {
	upperFull := strings.ToUpper(fullType)
	if !strings.Contains(upperFull, "NULLABLE(") {
		colConstraints := gorm_plugin.EnsureConstraintEntry(constraints, columnName)
		colConstraints["nullable"] = false
	}
	innerType := unwrapClickHouseModifiers(fullType)
	upperInner := strings.ToUpper(innerType)
	if strings.HasPrefix(upperInner, "ENUM8(") || strings.HasPrefix(upperInner, "ENUM16(") {
		values := parseClickHouseEnumValues(innerType)
		if len(values) > 0 {
			colConstraints := gorm_plugin.EnsureConstraintEntry(constraints, columnName)
			colConstraints["check_values"] = values
		}
		return
	}
	if strings.HasPrefix(upperInner, "DECIMAL") {
		scale, precision := parseClickHouseDecimalParams(innerType)
		if scale >= 0 || precision > 0 {
			colConstraints := gorm_plugin.EnsureConstraintEntry(constraints, columnName)
			if scale >= 0 {
				colConstraints["scale"] = scale
			}
			if precision > 0 {
				colConstraints["precision"] = precision
			}
		}
	}
}
func unwrapClickHouseModifiers(typeName string) string {
	for {
		lower := strings.ToLower(strings.TrimSpace(typeName))
		unwrapped := false
		for _, prefix := range []string{"nullable(", "lowcardinality("} {
			if strings.HasPrefix(lower, prefix) && strings.HasSuffix(lower, ")") {
				typeName = strings.TrimSpace(typeName[len(prefix) : len(typeName)-1])
				unwrapped = true
				break
			}
		}
		if !unwrapped {
			return typeName
		}
	}
}
func parseClickHouseEnumValues(enumType string) []string {
	start := strings.Index(enumType, "(")
	end := strings.LastIndex(enumType, ")")
	if start == -1 || end == -1 || end <= start {
		return nil
	}
	content := enumType[start+1 : end]
	matches := enumValuePattern.FindAllStringSubmatch(content, -1)
	var values []string
	for _, match := range matches {
		if len(match) > 1 {
			if sanitized, ok := common.SanitizeConstraintValue(match[1]); ok {
				values = append(values, sanitized)
			}
		}
	}
	return values
}
func parseClickHouseDecimalParams(decimalType string) (scale int, precision int) {
	scale = -1
	start := strings.Index(decimalType, "(")
	end := strings.LastIndex(decimalType, ")")
	if start == -1 || end == -1 || end <= start {
		return
	}
	baseName := strings.ToUpper(strings.TrimSpace(decimalType[:start]))
	content := decimalType[start+1 : end]
	parts := strings.Split(content, ",")
	if len(parts) == 1 {
		if s, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
			scale = s
		}
		switch baseName {
		case "DECIMAL32":
			precision = 9
		case "DECIMAL64":
			precision = 18
		case "DECIMAL128":
			precision = 38
		case "DECIMAL256":
			precision = 76
		}
	} else if len(parts) == 2 {
		if p, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
			precision = p
		}
		if s, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
			scale = s
		}
	}
	return
}