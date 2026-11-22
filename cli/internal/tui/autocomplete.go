package tui
import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
	"github.com/charmbracelet/lipgloss"
	"github.com/clidey/whodb/cli/pkg/styles"
)
type suggestionType string
const (
	suggestionTypeKeyword  suggestionType = "keyword"
	suggestionTypeTable    suggestionType = "table"
	suggestionTypeColumn   suggestionType = "column"
	suggestionTypeSchema   suggestionType = "schema"
	suggestionTypeMixed    suggestionType = "mixed"
	suggestionTypeFunction suggestionType = "function"
	suggestionTypeSnippet  suggestionType = "snippet"
	suggestionTypeOperator suggestionType = "operator"
)
type suggestion struct {
	label    string
	typ      suggestionType
	detail   string
	apply    string
	priority int
}
type sqlContext struct {
	contextType    suggestionType
	schema         string
	table          string
	alias          string
	tablesInQuery  []tableInfo
	tokenBeforeDot string
}
type tableInfo struct {
	schema string
	table  string
	alias  string
}
const (
	minSuggestionHeight = 3
	maxSuggestionHeight = 12
)
var (
	identifierTokenPattern = regexp.MustCompile(`[A-Za-z0-9_\.` + "`" + `]+$`)
	fromTrailingPattern    = regexp.MustCompile(`\bFROM\s*$`)
	joinTrailingPattern    = regexp.MustCompile(`\bJOIN\s*$`)
	fromPartialPattern     = regexp.MustCompile(`\bFROM\s+\w*$`)
	joinPartialPattern     = regexp.MustCompile(`\bJOIN\s+\w*$`)
	operatorContextPattern = regexp.MustCompile(`(?i)\b(?:WHERE|AND|OR|ON)\s+[\w.]+\s+$`)
	fromTablePattern       = regexp.MustCompile(`(?i)\bFROM\s+(?:` + "`" + `?(\w+)` + "`" + `?\.)?` + "`" + `?(\w+)` + "`" + `?(?:\s+(?:AS\s+)?` + "`" + `?(\w+)` + "`" + `?)?`)
	joinTablePattern       = regexp.MustCompile(`(?i)\bJOIN\s+(?:` + "`" + `?(\w+)` + "`" + `?\.)?` + "`" + `?(\w+)` + "`" + `?(?:\s+(?:AS\s+)?` + "`" + `?(\w+)` + "`" + `?)?`)
	sqlKeywordSet          = map[string]struct{}{
		"SELECT": {}, "FROM": {}, "WHERE": {}, "AND": {}, "OR": {}, "JOIN": {},
		"LEFT": {}, "RIGHT": {}, "INNER": {}, "OUTER": {}, "ON": {}, "AS": {},
		"BY": {}, "ORDER": {}, "GROUP": {},
	}
	sqlKeywordSuggestions  = buildKeywordSuggestions()
	sqlFunctionSuggestions = []suggestion{
		{label: "COUNT", typ: suggestionTypeFunction, detail: "COUNT(expr)", apply: "COUNT()", priority: 3},
		{label: "SUM", typ: suggestionTypeFunction, detail: "SUM(expr)", apply: "SUM()", priority: 3},
		{label: "AVG", typ: suggestionTypeFunction, detail: "AVG(expr)", apply: "AVG()", priority: 3},
		{label: "MIN", typ: suggestionTypeFunction, detail: "MIN(expr)", apply: "MIN()", priority: 3},
		{label: "MAX", typ: suggestionTypeFunction, detail: "MAX(expr)", apply: "MAX()", priority: 3},
		{label: "COALESCE", typ: suggestionTypeFunction, detail: "COALESCE(expr1, expr2)", apply: "COALESCE()", priority: 3},
		{label: "CAST", typ: suggestionTypeFunction, detail: "CAST(expr AS type)", apply: "CAST()", priority: 3},
		{label: "CONCAT", typ: suggestionTypeFunction, detail: "CONCAT(expr, ...)", apply: "CONCAT()", priority: 3},
	}
	sqlSnippetSuggestions = []suggestion{
		{label: "JOIN ... ON ...", typ: suggestionTypeSnippet, detail: "Snippet: JOIN with ON", apply: "JOIN schema.table alias ON alias.column = other.column", priority: 5},
		{label: "LEFT JOIN ... ON ...", typ: suggestionTypeSnippet, detail: "Snippet: LEFT JOIN with ON", apply: "LEFT JOIN schema.table alias ON alias.column = other.column", priority: 5},
		{label: "WHERE IN (...)", typ: suggestionTypeSnippet, detail: "Snippet: WHERE IN", apply: "WHERE column IN (value1, value2)", priority: 5},
		{label: "GROUP BY ...", typ: suggestionTypeSnippet, detail: "Snippet: GROUP BY", apply: "GROUP BY column1, column2", priority: 5},
		{label: "SELECT DISTINCT ...", typ: suggestionTypeSnippet, detail: "Snippet: SELECT DISTINCT", apply: "SELECT DISTINCT column FROM schema.table", priority: 5},
	}
	sqlOperatorSuggestions = []suggestion{
		{label: "=", typ: suggestionTypeOperator, detail: "Equal", apply: "= ", priority: 2},
		{label: "!=", typ: suggestionTypeOperator, detail: "Not equal", apply: "!= ", priority: 2},
		{label: "<>", typ: suggestionTypeOperator, detail: "Not equal", apply: "<> ", priority: 2},
		{label: ">", typ: suggestionTypeOperator, detail: "Greater than", apply: "> ", priority: 2},
		{label: "<", typ: suggestionTypeOperator, detail: "Less than", apply: "< ", priority: 2},
		{label: ">=", typ: suggestionTypeOperator, detail: "Greater or equal", apply: ">= ", priority: 2},
		{label: "<=", typ: suggestionTypeOperator, detail: "Less or equal", apply: "<= ", priority: 2},
		{label: "LIKE", typ: suggestionTypeOperator, detail: "Pattern match", apply: "LIKE ", priority: 2},
		{label: "NOT LIKE", typ: suggestionTypeOperator, detail: "Negated pattern", apply: "NOT LIKE ", priority: 2},
		{label: "IN", typ: suggestionTypeOperator, detail: "In set", apply: "IN ()", priority: 2},
		{label: "NOT IN", typ: suggestionTypeOperator, detail: "Not in set", apply: "NOT IN ()", priority: 2},
		{label: "BETWEEN", typ: suggestionTypeOperator, detail: "Range", apply: "BETWEEN  AND ", priority: 2},
		{label: "IS NULL", typ: suggestionTypeOperator, detail: "Is null", apply: "IS NULL", priority: 2},
		{label: "IS NOT NULL", typ: suggestionTypeOperator, detail: "Is not null", apply: "IS NOT NULL", priority: 2},
	}
)
const autocompleteDebounceDelay = 100 * time.Millisecond
func (v *EditorView) updateCursorPosition() {
	text := v.textarea.Value()
	row := v.textarea.Line()
	li := v.textarea.LineInfo()
	col := li.ColumnOffset
	lines := strings.Split(text, "\n")
	pos := 0
	for i := 0; i < row && i < len(lines); i++ {
		pos += len(lines[i]) + 1
	}
	if row < len(lines) {
		runes := []rune(lines[row])
		if col > len(runes) {
			col = len(runes)
		}
		pos += len(string(runes[:col]))
	}
	if pos > len(text) {
		pos = len(text)
	}
	v.cursorPos = pos
	v.lastText = text
}
func (v *EditorView) triggerAutocomplete() {
	v.updateCursorPosition()
	text := v.textarea.Value()
	v.updateAutocomplete(text, v.cursorPos)
	if len(v.filteredSuggestions) > 0 {
		v.showSuggestions = true
		v.refreshLayout()
	}
}
func (v *EditorView) computeSuggestionHeight(totalHeight int) int {
	if !v.showSuggestions || len(v.filteredSuggestions) == 0 {
		return 0
	}
	minTextarea := 5
	overhead := 4
	if !v.compact {
		overhead += 4
	}
	available := totalHeight - overhead - minTextarea
	if available < minSuggestionHeight {
		return 0
	}
	height := available
	if height > maxSuggestionHeight {
		height = maxSuggestionHeight
	}
	return height
}
func (v *EditorView) updateAutocomplete(text string, pos int) {
	previouslyShown := v.showSuggestions
	if pos == 0 {
		v.showSuggestions = false
		if previouslyShown {
			v.refreshLayout()
		}
		return
	}
	beforeCursor := text[:pos]
	token := identifierTokenPattern.FindString(beforeCursor)
	ctx := v.parseSQLContext(text, pos)
	v.loadSuggestionsForContext(ctx, text)
	v.filterSuggestions(getLastWord(beforeCursor))
	if len(v.filteredSuggestions) > 0 {
		if token != "" || ctx.contextType != suggestionTypeKeyword {
			v.showSuggestions = true
			if v.selectedSuggestion >= len(v.filteredSuggestions) {
				v.selectedSuggestion = 0
			}
			if !previouslyShown {
				v.refreshLayout()
			}
			return
		}
	}
	v.showSuggestions = false
	if previouslyShown {
		v.refreshLayout()
	}
}
func (v *EditorView) loadSuggestionsForContext(ctx sqlContext, fullText string) {
	v.allSuggestions = v.allSuggestions[:0]
	switch ctx.contextType {
	case suggestionTypeSchema:
		schemas, _ := v.parent.dbManager.GetSchemas()
		for _, s := range schemas {
			v.allSuggestions = append(v.allSuggestions, suggestion{
				label:  s,
				typ:    suggestionTypeSchema,
				detail: "Schema",
				apply:  s,
			})
		}
		v.addTableSuggestions(v.getCurrentSchema())
	case suggestionTypeTable:
		if ctx.schema != "" {
			v.addTableSuggestions(ctx.schema)
		} else {
			v.addTableSuggestions(v.getCurrentSchema())
			schemas, _ := v.parent.dbManager.GetSchemas()
			for _, s := range schemas {
				v.allSuggestions = append(v.allSuggestions, suggestion{
					label:  s,
					typ:    suggestionTypeSchema,
					detail: "Schema",
					apply:  s,
				})
			}
		}
	case suggestionTypeColumn:
		if ctx.tokenBeforeDot != "" {
			v.addQualifiedColumnSuggestions(ctx, fullText)
		} else {
			v.addUnqualifiedColumnSuggestions(ctx)
		}
		v.addFunctionSuggestions()
	case suggestionTypeOperator:
		v.addOperatorSuggestions()
	case suggestionTypeMixed:
		v.addMixedSuggestions(ctx, fullText)
	default:
		v.addTableSuggestions(v.getCurrentSchema())
		v.addKeywordSuggestions()
		v.addFunctionSuggestions()
		v.addSnippetSuggestions()
	}
}
func (v *EditorView) addTableSuggestions(schema string) {
	units, err := v.parent.dbManager.GetStorageUnits(schema)
	if err != nil {
		return
	}
	for _, unit := range units {
		v.allSuggestions = append(v.allSuggestions, suggestion{
			label:    unit.Name,
			typ:      suggestionTypeTable,
			detail:   "Table",
			apply:    unit.Name,
			priority: 1,
		})
	}
}
func (v *EditorView) addQualifiedColumnSuggestions(ctx sqlContext, fullText string) {
	tokenBefore := ctx.tokenBeforeDot
	for _, t := range ctx.tablesInQuery {
		if strings.EqualFold(t.alias, tokenBefore) || strings.EqualFold(t.table, tokenBefore) {
			schema := t.schema
			if schema == "" {
				schema = v.getCurrentSchema()
			}
			columns, err := v.parent.dbManager.GetColumns(schema, t.table)
			if err != nil {
				return
			}
			for _, col := range columns {
				v.allSuggestions = append(v.allSuggestions, suggestion{
					label:  fmt.Sprintf("%s.%s", tokenBefore, col.Name),
					typ:    suggestionTypeColumn,
					detail: col.Type,
					apply:  fmt.Sprintf("%s.%s", tokenBefore, col.Name),
				})
				v.allSuggestions = append(v.allSuggestions, suggestion{
					label:  col.Name,
					typ:    suggestionTypeColumn,
					detail: col.Type,
					apply:  col.Name,
				})
			}
			return
		}
	}
	v.addTableSuggestions(tokenBefore)
}
func (v *EditorView) addUnqualifiedColumnSuggestions(ctx sqlContext) {
	tables := ctx.tablesInQuery
	if len(tables) == 0 {
		return
	}
	if len(tables) == 1 {
		t := tables[0]
		schema := t.schema
		if schema == "" {
			schema = v.getCurrentSchema()
		}
		columns, err := v.parent.dbManager.GetColumns(schema, t.table)
		if err != nil {
			return
		}
		for _, col := range columns {
			v.allSuggestions = append(v.allSuggestions, suggestion{
				label:    col.Name,
				typ:      suggestionTypeColumn,
				detail:   col.Type,
				apply:    col.Name,
				priority: 0,
			})
		}
	} else {
		for i, t := range tables {
			if i >= 3 {
				break
			}
			if t.alias != "" {
				v.allSuggestions = append(v.allSuggestions, suggestion{
					label:  t.alias,
					typ:    suggestionTypeTable,
					detail: fmt.Sprintf("Alias for %s", t.table),
					apply:  t.alias,
				})
			}
			v.allSuggestions = append(v.allSuggestions, suggestion{
				label:  t.table,
				typ:    suggestionTypeTable,
				detail: "Table",
				apply:  t.table,
			})
		}
	}
}
func (v *EditorView) addMixedSuggestions(ctx sqlContext, fullText string) {
	if len(ctx.tablesInQuery) == 0 {
		v.addTableSuggestions(v.getCurrentSchema())
	}
	for _, t := range ctx.tablesInQuery {
		if t.alias != "" {
			v.allSuggestions = append(v.allSuggestions, suggestion{
				label:  t.alias,
				typ:    suggestionTypeTable,
				detail: fmt.Sprintf("Alias for %s", t.table),
				apply:  t.alias,
			})
		}
		v.allSuggestions = append(v.allSuggestions, suggestion{
			label:  t.table,
			typ:    suggestionTypeTable,
			detail: "Table",
			apply:  t.table,
		})
	}
	for i, t := range ctx.tablesInQuery {
		if i >= 3 {
			break
		}
		schema := t.schema
		if schema == "" {
			schema = v.getCurrentSchema()
		}
		columns, err := v.parent.dbManager.GetColumns(schema, t.table)
		if err != nil {
			continue
		}
		for _, col := range columns {
			if t.alias != "" {
				v.allSuggestions = append(v.allSuggestions, suggestion{
					label:  fmt.Sprintf("%s.%s", t.alias, col.Name),
					typ:    suggestionTypeColumn,
					detail: col.Type,
					apply:  fmt.Sprintf("%s.%s", t.alias, col.Name),
				})
			}
			v.allSuggestions = append(v.allSuggestions, suggestion{
				label:  fmt.Sprintf("%s.%s", t.table, col.Name),
				typ:    suggestionTypeColumn,
				detail: col.Type,
				apply:  fmt.Sprintf("%s.%s", t.table, col.Name),
			})
			v.allSuggestions = append(v.allSuggestions, suggestion{
				label:    col.Name,
				typ:      suggestionTypeColumn,
				detail:   col.Type,
				apply:    col.Name,
				priority: 0,
			})
		}
	}
	v.addKeywordSuggestions()
	v.addFunctionSuggestions()
	v.addSnippetSuggestions()
}
func (v *EditorView) addKeywordSuggestions() {
	v.allSuggestions = append(v.allSuggestions, sqlKeywordSuggestions...)
}
func (v *EditorView) addFunctionSuggestions() {
	v.allSuggestions = append(v.allSuggestions, sqlFunctionSuggestions...)
}
func (v *EditorView) addSnippetSuggestions() {
	v.allSuggestions = append(v.allSuggestions, sqlSnippetSuggestions...)
}
func (v *EditorView) addOperatorSuggestions() {
	v.allSuggestions = append(v.allSuggestions, sqlOperatorSuggestions...)
}
func (v *EditorView) getCurrentSchema() string {
	if v.parent.browserView.currentSchema != "" {
		return v.parent.browserView.currentSchema
	}
	schemas, err := v.parent.dbManager.GetSchemas()
	if err != nil || len(schemas) == 0 {
		return ""
	}
	return selectBestSchema(schemas)
}
func (v *EditorView) filterSuggestions(prefix string) {
	if prefix == "" {
		v.filteredSuggestions = append(v.filteredSuggestions[:0], v.allSuggestions...)
	} else {
		v.filteredSuggestions = v.filteredSuggestions[:0]
		lowerPrefix := strings.ToLower(prefix)
		for _, sug := range v.allSuggestions {
			if strings.HasPrefix(strings.ToLower(sug.label), lowerPrefix) {
				v.filteredSuggestions = append(v.filteredSuggestions, sug)
				continue
			}
			if lastDot := strings.LastIndex(sug.label, "."); lastDot >= 0 &&
				strings.HasPrefix(strings.ToLower(sug.label[lastDot+1:]), lowerPrefix) {
				v.filteredSuggestions = append(v.filteredSuggestions, sug)
			}
		}
	}
	for i := range v.filteredSuggestions {
		v.filteredSuggestions[i].priority = suggestionPriority(v.filteredSuggestions[i].typ)
	}
	sort.SliceStable(v.filteredSuggestions, func(i, j int) bool {
		return v.filteredSuggestions[i].priority < v.filteredSuggestions[j].priority
	})
}
func (v *EditorView) renderAutocompletePanel() string {
	var panel strings.Builder
	count := len(v.filteredSuggestions)
	header := styles.RenderMuted(fmt.Sprintf("Suggestions (%d)", count))
	panel.WriteString(header)
	panel.WriteString("\n\n")
	maxDisplay := 8
	if len(v.filteredSuggestions) < maxDisplay {
		maxDisplay = len(v.filteredSuggestions)
	}
	startIdx := 0
	if v.selectedSuggestion >= maxDisplay {
		startIdx = v.selectedSuggestion - maxDisplay + 1
	}
	for i := startIdx; i < startIdx+maxDisplay && i < len(v.filteredSuggestions); i++ {
		sug := v.filteredSuggestions[i]
		line := fmt.Sprintf("%s (%s)", sug.label, sug.detail)
		if i == v.selectedSuggestion {
			panel.WriteString(styles.ActiveListItemStyle.Render("> " + line))
		} else {
			panel.WriteString(styles.RenderMuted("  " + line))
		}
		panel.WriteString("\n")
	}
	if len(v.filteredSuggestions) > startIdx+maxDisplay {
		panel.WriteString(styles.RenderMuted(fmt.Sprintf("  ... and %d more", len(v.filteredSuggestions)-startIdx-maxDisplay)))
		panel.WriteString("\n")
	}
	panel.WriteString("\n")
	controlPanel := styles.HelpStyle.Render(
		styles.RenderKey("↑↓") + " " + styles.RenderMuted("navigate") + " " +
			styles.RenderKey("tab") + " " + styles.RenderMuted("cycle") + " " +
			styles.RenderKey("↵") + " " + styles.RenderMuted("accept") + " " +
			styles.RenderKey("esc") + " " + styles.RenderMuted("dismiss"),
	)
	panel.WriteString(controlPanel)
	return styles.BoxStyle.Render(panel.String())
}
func (v *EditorView) renderSuggestionArea() string {
	content := ""
	if v.showSuggestions && len(v.filteredSuggestions) > 0 {
		content = v.renderAutocompletePanel()
	}
	return lipgloss.NewStyle().
		Height(v.suggestionHeight).
		MaxHeight(v.suggestionHeight).
		Render(content)
}
func (v *EditorView) acceptSuggestion() {
	if v.selectedSuggestion >= 0 && v.selectedSuggestion < len(v.filteredSuggestions) {
		sug := v.filteredSuggestions[v.selectedSuggestion]
		text := v.textarea.Value()
		beforeCursor := text[:v.cursorPos]
		afterCursor := text[v.cursorPos:]
		tokenMatch := identifierTokenPattern.FindString(beforeCursor)
		var newText string
		var newCursorPos int
		if tokenMatch != "" {
			if strings.Contains(tokenMatch, ".") && !strings.Contains(sug.apply, ".") {
				lastDot := strings.LastIndex(tokenMatch, ".")
				prefixToKeep := tokenMatch[:lastDot+1]
				startPos := v.cursorPos - len(tokenMatch)
				newText = text[:startPos] + prefixToKeep + sug.apply + afterCursor
				newCursorPos = startPos + len(prefixToKeep) + len(sug.apply)
			} else {
				startPos := v.cursorPos - len(tokenMatch)
				newText = text[:startPos] + sug.apply + afterCursor
				newCursorPos = startPos + len(sug.apply)
			}
		} else {
			newText = beforeCursor + sug.apply + afterCursor
			newCursorPos = v.cursorPos + len(sug.apply)
		}
		v.textarea.SetValue(newText)
		v.cursorPos = newCursorPos
		v.lastText = newText
		v.showSuggestions = false
		v.selectedSuggestion = 0
		v.refreshLayout()
	}
}
func (v *EditorView) parseSQLContext(text string, pos int) sqlContext {
	beforeCursor := text[:pos]
	beforeUpper := strings.ToUpper(beforeCursor)
	tablesInQuery := v.extractTablesAndAliases(text)
	tokenMatch := identifierTokenPattern.FindString(beforeCursor)
	if strings.Contains(tokenMatch, ".") {
		parts := strings.Split(tokenMatch, ".")
		if len(parts) >= 2 {
			tokenBeforeDot := strings.Trim(parts[0], "`")
			for _, t := range tablesInQuery {
				if strings.EqualFold(t.alias, tokenBeforeDot) || strings.EqualFold(t.table, tokenBeforeDot) {
					return sqlContext{
						contextType:    suggestionTypeColumn,
						schema:         t.schema,
						table:          t.table,
						alias:          t.alias,
						tablesInQuery:  tablesInQuery,
						tokenBeforeDot: tokenBeforeDot,
					}
				}
			}
			return sqlContext{
				contextType:    suggestionTypeTable,
				schema:         tokenBeforeDot,
				tablesInQuery:  tablesInQuery,
				tokenBeforeDot: tokenBeforeDot,
			}
		}
	}
	if fromTrailingPattern.MatchString(beforeUpper) ||
		joinTrailingPattern.MatchString(beforeUpper) {
		return sqlContext{
			contextType:   suggestionTypeSchema,
			tablesInQuery: tablesInQuery,
		}
	}
	if fromPartialPattern.MatchString(beforeUpper) ||
		joinPartialPattern.MatchString(beforeUpper) {
		return sqlContext{
			contextType:   suggestionTypeTable,
			tablesInQuery: tablesInQuery,
		}
	}
	lastSelect := strings.LastIndex(beforeUpper, "SELECT")
	lastFrom := strings.LastIndex(beforeUpper, "FROM")
	lastWhere := strings.LastIndex(beforeUpper, "WHERE")
	lastOn := strings.LastIndex(beforeUpper, " ON ")
	if ((lastWhere > -1 && lastWhere > lastFrom) || (lastOn > -1)) && len(tablesInQuery) > 0 {
		if operatorContextPattern.MatchString(beforeCursor) {
			return sqlContext{
				contextType:   suggestionTypeOperator,
				tablesInQuery: tablesInQuery,
			}
		}
		return sqlContext{
			contextType:   suggestionTypeMixed,
			tablesInQuery: tablesInQuery,
		}
	}
	fullTextUpper := strings.ToUpper(text)
	hasFromInQuery := strings.Contains(fullTextUpper, "FROM")
	if lastSelect > -1 {
		if len(tablesInQuery) > 0 && ((lastFrom > lastSelect) || (lastFrom == -1 && hasFromInQuery)) {
			return sqlContext{
				contextType:   suggestionTypeColumn,
				tablesInQuery: tablesInQuery,
			}
		}
		if lastFrom == -1 || lastFrom < lastSelect {
			return sqlContext{
				contextType:   suggestionTypeMixed,
				tablesInQuery: tablesInQuery,
			}
		}
	}
	if tokenMatch != "" && !strings.Contains(tokenMatch, ".") && len(tablesInQuery) > 0 {
		if len(tablesInQuery) == 1 {
			return sqlContext{
				contextType:   suggestionTypeColumn,
				schema:        tablesInQuery[0].schema,
				table:         tablesInQuery[0].table,
				alias:         tablesInQuery[0].alias,
				tablesInQuery: tablesInQuery,
			}
		}
		return sqlContext{
			contextType:   suggestionTypeMixed,
			tablesInQuery: tablesInQuery,
		}
	}
	return sqlContext{
		contextType:   suggestionTypeKeyword,
		tablesInQuery: tablesInQuery,
	}
}
func (v *EditorView) extractTablesAndAliases(text string) []tableInfo {
	tables := []tableInfo{}
	for _, pattern := range []*regexp.Regexp{fromTablePattern, joinTablePattern} {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				schema := strings.Trim(match[1], "`")
				table := strings.Trim(match[2], "`")
				alias := ""
				if len(match) >= 4 {
					alias = strings.Trim(match[3], "`")
				}
				if !v.isKeyword(table) {
					tables = append(tables, tableInfo{
						schema: schema,
						table:  table,
						alias:  alias,
					})
				}
			}
		}
	}
	return tables
}
func (v *EditorView) isKeyword(word string) bool {
	_, ok := sqlKeywordSet[strings.ToUpper(word)]
	return ok
}
func getLastWord(text string) string {
	if len(text) == 0 {
		return ""
	}
	lastChar, _ := utf8.DecodeLastRuneInString(text)
	if lastChar == ' ' || lastChar == '\n' || lastChar == '\t' || lastChar == ',' || lastChar == '(' || lastChar == ')' {
		return ""
	}
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t' || r == ',' || r == '(' || r == ')'
	})
	if len(words) == 0 {
		return ""
	}
	lastWord := words[len(words)-1]
	if strings.Contains(lastWord, ".") {
		parts := strings.Split(lastWord, ".")
		if len(parts) > 1 {
			return strings.Trim(parts[len(parts)-1], "`")
		}
	}
	return strings.Trim(lastWord, "`")
}
func suggestionPriority(typ suggestionType) int {
	switch typ {
	case suggestionTypeColumn:
		return 0
	case suggestionTypeTable, suggestionTypeSchema:
		return 1
	case suggestionTypeOperator:
		return 2
	case suggestionTypeFunction:
		return 3
	case suggestionTypeKeyword:
		return 4
	case suggestionTypeSnippet:
		return 5
	default:
		return 6
	}
}
func buildKeywordSuggestions() []suggestion {
	keywords := []string{
		"SELECT", "FROM", "WHERE", "JOIN", "INNER", "LEFT", "RIGHT", "OUTER", "FULL",
		"ON", "AND", "OR", "NOT", "NULL", "TRUE", "FALSE", "ORDER", "BY", "GROUP",
		"HAVING", "LIMIT", "OFFSET", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP",
		"ALTER", "TABLE", "INDEX", "VIEW", "DATABASE", "SCHEMA", "INTO", "VALUES",
		"SET", "AS", "DISTINCT", "COUNT", "SUM", "AVG", "MAX", "MIN", "CASE", "WHEN",
		"THEN", "ELSE", "END", "IF", "EXISTS", "LIKE", "IN", "BETWEEN", "IS", "ASC",
		"DESC", "UNION", "ALL", "ANY", "SOME", "WITH", "RECURSIVE", "CASCADE",
		"CONSTRAINT", "PRIMARY", "KEY", "FOREIGN", "REFERENCES", "UNIQUE", "CHECK",
		"DEFAULT", "TRUNCATE", "EXPLAIN", "ANALYZE", "GRANT", "REVOKE", "COMMIT",
		"ROLLBACK", "TRANSACTION", "BEGIN", "START",
	}
	suggestions := make([]suggestion, 0, len(keywords))
	for _, kw := range keywords {
		suggestions = append(suggestions, suggestion{
			label:    kw,
			typ:      suggestionTypeKeyword,
			detail:   "SQL Keyword",
			apply:    kw,
			priority: 4,
		})
	}
	return suggestions
}