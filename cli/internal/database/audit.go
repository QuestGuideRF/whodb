package database
import (
	"fmt"
	"strconv"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
)
type AuditConfig struct {
	NullWarningPct    float64
	NullErrorPct      float64
	LowCardinalityMax int
}
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		NullWarningPct:    10,
		NullErrorPct:      50,
		LowCardinalityMax: 5,
	}
}
type AuditSeverity string
const (
	SeverityOK AuditSeverity = "ok"
	SeverityWarning AuditSeverity = "warning"
	SeverityError AuditSeverity = "error"
)
type ColumnAudit struct {
	Name          string
	Type          string
	IsPrimary     bool
	NullCount     int
	NullPct       float64
	DistinctCount int
	TotalRows     int
	Severity      AuditSeverity
	Issues        []string
}
type FKAudit struct {
	SourceTable  string
	SourceColumn string
	TargetTable  string
	TargetColumn string
	OrphanCount  int
	Severity     AuditSeverity
}
type DuplicateAudit struct {
	Columns            []string
	DuplicateCount     int
	TotalDuplicateRows int
}
type TableAudit struct {
	TableName     string
	RowCount      int
	Columns       []ColumnAudit
	ForeignKeys   []FKAudit
	Duplicates    []DuplicateAudit
	HasPrimaryKey bool
	Issues        []AuditIssue
}
type AuditIssue struct {
	Severity AuditSeverity
	Message  string
	Query    string
}
func (m *Manager) AuditTable(schema, tableName string, config AuditConfig) (*TableAudit, error) {
	columns, err := m.GetColumns(schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns for %s: %w", tableName, err)
	}
	audit := &TableAudit{
		TableName: tableName,
	}
	for _, col := range columns {
		if col.IsPrimary {
			audit.HasPrimaryKey = true
			break
		}
	}
	if !audit.HasPrimaryKey {
		audit.Issues = append(audit.Issues, AuditIssue{
			Severity: SeverityError,
			Message:  fmt.Sprintf("Table %q has no primary key", tableName),
		})
	}
	m.auditNullRates(schema, tableName, columns, config, audit)
	m.auditDistinctCounts(schema, tableName, columns, config, audit)
	m.auditTypeMismatches(columns, audit)
	m.auditDuplicates(schema, tableName, columns, audit)
	m.auditOrphanedFKs(schema, tableName, columns, audit)
	return audit, nil
}
func (m *Manager) AuditSchema(schema string, config AuditConfig) ([]*TableAudit, error) {
	units, err := m.GetStorageUnits(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	var results []*TableAudit
	for _, unit := range units {
		audit, err := m.AuditTable(schema, unit.Name, config)
		if err != nil {
			results = append(results, &TableAudit{
				TableName: unit.Name,
				Issues: []AuditIssue{{
					Severity: SeverityError,
					Message:  fmt.Sprintf("Audit failed: %v", err),
				}},
			})
			continue
		}
		results = append(results, audit)
	}
	return results, nil
}
func (m *Manager) auditNullRates(schema, tableName string, columns []engine.Column, config AuditConfig, audit *TableAudit) {
	if len(columns) == 0 {
		return
	}
	var countExprs []string
	countExprs = append(countExprs, "count(*) AS total_rows")
	for _, col := range columns {
		quoted := quoteIdent(col.Name)
		countExprs = append(countExprs, fmt.Sprintf("count(*) - count(%s) AS %s",
			quoted, quoteIdent("null_"+col.Name)))
	}
	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(countExprs, ", "),
		qualifiedTable(schema, tableName))
	result, err := m.ExecuteQuery(query)
	if err != nil {
		audit.Issues = append(audit.Issues, AuditIssue{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Failed to check null rates: %v", err),
			Query:    query,
		})
		for _, col := range columns {
			audit.Columns = append(audit.Columns, ColumnAudit{
				Name:      col.Name,
				Type:      col.Type,
				IsPrimary: col.IsPrimary,
				Severity:  SeverityOK,
			})
		}
		return
	}
	if len(result.Rows) == 0 || len(result.Rows[0]) < 1+len(columns) {
		for _, col := range columns {
			audit.Columns = append(audit.Columns, ColumnAudit{
				Name:      col.Name,
				Type:      col.Type,
				IsPrimary: col.IsPrimary,
				Severity:  SeverityOK,
			})
		}
		return
	}
	row := result.Rows[0]
	totalRows := parseIntSafe(row[0])
	audit.RowCount = totalRows
	for i, col := range columns {
		nullCount := parseIntSafe(row[i+1])
		var nullPct float64
		if totalRows > 0 {
			nullPct = float64(nullCount) / float64(totalRows) * 100
		}
		colAudit := ColumnAudit{
			Name:      col.Name,
			Type:      col.Type,
			IsPrimary: col.IsPrimary,
			NullCount: nullCount,
			NullPct:   nullPct,
			TotalRows: totalRows,
			Severity:  SeverityOK,
		}
		if nullPct >= config.NullErrorPct {
			colAudit.Severity = SeverityError
			colAudit.Issues = append(colAudit.Issues,
				fmt.Sprintf("%.1f%% null values (threshold: %.0f%%)", nullPct, config.NullErrorPct))
		} else if nullPct >= config.NullWarningPct {
			colAudit.Severity = SeverityWarning
			colAudit.Issues = append(colAudit.Issues,
				fmt.Sprintf("%.1f%% null values (threshold: %.0f%%)", nullPct, config.NullWarningPct))
		}
		audit.Columns = append(audit.Columns, colAudit)
	}
}
func (m *Manager) auditDistinctCounts(schema, tableName string, columns []engine.Column, config AuditConfig, audit *TableAudit) {
	for i, col := range columns {
		quoted := quoteIdent(col.Name)
		query := fmt.Sprintf("SELECT count(DISTINCT %s) FROM %s",
			quoted, qualifiedTable(schema, tableName))
		result, err := m.ExecuteQuery(query)
		if err != nil {
			continue
		}
		if len(result.Rows) == 0 || len(result.Rows[0]) == 0 {
			continue
		}
		distinctCount := parseIntSafe(result.Rows[0][0])
		if i < len(audit.Columns) {
			audit.Columns[i].DistinctCount = distinctCount
			if distinctCount <= config.LowCardinalityMax && !isBooleanLike(col.Type, col.Name) && audit.RowCount > config.LowCardinalityMax {
				if audit.Columns[i].Severity == SeverityOK {
					audit.Columns[i].Severity = SeverityWarning
				}
				audit.Columns[i].Issues = append(audit.Columns[i].Issues,
					fmt.Sprintf("Low cardinality: only %d distinct values", distinctCount))
			}
		}
	}
}
func (m *Manager) auditTypeMismatches(columns []engine.Column, audit *TableAudit) {
	for i, col := range columns {
		nameLower := strings.ToLower(col.Name)
		typeLower := strings.ToLower(col.Type)
		if strings.HasSuffix(nameLower, "_id") && isTextType(typeLower) {
			issue := fmt.Sprintf("Column %q ends with _id but has type %s", col.Name, col.Type)
			audit.Issues = append(audit.Issues, AuditIssue{
				Severity: SeverityWarning,
				Message:  issue,
			})
			if i < len(audit.Columns) {
				if audit.Columns[i].Severity == SeverityOK {
					audit.Columns[i].Severity = SeverityWarning
				}
				audit.Columns[i].Issues = append(audit.Columns[i].Issues, issue)
			}
		}
		if (strings.HasSuffix(nameLower, "_at") || strings.HasSuffix(nameLower, "_date")) &&
			(isTextType(typeLower) || isIntegerType(typeLower)) {
			issue := fmt.Sprintf("Column %q looks temporal but has type %s", col.Name, col.Type)
			audit.Issues = append(audit.Issues, AuditIssue{
				Severity: SeverityWarning,
				Message:  issue,
			})
			if i < len(audit.Columns) {
				if audit.Columns[i].Severity == SeverityOK {
					audit.Columns[i].Severity = SeverityWarning
				}
				audit.Columns[i].Issues = append(audit.Columns[i].Issues, issue)
			}
		}
	}
}
func (m *Manager) auditDuplicates(schema, tableName string, columns []engine.Column, audit *TableAudit) {
	var candidateCol string
	for _, col := range columns {
		if col.IsPrimary {
			continue
		}
		nameLower := strings.ToLower(col.Name)
		if strings.Contains(nameLower, "id") ||
			strings.Contains(nameLower, "email") ||
			strings.Contains(nameLower, "key") ||
			strings.Contains(nameLower, "code") ||
			strings.Contains(nameLower, "name") {
			candidateCol = col.Name
			break
		}
	}
	if candidateCol == "" {
		return
	}
	quoted := quoteIdent(candidateCol)
	query := fmt.Sprintf("SELECT count(*) AS dup_groups, COALESCE(sum(cnt), 0) AS total_dup_rows FROM (SELECT %s, count(*) AS cnt FROM %s.%s GROUP BY %s HAVING count(*) > 1) subq",
		quoted, quoteIdent(schema), quoteIdent(tableName), quoted)
	result, err := m.ExecuteQuery(query)
	if err != nil {
		return
	}
	if len(result.Rows) == 0 || len(result.Rows[0]) < 2 {
		return
	}
	dupGroups := parseIntSafe(result.Rows[0][0])
	totalDupRows := parseIntSafe(result.Rows[0][1])
	if dupGroups > 0 {
		audit.Duplicates = append(audit.Duplicates, DuplicateAudit{
			Columns:            []string{candidateCol},
			DuplicateCount:     dupGroups,
			TotalDuplicateRows: totalDupRows,
		})
		audit.Issues = append(audit.Issues, AuditIssue{
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("%d duplicate groups in column %q (%d rows)", dupGroups, candidateCol, totalDupRows),
			Query: fmt.Sprintf("SELECT %s, count(*) AS cnt FROM %s.%s GROUP BY %s HAVING count(*) > 1",
				quoted, quoteIdent(schema), quoteIdent(tableName), quoted),
		})
	}
}
func (m *Manager) auditOrphanedFKs(schema, tableName string, columns []engine.Column, audit *TableAudit) {
	for _, col := range columns {
		if !col.IsForeignKey || col.ReferencedTable == nil || col.ReferencedColumn == nil {
			continue
		}
		refTable := *col.ReferencedTable
		refCol := *col.ReferencedColumn
		query := fmt.Sprintf(
			"SELECT count(*) FROM %s t LEFT JOIN %s r ON t.%s = r.%s WHERE r.%s IS NULL AND t.%s IS NOT NULL",
			qualifiedTable(schema, tableName),
			qualifiedTable(schema, refTable),
			quoteIdent(col.Name), quoteIdent(refCol),
			quoteIdent(refCol), quoteIdent(col.Name),
		)
		result, err := m.ExecuteQuery(query)
		if err != nil {
			continue
		}
		if len(result.Rows) == 0 || len(result.Rows[0]) == 0 {
			continue
		}
		orphanCount := parseIntSafe(result.Rows[0][0])
		severity := SeverityOK
		if orphanCount > 0 {
			severity = SeverityError
			audit.Issues = append(audit.Issues, AuditIssue{
				Severity: SeverityError,
				Message:  fmt.Sprintf("%d orphaned FK references: %s.%s -> %s.%s", orphanCount, tableName, col.Name, refTable, refCol),
				Query:    query,
			})
		}
		audit.ForeignKeys = append(audit.ForeignKeys, FKAudit{
			SourceTable:  tableName,
			SourceColumn: col.Name,
			TargetTable:  refTable,
			TargetColumn: refCol,
			OrphanCount:  orphanCount,
			Severity:     severity,
		})
	}
}
func quoteIdent(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return `"` + escaped + `"`
}
func qualifiedTable(schema, table string) string {
	if schema == "" {
		return quoteIdent(table)
	}
	return quoteIdent(schema) + "." + quoteIdent(table)
}
func parseIntSafe(s string) int {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	n, _ := strconv.Atoi(s)
	return n
}
func isBooleanLike(colType, colName string) bool {
	typeLower := strings.ToLower(colType)
	nameLower := strings.ToLower(colName)
	if strings.Contains(typeLower, "bool") || typeLower == "bit" || typeLower == "tinyint(1)" {
		return true
	}
	if strings.HasPrefix(nameLower, "is_") || strings.HasPrefix(nameLower, "has_") {
		return true
	}
	return false
}
func isTextType(typeLower string) bool {
	return strings.Contains(typeLower, "text") ||
		strings.Contains(typeLower, "varchar") ||
		strings.Contains(typeLower, "char") ||
		strings.Contains(typeLower, "string")
}
func isIntegerType(typeLower string) bool {
	return strings.Contains(typeLower, "int") ||
		strings.Contains(typeLower, "integer") ||
		strings.Contains(typeLower, "bigint")
}