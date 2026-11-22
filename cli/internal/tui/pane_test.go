package tui
import (
	"testing"
	tea "github.com/charmbracelet/bubbletea"
)
func TestPaneRegistryPopulated(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel failed: %v", m.err)
	}
	expectedModes := []ViewMode{
		ViewConnection, ViewBrowser, ViewEditor, ViewResults,
		ViewHistory, ViewExport, ViewWhere, ViewColumns,
		ViewChat, ViewSchema, ViewImport, ViewMockData, ViewBookmarks, ViewJSON,
		ViewRowWrite, ViewCmdLog, ViewExplain, ViewDiff, ViewERD, ViewAudit, ViewProfiles,
	}
	if len(m.panes) != len(expectedModes) {
		t.Errorf("Expected %d panes, got %d", len(expectedModes), len(m.panes))
	}
	for _, mode := range expectedModes {
		pane := m.GetPane(mode)
		if pane == nil {
			t.Errorf("Pane for mode %d is nil", mode)
		}
	}
}
func TestActivePaneReturnsCurrentView(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel failed: %v", m.err)
	}
	pane := m.ActivePane()
	if pane == nil {
		t.Fatal("ActivePane returned nil")
	}
	if pane != Pane(m.connectionView) {
		t.Error("ActivePane should return the connectionView pane")
	}
}
func TestPaneSetDimensions(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel failed: %v", m.err)
	}
	for mode, pane := range m.panes {
		pane.SetDimensions(120, 40)
		output := pane.View()
		if output == "" {
			t.Errorf("Pane for mode %d returned empty View() after SetDimensions", mode)
		}
	}
}
func TestPaneFocusable(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel failed: %v", m.err)
	}
	for mode, pane := range m.panes {
		if !pane.Focusable() {
			t.Errorf("Pane for mode %d should be focusable", mode)
		}
	}
}
func TestPaneUpdatePaneDoesNotPanic(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel failed: %v", m.err)
	}
	msg := tea.WindowSizeMsg{Width: 100, Height: 40}
	for mode, pane := range m.panes {
		cmd := pane.UpdatePane(msg)
		_ = cmd
		_ = mode
	}
}
func TestPaneOnFocusOnBlurDoNotPanic(t *testing.T) {
	setupTestEnv(t)
	m := NewMainModel()
	if m.err != nil {
		t.Fatalf("NewMainModel failed: %v", m.err)
	}
	for _, pane := range m.panes {
		pane.OnFocus()
		pane.OnBlur()
	}
}