package mcp
import (
	"errors"
	"fmt"
	"strings"
)
type SecurityLevel string
const (
	SecurityLevelStrict   SecurityLevel = "strict"
	SecurityLevelStandard SecurityLevel = "standard"
	SecurityLevelMinimal  SecurityLevel = "minimal"
)
var (
	ErrWriteNotAllowed      = errors.New("write operations are not allowed in read-only mode")
	ErrMultipleStatements   = errors.New("multiple SQL statements are not allowed")
	ErrDangerousFunction    = errors.New("dangerous database function detected")
	ErrDestructiveOperation = errors.New("destructive operation detected")
)
type StatementType string
const (
	StatementSelect   StatementType = "SELECT"
	StatementInsert   StatementType = "INSERT"
	StatementUpdate   StatementType = "UPDATE"
	StatementDelete   StatementType = "DELETE"
	StatementDrop     StatementType = "DROP"
	StatementCreate   StatementType = "CREATE"
	StatementAlter    StatementType = "ALTER"
	StatementTruncate StatementType = "TRUNCATE"
	StatementShow     StatementType = "SHOW"
	StatementDescribe StatementType = "DESCRIBE"
	StatementExplain  StatementType = "EXPLAIN"
	StatementWith     StatementType = "WITH"
	StatementUnknown  StatementType = "UNKNOWN"
)
var dangerousFunctions = []string{
	"PG_TERMINATE_BACKEND", "PG_READ_FILE", "PG_WRITE_FILE", "LO_IMPORT", "LO_EXPORT",
	"LOAD_FILE", "INTO OUTFILE", "INTO DUMPFILE",
	"XP_CMDSHELL", "OPENROWSET", "OPENDATASOURCE",
	"LOAD_EXTENSION", "WRITEFILE", "READFILE",
	"URL(", "FILE(", "REMOTE(",
	"UTL_FILE", "UTL_HTTP",
	"COPY",
}
func DetectStatementType(query string) StatementType {
	q := strings.ToUpper(strings.TrimSpace(query))
	keywords := []struct {
		kw  string
		typ StatementType
	}{
		{"SELECT ", StatementSelect}, {"INSERT ", StatementInsert},
		{"UPDATE ", StatementUpdate}, {"DELETE ", StatementDelete},
		{"DROP ", StatementDrop}, {"CREATE ", StatementCreate},
		{"ALTER ", StatementAlter}, {"TRUNCATE ", StatementTruncate},
		{"SHOW ", StatementShow}, {"DESCRIBE ", StatementDescribe},
		{"DESC ", StatementDescribe}, {"EXPLAIN ", StatementExplain},
		{"WITH ", StatementWith},
	}
	for _, k := range keywords {
		if strings.HasPrefix(q, k.kw) {
			return k.typ
		}
	}
	return StatementUnknown
}
func ValidateSQLStatement(query string, allowWrite bool, securityLevel SecurityLevel, allowMultiStatement bool, allowDestructive bool) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("empty query")
	}
	if !allowDestructive {
		upper := strings.ToUpper(query)
		if strings.Contains(upper, "DROP ") {
			return fmt.Errorf("%w: DROP detected (use --confirm-writes or --allow-drop to enable)", ErrDestructiveOperation)
		}
		if strings.Contains(upper, "TRUNCATE ") {
			return fmt.Errorf("%w: TRUNCATE detected (use --confirm-writes or --allow-drop to enable)", ErrDestructiveOperation)
		}
	}
	if strings.Contains(query, "--") || strings.Contains(query, "/*") {
		return errors.New("SQL comments are not allowed")
	}
	if !allowMultiStatement {
		trimmed := strings.TrimSuffix(strings.TrimSpace(query), ";")
		if strings.Contains(trimmed, ";") {
			return ErrMultipleStatements
		}
	}
	if securityLevel == SecurityLevelStrict {
		upper := strings.ToUpper(query)
		for _, fn := range dangerousFunctions {
			if strings.Contains(upper, fn) {
				return fmt.Errorf("%w: %s", ErrDangerousFunction, fn)
			}
		}
	}
	stmtType := DetectStatementType(query)
	if allowWrite {
		if securityLevel == SecurityLevelMinimal {
			if stmtType == StatementDelete && !strings.Contains(strings.ToUpper(query), " WHERE ") {
				return fmt.Errorf("%w: DELETE without WHERE clause", ErrDestructiveOperation)
			}
		}
		return nil
	}
	switch stmtType {
	case StatementSelect, StatementShow, StatementDescribe, StatementExplain, StatementWith:
		return nil
	case StatementInsert, StatementUpdate, StatementDelete, StatementDrop,
		StatementCreate, StatementAlter, StatementTruncate:
		return fmt.Errorf("%w: %s", ErrWriteNotAllowed, stmtType)
	default:
		if securityLevel != SecurityLevelMinimal {
			return fmt.Errorf("%w: unknown statement type", ErrWriteNotAllowed)
		}
		return nil
	}
}
func IsReadOnlyStatement(stmtType StatementType) bool {
	switch stmtType {
	case StatementSelect, StatementShow, StatementDescribe, StatementExplain, StatementWith:
		return true
	}
	return false
}
func IsWriteStatement(stmtType StatementType) bool {
	switch stmtType {
	case StatementInsert, StatementUpdate, StatementDelete, StatementDrop,
		StatementCreate, StatementAlter, StatementTruncate:
		return true
	}
	return false
}
var (
	ErrQueryRequired      = errors.New("query is required and cannot be empty")
	ErrTableRequired      = errors.New("table is required - specify which table to describe")
	ErrTokenRequired      = errors.New("token is required - use the confirmation_token from the previous query response")
	ErrTokenInvalid       = errors.New("token is not a valid confirmation token format")
	ErrConnectionRequired = errors.New("connection is required when multiple connections are configured (use whodb_connections to list available connections)")
)
func ValidateQueryInput(input *QueryInput, connectionCount int) error {
	if strings.TrimSpace(input.Query) == "" {
		return ErrQueryRequired
	}
	if connectionCount > 1 && strings.TrimSpace(input.Connection) == "" {
		return ErrConnectionRequired
	}
	return nil
}
func ValidateSchemasInput(input *SchemasInput, connectionCount int) error {
	if connectionCount > 1 && strings.TrimSpace(input.Connection) == "" {
		return ErrConnectionRequired
	}
	return nil
}
func ValidateTablesInput(input *TablesInput, connectionCount int) error {
	if connectionCount > 1 && strings.TrimSpace(input.Connection) == "" {
		return ErrConnectionRequired
	}
	return nil
}
func ValidateColumnsInput(input *ColumnsInput, connectionCount int) error {
	if strings.TrimSpace(input.Table) == "" {
		return ErrTableRequired
	}
	if connectionCount > 1 && strings.TrimSpace(input.Connection) == "" {
		return ErrConnectionRequired
	}
	return nil
}
func ValidateConfirmInput(input *ConfirmInput) error {
	token := strings.TrimSpace(input.Token)
	if token == "" {
		return ErrTokenRequired
	}
	if len(token) != 36 || token[8] != '-' || token[13] != '-' || token[18] != '-' || token[23] != '-' {
		return ErrTokenInvalid
	}
	return nil
}