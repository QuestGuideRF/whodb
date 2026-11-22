package postgres
import (
	"regexp"
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
	"gorm.io/gorm"
)
var pgTypeCastPattern = regexp.MustCompile(`::\w+(\s+\w+)?`)
func (p *PostgresPlugin) GetColumnConstraints(config *engine.PluginConfig, schema string, storageUnit string) (map[string]map[string]any, error) {
	constraints := make(map[string]map[string]any)
	_, err := plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		fullTableName := schema + "." + storageUnit
		primaryRows, err := db.Raw(`
			SELECT a.attname AS column_name
			FROM pg_index i
			JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
			WHERE i.indrelid = $1::regclass AND i.indisprimary
		`, fullTableName).Rows()
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
			Select("column_name, is_nullable").
			Where("table_schema = ? AND table_name = ?", schema, storageUnit).
			Rows()
		if err != nil {
			return false, err
		}
		defer rows.Close()
		for rows.Next() {
			var columnName, isNullable string
			if err := rows.Scan(&columnName, &isNullable); err != nil {
				continue
			}
			gorm_plugin.EnsureConstraintEntry(constraints, columnName)["nullable"] = strings.EqualFold(isNullable, "YES")
		}
		uniqueRows, err := db.Table("pg_index i").
			Select("a.attname AS column_name").
			Joins("JOIN pg_class c ON c.oid = i.indrelid").
			Joins("JOIN pg_namespace n ON n.oid = c.relnamespace").
			Joins("JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(i.indkey)").
			Where("c.relname = ? AND n.nspname = ? AND i.indisunique = true AND i.indnkeyatts = 1", storageUnit, schema).
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
		checkRows, err := db.Table("pg_constraint").
			Select("conname AS constraint_name, pg_get_constraintdef(oid) AS check_clause").
			Where("contype = 'c' AND conrelid = ?::regclass", fullTableName).
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
		enumRows, err := db.Raw(`
			SELECT c.column_name, e.enumlabel
			FROM information_schema.columns c
			JOIN pg_type t ON t.typname = c.udt_name
			JOIN pg_enum e ON e.enumtypid = t.oid
			WHERE c.table_schema = $1 AND c.table_name = $2
			ORDER BY c.column_name, e.enumsortorder
		`, schema, storageUnit).Rows()
		if err == nil {
			defer enumRows.Close()
			enumValues := make(map[string][]string)
			for enumRows.Next() {
				var columnName, enumLabel string
				if err := enumRows.Scan(&columnName, &enumLabel); err != nil {
					continue
				}
				enumValues[columnName] = append(enumValues[columnName], enumLabel)
			}
			for columnName, values := range enumValues {
				gorm_plugin.EnsureConstraintEntry(constraints, columnName)["check_values"] = values
			}
		}
		return true, nil
	})
	if err != nil {
		return constraints, err
	}
	return constraints, nil
}
func (p *PostgresPlugin) parseCheckConstraint(checkClause string, constraints map[string]map[string]any) {
	log.WithField("checkClause", checkClause).Debug("Parsing CHECK constraint")
	clause := strings.TrimPrefix(checkClause, "CHECK ")
	clause = strings.Trim(clause, "()")
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
	if values := extractArrayValues(clause); len(values) > 0 {
		colConstraints["check_values"] = values
		log.WithField("column", columnName).WithField("values", values).Debug("Extracted check_values via ARRAY pattern")
		return
	}
	if values := gorm_plugin.ParseINClauseValues(clause); len(values) > 0 {
		colConstraints["check_values"] = values
		log.WithField("column", columnName).WithField("values", values).Debug("Extracted check_values via IN clause pattern")
	} else {
		log.WithField("column", columnName).WithField("clause", clause).Debug("No check_values extracted from clause")
	}
}
func extractArrayValues(clause string) []string {
	upper := strings.ToUpper(clause)
	arrayIdx := strings.Index(upper, "ARRAY[")
	if arrayIdx == -1 {
		return nil
	}
	startIdx := arrayIdx + 6
	depth := 1
	endIdx := -1
	for i := startIdx; i < len(clause); i++ {
		switch clause[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				endIdx = i
			}
		}
		if endIdx != -1 {
			break
		}
	}
	if endIdx == -1 || endIdx <= startIdx {
		return nil
	}
	content := clause[startIdx:endIdx]
	var values []string
	parts := strings.Split(content, ",")
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		cleaned = pgTypeCastPattern.ReplaceAllString(cleaned, "")
		cleaned = strings.Trim(cleaned, "'\"")
		if cleaned != "" {
			values = append(values, cleaned)
		}
	}
	return values
}