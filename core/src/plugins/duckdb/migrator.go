package duckdb
import (
	"database/sql"
	"strings"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)
type DuckDBMigrator struct {
	migrator.Migrator
}
func (m DuckDBMigrator) ColumnTypes(value any) ([]gorm.ColumnType, error) {
	var resolved string
	switch v := value.(type) {
	case string:
		resolved = v
	default:
		stmt := &gorm.Statement{DB: m.DB}
		if err := stmt.Parse(value); err != nil {
			return nil, err
		}
		resolved = stmt.Table
	}
	schemaName := "main"
	tableName := resolved
	if parts := strings.SplitN(resolved, ".", 2); len(parts) == 2 {
		schemaName = parts[0]
		tableName = parts[1]
	}
	query := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			CASE WHEN tc.constraint_type = 'PRIMARY KEY' THEN 'YES' ELSE 'NO' END AS is_primary,
			CASE WHEN c.column_default LIKE 'nextval%%' OR c.is_identity = 'YES' THEN 'YES' ELSE 'NO' END AS is_autoincrement
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu
			ON c.table_schema = kcu.table_schema AND c.table_name = kcu.table_name AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints tc
			ON kcu.constraint_name = tc.constraint_name AND kcu.table_schema = tc.table_schema AND tc.constraint_type = 'PRIMARY KEY'
		WHERE c.table_schema = ? AND c.table_name = ?
		ORDER BY c.ordinal_position
	`
	rows, err := m.DB.Raw(query, schemaName, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var columnTypes []gorm.ColumnType
	for rows.Next() {
		var (
			name            string
			dataType        string
			nullable        string
			columnDefault   sql.NullString
			charMaxLen      sql.NullInt64
			numericPrec     sql.NullInt64
			numericScale    sql.NullInt64
			isPrimary       string
			isAutoIncrement string
		)
		if err := rows.Scan(&name, &dataType, &nullable, &columnDefault, &charMaxLen, &numericPrec, &numericScale, &isPrimary, &isAutoIncrement); err != nil {
			return nil, err
		}
		if !charMaxLen.Valid {
			charMaxLen = sql.NullInt64{Int64: 0, Valid: true}
		}
		if !numericPrec.Valid {
			numericPrec = sql.NullInt64{Int64: 0, Valid: true}
		}
		if !numericScale.Valid {
			numericScale = sql.NullInt64{Int64: 0, Valid: true}
		}
		ct := migrator.ColumnType{
			NameValue:          sql.NullString{String: name, Valid: true},
			DataTypeValue:      sql.NullString{String: dataType, Valid: true},
			ColumnTypeValue:    sql.NullString{String: dataType, Valid: true},
			NullableValue:      sql.NullBool{Bool: nullable == "YES", Valid: true},
			PrimaryKeyValue:    sql.NullBool{Bool: isPrimary == "YES", Valid: true},
			AutoIncrementValue: sql.NullBool{Bool: isAutoIncrement == "YES", Valid: true},
			DefaultValueValue:  columnDefault,
			LengthValue:        charMaxLen,
			DecimalSizeValue:   numericPrec,
			ScaleValue:         numericScale,
		}
		columnTypes = append(columnTypes, ct)
	}
	return columnTypes, rows.Err()
}