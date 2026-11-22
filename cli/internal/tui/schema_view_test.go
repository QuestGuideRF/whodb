package tui
import (
	"errors"
	"strings"
	"testing"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clidey/whodb/core/src/engine"
)
func setupSchemaViewTest(t *testing.T) (*SchemaView, func()) {
	t.Helper()
	setupTestEnv(t)
	parent := NewMainModel()
	if parent.err != nil {
		t.Fatalf("Failed to create MainModel: %v", parent.err)
	}
	cleanup := func() {}
	return parent.schemaView, cleanup
}
func TestNewSchemaView(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	if v == nil {
		t.Fatal("NewSchemaView returned nil")
	}
	if !v.loading {
		t.Error("Expected loading to be true initially")
	}
	if v.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", v.selectedIndex)
	}
	if v.expandedTables == nil {
		t.Error("Expected expandedTables map to be initialized")
	}
	if v.filtering {
		t.Error("Expected filtering to be false initially")
	}
	if v.width != 80 {
		t.Errorf("Expected default width 80, got %d", v.width)
	}
	if v.height != 20 {
		t.Errorf("Expected default height 20, got %d", v.height)
	}
}
func TestSchemaView_SchemaLoadedMsg(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	tables := []tableWithColumns{
		{
			StorageUnit: engine.StorageUnit{Name: "users"},
			Columns:     []engine.Column{{Name: "id", Type: "integer"}},
		},
		{
			StorageUnit: engine.StorageUnit{Name: "orders"},
			Columns:     []engine.Column{{Name: "id", Type: "integer"}},
		},
	}
	msg := schemaLoadedMsg{tables: tables, err: nil}
	v, _ = v.Update(msg)
	if v.loading {
		t.Error("Expected loading to be false after schema loaded")
	}
	if len(v.tables) != 2 {
		t.Errorf("Expected 2 tables, got %d", len(v.tables))
	}
	if len(v.filteredTables) != 2 {
		t.Errorf("Expected 2 filtered tables, got %d", len(v.filteredTables))
	}
	if v.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex reset to 0, got %d", v.selectedIndex)
	}
}
func TestSchemaView_SchemaLoadedMsg_Error(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	msg := schemaLoadedMsg{tables: nil, err: errors.New("connection failed")}
	v, _ = v.Update(msg)
	if v.loading {
		t.Error("Expected loading to be false after error")
	}
	if v.err == nil {
		t.Error("Expected error to be set")
	}
}
func TestSchemaView_Navigation_UpDown(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "users"}},
		{StorageUnit: engine.StorageUnit{Name: "orders"}},
		{StorageUnit: engine.StorageUnit{Name: "products"}},
	}
	v.filteredTables = v.tables
	v.selectedIndex = 0
	v.height = 50
	msg := tea.KeyMsg{Type: tea.KeyDown}
	v, _ = v.Update(msg)
	if v.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after down, got %d", v.selectedIndex)
	}
	v, _ = v.Update(msg)
	if v.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2 after second down, got %d", v.selectedIndex)
	}
	v, _ = v.Update(msg)
	if v.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex to stay 2 at end, got %d", v.selectedIndex)
	}
	msg = tea.KeyMsg{Type: tea.KeyUp}
	v, _ = v.Update(msg)
	if v.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after up, got %d", v.selectedIndex)
	}
}
func TestSchemaView_Navigation_VimKeys(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "users"}},
		{StorageUnit: engine.StorageUnit{Name: "orders"}},
	}
	v.filteredTables = v.tables
	v.selectedIndex = 0
	v.height = 50
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	v, _ = v.Update(msg)
	if v.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after 'j', got %d", v.selectedIndex)
	}
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	v, _ = v.Update(msg)
	if v.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after 'k', got %d", v.selectedIndex)
	}
}
func TestSchemaView_Navigation_AtTop(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "users"}},
	}
	v.filteredTables = v.tables
	v.selectedIndex = 0
	msg := tea.KeyMsg{Type: tea.KeyUp}
	v, _ = v.Update(msg)
	if v.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex to stay 0 at top, got %d", v.selectedIndex)
	}
}
func TestSchemaView_ExpandCollapse_Enter(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{
			StorageUnit: engine.StorageUnit{Name: "users"},
			Columns:     []engine.Column{{Name: "id", Type: "integer"}},
		},
	}
	v.filteredTables = v.tables
	v.selectedIndex = 0
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	v, _ = v.Update(msg)
	if !v.expandedTables["users"] {
		t.Error("Expected table 'users' to be expanded after Enter")
	}
	v, _ = v.Update(msg)
	if v.expandedTables["users"] {
		t.Error("Expected table 'users' to be collapsed after second Enter")
	}
}
func TestSchemaView_ExpandCollapse_Space(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{
			StorageUnit: engine.StorageUnit{Name: "orders"},
			Columns:     []engine.Column{{Name: "id", Type: "integer"}},
		},
	}
	v.filteredTables = v.tables
	v.selectedIndex = 0
	msg := tea.KeyMsg{Type: tea.KeySpace}
	v, _ = v.Update(msg)
	if !v.expandedTables["orders"] {
		t.Error("Expected table 'orders' to be expanded after Space")
	}
}
func TestSchemaView_Filter_Enter(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.filtering = false
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	v, _ = v.Update(msg)
	if !v.filtering {
		t.Error("Expected filtering to be true after '/'")
	}
}
func TestSchemaView_Filter_EnterWithF(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.filtering = false
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	v, _ = v.Update(msg)
	if !v.filtering {
		t.Error("Expected filtering to be true after 'f'")
	}
}
func TestSchemaView_Filter_Cancel(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "users"}},
		{StorageUnit: engine.StorageUnit{Name: "orders"}},
	}
	v.filteredTables = v.tables
	v.filtering = true
	v.filterInput.SetValue("usr")
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	v, _ = v.Update(msg)
	if v.filtering {
		t.Error("Expected filtering to be false after Esc")
	}
	if v.filterInput.Value() != "" {
		t.Error("Expected filter value to be cleared after Esc")
	}
	if len(v.filteredTables) != 2 {
		t.Errorf("Expected all tables after cancel, got %d", len(v.filteredTables))
	}
}
func TestSchemaView_Filter_Apply(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "users"}},
		{StorageUnit: engine.StorageUnit{Name: "orders"}},
	}
	v.filteredTables = v.tables
	v.filtering = true
	v.filterInput.SetValue("user")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	v, _ = v.Update(msg)
	if v.filtering {
		t.Error("Expected filtering to be false after Enter")
	}
}
func TestSchemaView_ApplyFilter(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "users"}},
		{StorageUnit: engine.StorageUnit{Name: "orders"}},
		{StorageUnit: engine.StorageUnit{Name: "user_roles"}},
	}
	v.filterInput.SetValue("user")
	v.applyFilter()
	if len(v.filteredTables) != 2 {
		t.Errorf("Expected 2 filtered tables (users, user_roles), got %d", len(v.filteredTables))
	}
	v.filterInput.SetValue("")
	v.applyFilter()
	if len(v.filteredTables) != 3 {
		t.Errorf("Expected 3 tables after clearing filter, got %d", len(v.filteredTables))
	}
}
func TestSchemaView_ApplyFilter_CaseInsensitive(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "Users"}},
		{StorageUnit: engine.StorageUnit{Name: "ORDERS"}},
	}
	v.filterInput.SetValue("user")
	v.applyFilter()
	if len(v.filteredTables) != 1 {
		t.Errorf("Expected 1 filtered table, got %d", len(v.filteredTables))
	}
	if v.filteredTables[0].StorageUnit.Name != "Users" {
		t.Error("Expected 'Users' table to match")
	}
}
func TestSchemaView_ApplyFilter_ResetsSelection(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "users"}},
		{StorageUnit: engine.StorageUnit{Name: "orders"}},
	}
	v.filteredTables = v.tables
	v.selectedIndex = 1
	v.filterInput.SetValue("users")
	v.applyFilter()
	if v.selectedIndex >= len(v.filteredTables) {
		t.Error("Expected selectedIndex to be within bounds after filter")
	}
}
func TestSchemaView_Escape(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.filtering = false
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	v, _ = v.Update(msg)
	if v.parent.mode != ViewBrowser {
		t.Errorf("Expected mode ViewBrowser after Esc, got %v", v.parent.mode)
	}
}
func TestSchemaView_WindowSizeMsg(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	v, _ = v.Update(msg)
	if v.width != 120 {
		t.Errorf("Expected width 120, got %d", v.width)
	}
	if v.height != 40 {
		t.Errorf("Expected height 40, got %d", v.height)
	}
}
func TestSchemaView_MouseScroll(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = make([]tableWithColumns, 30)
	for i := 0; i < 30; i++ {
		v.tables[i] = tableWithColumns{
			StorageUnit: engine.StorageUnit{Name: string(rune('a' + i))},
		}
	}
	v.filteredTables = v.tables
	v.height = 20
	v.scrollOffset = 0
	msg := tea.MouseMsg{Type: tea.MouseWheelDown}
	v, _ = v.Update(msg)
	if v.scrollOffset <= 0 {
		t.Error("Expected scrollOffset to increase after wheel down")
	}
	initialOffset := v.scrollOffset
	msg = tea.MouseMsg{Type: tea.MouseWheelUp}
	v, _ = v.Update(msg)
	if v.scrollOffset >= initialOffset {
		t.Error("Expected scrollOffset to decrease after wheel up")
	}
}
func TestSchemaView_MouseScrollUp_AtTop(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.scrollOffset = 0
	msg := tea.MouseMsg{Type: tea.MouseWheelUp}
	v, _ = v.Update(msg)
	if v.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to stay 0 at top, got %d", v.scrollOffset)
	}
}
func TestSchemaView_View_Loading(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = true
	view := v.View()
	if !strings.Contains(view, "Loading schema...") {
		t.Error("Expected 'Loading schema...' when loading")
	}
}
func TestSchemaView_View_Error(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.err = errors.New("connection failed")
	view := v.View()
	if !strings.Contains(view, "connection failed") {
		t.Error("Expected error message in view")
	}
	if !strings.Contains(view, "Press 'r' to retry") {
		t.Error("Expected retry hint in view")
	}
}
func TestSchemaView_View_NoTables(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{}
	v.filteredTables = []tableWithColumns{}
	view := v.View()
	if !strings.Contains(view, "No tables found") {
		t.Error("Expected 'No tables found' when empty")
	}
}
func TestSchemaView_View_WithTables(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{
			StorageUnit: engine.StorageUnit{Name: "users"},
			Columns:     []engine.Column{{Name: "id", Type: "integer"}},
		},
	}
	v.filteredTables = v.tables
	v.height = 30
	view := v.View()
	if !strings.Contains(view, "Database Schema") {
		t.Error("Expected 'Database Schema' title")
	}
	if !strings.Contains(view, "users") {
		t.Error("Expected table name 'users' in view")
	}
}
func TestSchemaView_View_FilterActive(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.filtering = true
	view := v.View()
	if !strings.Contains(view, "Filter:") {
		t.Error("Expected 'Filter:' when filtering")
	}
	if !strings.Contains(view, "cancel") {
		t.Error("Expected 'cancel' help when filtering")
	}
	if !strings.Contains(view, "apply") {
		t.Error("Expected 'apply' help when filtering")
	}
}
func TestSchemaView_View_FilterWithResults(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.tables = []tableWithColumns{
		{StorageUnit: engine.StorageUnit{Name: "users"}},
		{StorageUnit: engine.StorageUnit{Name: "orders"}},
	}
	v.filteredTables = v.tables[:1]
	v.filterInput.SetValue("user")
	v.filtering = false
	view := v.View()
	if !strings.Contains(view, "(1/2)") {
		t.Error("Expected filter count '(1/2)' in view")
	}
}
func TestSchemaView_View_HelpText(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.filtering = false
	view := v.View()
	if !strings.Contains(view, "expand") {
		t.Error("Expected 'expand' help text")
	}
	if !strings.Contains(view, "view data") {
		t.Error("Expected 'view data' help text")
	}
	if !strings.Contains(view, "filter") {
		t.Error("Expected 'filter' help text")
	}
	if !strings.Contains(view, "refresh") {
		t.Error("Expected 'refresh' help text")
	}
}
func TestSchemaView_Init(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	cmd := v.Init()
	if cmd == nil {
		t.Error("Expected Init to return a command")
	}
}
func TestSchemaView_Refresh(t *testing.T) {
	v, cleanup := setupSchemaViewTest(t)
	defer cleanup()
	v.loading = false
	v.filterInput.SetValue("old filter")
	v.expandedTables["users"] = true
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	v, cmd := v.Update(msg)
	if !v.loading {
		t.Error("Expected loading to be true after refresh")
	}
	if v.filterInput.Value() != "" {
		t.Error("Expected filter to be cleared after refresh")
	}
	if len(v.expandedTables) != 0 {
		t.Error("Expected expandedTables to be cleared after refresh")
	}
	if cmd == nil {
		t.Error("Expected refresh to return a command")
	}
}