package gorm_plugin
import (
	"regexp"
	"strconv"
	"strings"
)
var (
	gtePattern        = regexp.MustCompile(`>=\s*\(?(-?\d+(?:\.\d+)?)\)?`)
	gtPattern         = regexp.MustCompile(`>[^=]\s*\(?(-?\d+(?:\.\d+)?)\)?|>\s*\(?(-?\d+(?:\.\d+)?)\)?$`)
	ltePattern        = regexp.MustCompile(`<=\s*\(?(-?\d+(?:\.\d+)?)\)?`)
	ltPattern         = regexp.MustCompile(`<[^=]\s*\(?(-?\d+(?:\.\d+)?)\)?|<\s*\(?(-?\d+(?:\.\d+)?)\)?$`)
	betweenPattern    = regexp.MustCompile(`(?i)between\s+\(?(-?\d+(?:\.\d+)?)\)?\s+and\s+\(?(-?\d+(?:\.\d+)?)\)?`)
	typeCastPattern   = regexp.MustCompile(`::\w+(\s+\w+)?(\[\])?`)
	columnNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`)
	orEqualityPattern = regexp.MustCompile(`(?:\[?\w+\]?)\s*=\s*(?:N)?['"]([^'"]+)['"]`)
)
func EnsureConstraintEntry(constraints map[string]map[string]any, columnName string) map[string]any {
	if constraints[columnName] == nil {
		constraints[columnName] = map[string]any{}
	}
	return constraints[columnName]
}
func ParseINClauseValues(clause string) []string {
	normalizedClause := strings.Join(strings.Fields(clause), " ")
	clauseLower := strings.ToLower(normalizedClause)
	inIdx := strings.Index(clauseLower, " in ")
	if inIdx == -1 {
		inIdx = strings.Index(clauseLower, " in(")
	}
	if inIdx == -1 {
		return nil
	}
	clause = normalizedClause
	startIdx := strings.Index(clause[inIdx:], "(")
	if startIdx == -1 {
		return nil
	}
	startIdx += inIdx + 1
	depth := 1
	endIdx := -1
	for i := startIdx; i < len(clause); i++ {
		switch clause[i] {
		case '(':
			depth++
		case ')':
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
	return ParseValueList(content)
}
func ParseORClauseValues(clause string) []string {
	clauseLower := strings.ToLower(clause)
	if !strings.Contains(clauseLower, " or ") {
		return nil
	}
	matches := orEqualityPattern.FindAllStringSubmatch(clause, -1)
	if len(matches) == 0 {
		return nil
	}
	values := make([]string, 0, len(matches))
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			val := match[1]
			if !seen[val] {
				values = append(values, val)
				seen[val] = true
			}
		}
	}
	if len(values) == 0 {
		return nil
	}
	return values
}
func ParseValueList(content string) []string {
	var values []string
	parts := strings.Split(content, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.TrimPrefix(part, "N")
		part = strings.TrimPrefix(part, "n")
		if idx := strings.Index(part, "'"); idx > 0 && part[0] == '_' {
			part = part[idx:]
		}
		part = typeCastPattern.ReplaceAllString(part, "")
		if len(part) >= 2 {
			if (part[0] == '\'' && part[len(part)-1] == '\'') ||
				(part[0] == '"' && part[len(part)-1] == '"') {
				values = append(values, part[1:len(part)-1])
			}
		}
	}
	return values
}
type MinMaxResult struct {
	Min    float64
	Max    float64
	HasMin bool
	HasMax bool
}
func ParseMinMaxConstraints(clause string) MinMaxResult {
	result := MinMaxResult{}
	clauseLower := strings.ToLower(clause)
	if matches := gtePattern.FindStringSubmatch(clause); len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result.Min = val
			result.HasMin = true
		}
	} else if matches := gtPattern.FindStringSubmatch(clause); len(matches) > 0 {
		numStr := getFirstNonEmptyMatch(matches[1:])
		if numStr != "" {
			if val, err := strconv.ParseFloat(numStr, 64); err == nil {
				if val == float64(int64(val)) {
					result.Min = val + 1
				} else {
					result.Min = val
				}
				result.HasMin = true
			}
		}
	}
	if matches := ltePattern.FindStringSubmatch(clause); len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			result.Max = val
			result.HasMax = true
		}
	} else if matches := ltPattern.FindStringSubmatch(clause); len(matches) > 0 {
		numStr := getFirstNonEmptyMatch(matches[1:])
		if numStr != "" && !strings.Contains(clause[strings.Index(clauseLower, "<"):], "=") {
			if val, err := strconv.ParseFloat(numStr, 64); err == nil {
				if val == float64(int64(val)) {
					result.Max = val - 1
				} else {
					result.Max = val
				}
				result.HasMax = true
			}
		}
	}
	if strings.Contains(clauseLower, "between") {
		if matches := betweenPattern.FindStringSubmatch(clause); len(matches) > 2 {
			if minVal, err := strconv.ParseFloat(matches[1], 64); err == nil {
				result.Min = minVal
				result.HasMin = true
			}
			if maxVal, err := strconv.ParseFloat(matches[2], 64); err == nil {
				result.Max = maxVal
				result.HasMax = true
			}
		}
	}
	return result
}
func getFirstNonEmptyMatch(matches []string) string {
	for _, m := range matches {
		if m != "" {
			return m
		}
	}
	return ""
}
func ApplyMinMaxToConstraints(constraints map[string]any, result MinMaxResult) {
	if result.HasMin {
		constraints["check_min"] = result.Min
	}
	if result.HasMax {
		constraints["check_max"] = result.Max
	}
}
func SanitizeConstraintValue(value string) string {
	value = typeCastPattern.ReplaceAllString(value, "")
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "'\"")
	return value
}
func ExtractColumnNameFromClause(clause string) string {
	clause = strings.ReplaceAll(clause, "`", "")
	clause = strings.ReplaceAll(clause, "[", "")
	clause = strings.ReplaceAll(clause, "]", "")
	clause = strings.TrimSpace(clause)
	clause = strings.TrimPrefix(clause, "(")
	firstWord := columnNamePattern.FindString(clause)
	if firstWord != "" {
		afterWord := strings.TrimPrefix(clause, firstWord)
		afterWord = strings.TrimSpace(afterWord)
		if strings.HasPrefix(afterWord, "(") {
			col := extractColumnFromFunction(afterWord)
			if col != "" {
				return col
			}
		}
	}
	if idx := strings.Index(clause, "::"); idx > 0 {
		clause = clause[:idx]
	}
	clause = strings.TrimSuffix(clause, ")")
	clause = strings.TrimSpace(clause)
	match := columnNamePattern.FindString(clause)
	return match
}
func extractColumnFromFunction(funcArgs string) string {
	funcArgs = strings.TrimLeft(funcArgs, "(")
	if idx := strings.Index(funcArgs, "::"); idx > 0 {
		funcArgs = funcArgs[:idx]
	}
	funcArgs = strings.TrimRight(funcArgs, ")")
	funcArgs = strings.TrimSpace(funcArgs)
	return columnNamePattern.FindString(funcArgs)
}
func IsNullable(constraint map[string]any) bool {
	if constraint == nil {
		return true
	}
	if value, ok := constraint["nullable"].(bool); ok {
		return value
	}
	return true
}
func IsPrimary(constraint map[string]any) bool {
	if constraint == nil {
		return false
	}
	if value, ok := constraint["primary"].(bool); ok {
		return value
	}
	return false
}
func IsAutoIncrement(constraint map[string]any) bool {
	if constraint == nil {
		return false
	}
	if value, ok := constraint["auto_increment"].(bool); ok && value {
		return true
	}
	return false
}
func IsGenerated(constraint map[string]any) bool {
	if constraint == nil {
		return false
	}
	if value, ok := constraint["generated"].(bool); ok && value {
		return true
	}
	return false
}
func HasDefault(constraint map[string]any) bool {
	if constraint == nil {
		return false
	}
	if value, ok := constraint["default"].(string); ok && value != "" {
		return true
	}
	return false
}
func GetDefault(constraint map[string]any) string {
	if constraint == nil {
		return ""
	}
	if value, ok := constraint["default"].(string); ok {
		return value
	}
	return ""
}