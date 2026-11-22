package common
func EscapeFormula(value string) string {
	if len(value) == 0 {
		return value
	}
	firstChar := value[0]
	if firstChar == '=' || firstChar == '+' || firstChar == '-' || firstChar == '@' || firstChar == '\t' || firstChar == '\r' {
		return "'" + value
	}
	return value
}
func FormatCSVHeader(columnName, dataType string) string {
	return columnName
}