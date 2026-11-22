package mysql
import (
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/plugins"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
func (p *MySQLPlugin) GetColumnConstraints(config *engine.PluginConfig, schema string, storageUnit string) (map[string]map[string]any, error) {
	constraints := make(map[string]map[string]any)
	_, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		primaryRows, err := db.Table("information_schema.table_constraints t").
			Select("k.COLUMN_NAME").
			Joins("JOIN information_schema.key_column_usage k ON k.CONSTRAINT_NAME = t.CONSTRAINT_NAME AND k.TABLE_SCHEMA = t.TABLE_SCHEMA AND k.TABLE_NAME = t.TABLE_NAME").
			Where("t.CONSTRAINT_TYPE = 'PRIMARY KEY' AND t.TABLE_SCHEMA = DATABASE() AND t.TABLE_NAME = ?", storageUnit).
			Order("k.ORDINAL_POSITION").
			Rows()
		if err == nil {
			defer primaryRows.Close()
			for primaryRows.Next() {
				var columnName string
				if err := primaryRows.Scan(&columnName); err != nil {
					continue
				}
				gorm_plugin.EnsureConstraintEntry(constraints, columnName)["primary"] = true
			}
		}
		rows, err := db.Table("information_schema.columns").
			Select("COLUMN_NAME, IS_NULLABLE, COLUMN_TYPE").
			Where("table_schema = DATABASE() AND table_name = ?", storageUnit).
			Rows()
		if err != nil {
			return false, err
		}
		defer rows.Close()
		for rows.Next() {
			var columnName, isNullable, columnType string
			if err := rows.Scan(&columnName, &isNullable, &columnType); err != nil {
				continue
			}
			colConstraints := gorm_plugin.EnsureConstraintEntry(constraints, columnName)
			colConstraints["nullable"] = strings.EqualFold(isNullable, "YES")
			if strings.HasPrefix(strings.ToLower(columnType), "enum(") {
				values := parseEnumValues(columnType)
				if len(values) > 0 {
					colConstraints["check_values"] = values
				}
			}
		}
		uniqueRows, err := db.Table("information_schema.statistics").
			Select("COLUMN_NAME").
			Where("table_schema = DATABASE() AND table_name = ? AND NON_UNIQUE = 0", storageUnit).
			Group("COLUMN_NAME, INDEX_NAME").
			Having("COUNT(*) = 1").
			Rows()
		if err != nil {
			return false, err
		}
		defer uniqueRows.Close()
		for uniqueRows.Next() {
			var columnName string
			if err := uniqueRows.Scan(&columnName); err != nil {
				continue
			}
			gorm_plugin.EnsureConstraintEntry(constraints, columnName)["unique"] = true
		}
		checkRows, err := db.Table("information_schema.CHECK_CONSTRAINTS cc").
			Select("cc.CONSTRAINT_NAME, cc.CHECK_CLAUSE").
			Joins("JOIN information_schema.TABLE_CONSTRAINTS tc ON cc.CONSTRAINT_SCHEMA = tc.CONSTRAINT_SCHEMA AND cc.CONSTRAINT_NAME = tc.CONSTRAINT_NAME").
			Where("cc.CONSTRAINT_SCHEMA = DATABASE() AND tc.TABLE_NAME = ?", storageUnit).
			Rows()
		if err == nil {
			defer checkRows.Close()
			for checkRows.Next() {
				var constraintName, checkClause string
				if err := checkRows.Scan(&constraintName, &checkClause); err != nil {
					continue
				}
				p.parseCheckConstraint(checkClause, constraints)
			}
		}
		return true, nil
	})
	if err != nil {
		return constraints, err
	}
	return constraints, nil
}
func (p *MySQLPlugin) parseCheckConstraint(checkClause string, constraints map[string]map[string]any) {
	lowerClause := strings.ToLower(checkClause)
	if strings.Contains(lowerClause, "json_valid(") {
		columnName := extractJSONValidColumn(checkClause)
		if columnName != "" && common.ValidateColumnName(columnName) {
			colConstraints := gorm_plugin.EnsureConstraintEntry(constraints, columnName)
			colConstraints["is_json"] = true
			return
		}
	}
	clause := strings.Trim(checkClause, "()")
	columnName := gorm_plugin.ExtractColumnNameFromClause(clause)
	if columnName == "" {
		return
	}
	if !common.ValidateColumnName(columnName) {
		return
	}
	colConstraints := gorm_plugin.EnsureConstraintEntry(constraints, columnName)
	minMax := gorm_plugin.ParseMinMaxConstraints(clause)
	gorm_plugin.ApplyMinMaxToConstraints(colConstraints, minMax)
	if values := gorm_plugin.ParseINClauseValues(clause); len(values) > 0 {
		colConstraints["check_values"] = values
	}
}
func extractJSONValidColumn(clause string) string {
	lowerClause := strings.ToLower(clause)
	idx := strings.Index(lowerClause, "json_valid(")
	if idx == -1 {
		return ""
	}
	start := idx + len("json_valid(")
	if start >= len(clause) {
		return ""
	}
	end := strings.Index(clause[start:], ")")
	if end == -1 {
		return ""
	}
	colRef := strings.TrimSpace(clause[start : start+end])
	colRef = strings.Trim(colRef, "`\"'")
	return colRef
}
func parseEnumValues(columnType string) []string {
	start := strings.Index(columnType, "(")
	end := strings.LastIndex(columnType, ")")
	if start == -1 || end == -1 || end <= start {
		return nil
	}
	content := columnType[start+1 : end]
	var values []string
	parts := strings.Split(content, ",")
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		cleaned = strings.Trim(cleaned, "'\"")
		if cleaned != "" {
			if sanitized, ok := common.SanitizeConstraintValue(cleaned); ok {
				values = append(values, sanitized)
			}
		}
	}
	return values
}