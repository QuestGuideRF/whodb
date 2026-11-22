package tui
import "testing"
func TestParseSQLContext_OperatorAfterWhereColumn(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel: %v", m.err)
	}
	ev := m.editorView
	tests := []struct {
		name     string
		text     string
		expected suggestionType
	}{
		{
			name:     "after WHERE column space",
			text:     "SELECT * FROM users WHERE name ",
			expected: suggestionTypeOperator,
		},
		{
			name:     "after AND column space",
			text:     "SELECT * FROM users WHERE id = 1 AND name ",
			expected: suggestionTypeOperator,
		},
		{
			name:     "typing column name in WHERE",
			text:     "SELECT * FROM users WHERE na",
			expected: suggestionTypeMixed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ev.parseSQLContext(tt.text, len(tt.text))
			if ctx.contextType != tt.expected {
				t.Errorf("got context %q, want %q", ctx.contextType, tt.expected)
			}
		})
	}
}
func TestFilterSuggestions_SortedByPriority(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel: %v", m.err)
	}
	ev := m.editorView
	ev.allSuggestions = []suggestion{
		{label: "SELECT", typ: suggestionTypeKeyword},
		{label: "name", typ: suggestionTypeColumn},
		{label: "users", typ: suggestionTypeTable},
		{label: "COUNT", typ: suggestionTypeFunction},
		{label: "SELECT DISTINCT ...", typ: suggestionTypeSnippet},
	}
	ev.filterSuggestions("")
	if len(ev.filteredSuggestions) != 5 {
		t.Fatalf("expected 5 suggestions, got %d", len(ev.filteredSuggestions))
	}
	if ev.filteredSuggestions[0].typ != suggestionTypeColumn {
		t.Errorf("first suggestion should be column, got %s", ev.filteredSuggestions[0].typ)
	}
	if ev.filteredSuggestions[1].typ != suggestionTypeTable {
		t.Errorf("second suggestion should be table, got %s", ev.filteredSuggestions[1].typ)
	}
	keywordIdx, functionIdx := -1, -1
	for i, s := range ev.filteredSuggestions {
		if s.typ == suggestionTypeKeyword {
			keywordIdx = i
		}
		if s.typ == suggestionTypeFunction {
			functionIdx = i
		}
	}
	if keywordIdx < functionIdx {
		t.Error("keywords should rank below functions")
	}
}
func TestAddOperatorSuggestions(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel: %v", m.err)
	}
	ev := m.editorView
	ev.allSuggestions = nil
	ev.addOperatorSuggestions()
	if len(ev.allSuggestions) == 0 {
		t.Fatal("should have operator suggestions")
	}
	found := map[string]bool{}
	for _, s := range ev.allSuggestions {
		found[s.label] = true
		if s.typ != suggestionTypeOperator {
			t.Errorf("operator suggestion %q has wrong type %s", s.label, s.typ)
		}
	}
	for _, op := range []string{"=", "!=", "LIKE", "IN", "IS NULL", "BETWEEN"} {
		if !found[op] {
			t.Errorf("missing operator suggestion: %s", op)
		}
	}
}