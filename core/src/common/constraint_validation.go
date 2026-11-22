package common
import (
	"regexp"
	"strings"
)
func ValidateColumnName(columnName string) bool {
	validColumnPattern := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validColumnPattern.MatchString(columnName) {
		return false
	}
	sqlKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER",
		"EXEC", "EXECUTE", "UNION", "GRANT", "REVOKE", "TRUNCATE",
		"MERGE", "CALL", "EXPLAIN", "LOCK", "COMMENT", "COMMIT", "ROLLBACK",
		"SAVEPOINT", "SET", "SHOW", "--", "/*", "*/", ";",
	}
	upperColumn := strings.ToUpper(columnName)
	for _, keyword := range sqlKeywords {
		if strings.Contains(upperColumn, keyword) {
			return false
		}
	}
	if len(columnName) > 128 {
		return false
	}
	return true
}
func SanitizeConstraintValue(value string) (string, bool) {
	if strings.Contains(value, "--") || strings.Contains(value, "/*") || strings.Contains(value, "*/") {
		return "", false
	}
	if strings.Contains(value, ";") {
		return "", false
	}
	dangerousKeywords := []string{
		"DROP", "DELETE", "TRUNCATE", "EXEC", "EXECUTE",
		"CREATE", "ALTER", "GRANT", "REVOKE",
		"UNION", "INSERT", "UPDATE",
	}
	upperValue := strings.ToUpper(value)
	for _, keyword := range dangerousKeywords {
		if regexp.MustCompile(`\b` + keyword + `\b`).MatchString(upperValue) {
			return "", false
		}
	}
	if len(value) > 1000 {
		return "", false
	}
	return value, true
}