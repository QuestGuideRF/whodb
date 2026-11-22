package sqlite3
import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
	"gorm.io/gorm"
)
type DateTimeString string
func (ds *DateTimeString) Scan(value any) error {
	if value == nil {
		*ds = ""
		return nil
	}
	switch v := value.(type) {
	case string:
		*ds = DateTimeString(v)
	case []byte:
		*ds = DateTimeString(v)
	case driver.Value:
		str := fmt.Sprintf("%v", v)
		*ds = DateTimeString(str)
	default:
		if t, ok := value.(time.Time); ok {
			if t.IsZero() {
				*ds = ""
			} else if t.Nanosecond() > 0 {
				*ds = DateTimeString(t.Format("2006-01-02 15:04:05.999999999"))
			} else {
				*ds = DateTimeString(t.Format("2006-01-02 15:04:05"))
			}
		} else {
			str := fmt.Sprintf("%v", value)
			*ds = DateTimeString(str)
		}
	}
	return nil
}
func (ds DateTimeString) Value() (driver.Value, error) {
	return string(ds), nil
}
func (p *Sqlite3Plugin) ConvertStringValue(value, columnType string, isNullable bool) (any, error) {
	normalizedType := strings.ToUpper(columnType)
	if normalizedType == "DATE" || normalizedType == "DATETIME" || normalizedType == "TIMESTAMP" {
		return value, nil
	}
	return p.GormPlugin.ConvertStringValue(value, columnType, isNullable)
}
func (p *Sqlite3Plugin) GetPrimaryKeyColumns(db *gorm.DB, schema string, tableName string) ([]string, error) {
	query := p.GetPrimaryKeyColQuery()
	if query == "" {
		return nil, nil
	}
	rows, err := db.Raw(query, tableName).Rows()
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	var primaryKeys []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			continue
		}
		primaryKeys = append(primaryKeys, columnName)
	}
	return primaryKeys, nil
}
func (p *Sqlite3Plugin) GetPrimaryKeyColQuery() string {
	return `
		SELECT p.name AS pk_column
		FROM sqlite_master m,
			 pragma_table_info(m.name) p
		WHERE m.type = 'table'
		  AND m.name = ?
		  AND m.name NOT LIKE 'sqlite_%'
		  AND p.pk > 0
		ORDER BY m.name, p.pk;`
}