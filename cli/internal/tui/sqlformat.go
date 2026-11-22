package tui
import (
	"strings"
	"unicode"
)
func formatSQL(sql string) string {
	tokens := tokenizeSQL(sql)
	if len(tokens) == 0 {
		return sql
	}
	newlineKeywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true,
		"ORDER": true, "GROUP": true, "HAVING": true,
		"LIMIT": true, "OFFSET": true, "UNION": true,
		"INSERT": true, "UPDATE": true, "DELETE": true,
		"SET": true, "VALUES": true, "INTO": true,
		"CREATE": true, "ALTER": true, "DROP": true,
		"WITH": true, "RETURNING": true, "EXPLAIN": true,
	}
	indentKeywords := map[string]bool{
		"LEFT": true, "RIGHT": true, "INNER": true,
		"OUTER": true, "CROSS": true, "FULL": true,
		"JOIN": true, "ON": true, "AND": true, "OR": true,
	}
	var b strings.Builder
	prevWasNewline := false
	for i, tok := range tokens {
		upper := strings.ToUpper(tok.text)
		if tok.kind == tokenString || tok.kind == tokenQuoted {
			b.WriteString(tok.text)
			prevWasNewline = false
			continue
		}
		if tok.kind == tokenWord {
			if newlineKeywords[upper] {
				if i > 0 && !prevWasNewline {
					b.WriteString("\n")
				}
				b.WriteString(upper)
				prevWasNewline = false
				continue
			}
			if indentKeywords[upper] {
				prevUpper := ""
				for p := i - 1; p >= 0; p-- {
					if tokens[p].kind != tokenWhitespace {
						prevUpper = strings.ToUpper(tokens[p].text)
						break
					}
				}
				joinModifiers := map[string]bool{"LEFT": true, "RIGHT": true, "INNER": true, "OUTER": true, "CROSS": true, "FULL": true}
				if joinModifiers[upper] {
					if !prevWasNewline {
						b.WriteString("\n  ")
					}
					b.WriteString(upper)
				} else if upper == "JOIN" {
					if joinModifiers[prevUpper] {
						b.WriteString(" ")
					} else if !prevWasNewline {
						b.WriteString("\n  ")
					}
					b.WriteString(upper)
				} else if upper == "AND" || upper == "OR" {
					b.WriteString("\n  ")
					b.WriteString(upper)
				} else if upper == "ON" {
					b.WriteString(" ")
					b.WriteString(upper)
				} else {
					if !prevWasNewline {
						b.WriteString("\n  ")
					}
					b.WriteString(upper)
				}
				prevWasNewline = false
				continue
			}
			if isSQLKeyword(upper) {
				b.WriteString(upper)
			} else {
				b.WriteString(tok.text)
			}
			prevWasNewline = false
			continue
		}
		if tok.kind == tokenWhitespace {
			if i+1 < len(tokens) && tokens[i+1].kind == tokenWord {
				nextUpper := strings.ToUpper(tokens[i+1].text)
				if newlineKeywords[nextUpper] || indentKeywords[nextUpper] {
					continue
				}
			}
			if !prevWasNewline {
				b.WriteString(" ")
			}
			continue
		}
		b.WriteString(tok.text)
		prevWasNewline = false
	}
	return strings.TrimSpace(b.String())
}
type tokenKind int
const (
	tokenWord tokenKind = iota
	tokenString
	tokenQuoted
	tokenWhitespace
	tokenOther
)
type sqlToken struct {
	text string
	kind tokenKind
}
func tokenizeSQL(sql string) []sqlToken {
	var tokens []sqlToken
	runes := []rune(sql)
	i := 0
	for i < len(runes) {
		ch := runes[i]
		if ch == '\'' {
			end := i + 1
			for end < len(runes) {
				if runes[end] == '\'' {
					if end+1 < len(runes) && runes[end+1] == '\'' {
						end += 2
						continue
					}
					end++
					break
				}
				end++
			}
			tokens = append(tokens, sqlToken{string(runes[i:end]), tokenString})
			i = end
			continue
		}
		if ch == '"' {
			end := i + 1
			for end < len(runes) && runes[end] != '"' {
				end++
			}
			if end < len(runes) {
				end++
			}
			tokens = append(tokens, sqlToken{string(runes[i:end]), tokenQuoted})
			i = end
			continue
		}
		if ch == '`' {
			end := i + 1
			for end < len(runes) && runes[end] != '`' {
				end++
			}
			if end < len(runes) {
				end++
			}
			tokens = append(tokens, sqlToken{string(runes[i:end]), tokenQuoted})
			i = end
			continue
		}
		if unicode.IsLetter(ch) || ch == '_' {
			end := i + 1
			for end < len(runes) && (unicode.IsLetter(runes[end]) || unicode.IsDigit(runes[end]) || runes[end] == '_') {
				end++
			}
			tokens = append(tokens, sqlToken{string(runes[i:end]), tokenWord})
			i = end
			continue
		}
		if unicode.IsDigit(ch) {
			end := i + 1
			for end < len(runes) && (unicode.IsDigit(runes[end]) || runes[end] == '.') {
				end++
			}
			tokens = append(tokens, sqlToken{string(runes[i:end]), tokenOther})
			i = end
			continue
		}
		if unicode.IsSpace(ch) {
			end := i + 1
			for end < len(runes) && unicode.IsSpace(runes[end]) {
				end++
			}
			tokens = append(tokens, sqlToken{string(runes[i:end]), tokenWhitespace})
			i = end
			continue
		}
		tokens = append(tokens, sqlToken{string(ch), tokenOther})
		i++
	}
	return tokens
}
func isSQLKeyword(word string) bool {
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "JOIN": true,
		"INNER": true, "LEFT": true, "RIGHT": true, "OUTER": true,
		"FULL": true, "CROSS": true, "ON": true, "AND": true, "OR": true,
		"NOT": true, "NULL": true, "TRUE": true, "FALSE": true,
		"ORDER": true, "BY": true, "GROUP": true, "HAVING": true,
		"LIMIT": true, "OFFSET": true, "INSERT": true, "UPDATE": true,
		"DELETE": true, "CREATE": true, "DROP": true, "ALTER": true,
		"TABLE": true, "INDEX": true, "VIEW": true, "DATABASE": true,
		"SCHEMA": true, "INTO": true, "VALUES": true, "SET": true,
		"AS": true, "DISTINCT": true, "CASE": true, "WHEN": true,
		"THEN": true, "ELSE": true, "END": true, "IF": true,
		"EXISTS": true, "LIKE": true, "IN": true, "BETWEEN": true,
		"IS": true, "ASC": true, "DESC": true, "UNION": true,
		"ALL": true, "ANY": true, "SOME": true, "WITH": true,
		"CASCADE": true, "CONSTRAINT": true, "PRIMARY": true,
		"KEY": true, "FOREIGN": true, "REFERENCES": true,
		"UNIQUE": true, "CHECK": true, "DEFAULT": true,
		"TRUNCATE": true, "EXPLAIN": true, "ANALYZE": true,
		"COUNT": true, "SUM": true, "AVG": true, "MAX": true, "MIN": true,
		"CAST": true, "COALESCE": true, "CONCAT": true,
		"RETURNING": true, "RECURSIVE": true, "BEGIN": true,
		"COMMIT": true, "ROLLBACK": true, "TRANSACTION": true,
		"GRANT": true, "REVOKE": true, "USING": true,
	}
	return keywords[word]
}