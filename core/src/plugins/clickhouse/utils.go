package clickhouse
import "strings"
func (p *ClickHousePlugin) GetPrimaryKeyColQuery() string {
	return `
		SELECT name
		FROM system.columns
		WHERE database = ? AND table = ? AND is_in_primary_key = 1
	`
}
func isMultiStatement(query string) bool {
	var b strings.Builder
	inBlock := false
	for i := 0; i < len(query); i++ {
		if inBlock {
			if i+1 < len(query) && query[i] == '*' && query[i+1] == '/' {
				inBlock = false
				i++
			}
			continue
		}
		if i+1 < len(query) && query[i] == '/' && query[i+1] == '*' {
			inBlock = true
			i++
			continue
		}
		if i+1 < len(query) && query[i] == '-' && query[i+1] == '-' {
			for i < len(query) && query[i] != '\n' {
				i++
			}
			continue
		}
		if query[i] == '#' {
			for i < len(query) && query[i] != '\n' {
				i++
			}
			continue
		}
		b.WriteByte(query[i])
	}
	return strings.Count(strings.TrimSpace(b.String()), ";") > 1
}