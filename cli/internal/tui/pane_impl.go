package tui
import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clidey/whodb/cli/pkg/styles"
)
var (
	_ Pane = (*BrowserView)(nil)
	_ Pane = (*ConnectionView)(nil)
	_ Pane = (*EditorView)(nil)
	_ Pane = (*ResultsView)(nil)
	_ Pane = (*HistoryView)(nil)
	_ Pane = (*ExportView)(nil)
	_ Pane = (*WhereView)(nil)
	_ Pane = (*ColumnsView)(nil)
	_ Pane = (*ChatView)(nil)
	_ Pane = (*SchemaView)(nil)
	_ Pane = (*ImportView)(nil)
	_ Pane = (*RowWriteView)(nil)
	_ Pane = (*BookmarksView)(nil)
	_ Pane = (*JSONViewer)(nil)
	_ Pane = (*CmdLogView)(nil)
	_ Pane = (*ExplainView)(nil)
	_ Pane = (*SchemaDiffView)(nil)
	_ Pane = (*ERDView)(nil)
	_ Pane = (*AuditView)(nil)
	_ Pane = (*ProfilesView)(nil)
)
func (v *BrowserView) UpdatePane(msg tea.Msg) tea.Cmd { _, cmd := v.Update(msg); return cmd }
func (v *BrowserView) SetDimensions(width, height int) {
	v.width = width
	v.height = height
	columnWidth := 25
	available := width - 8
	v.columnsPerRow = clamp(available/columnWidth, 1, 6)
	v.filterInput.Width = clamp(width-20, 15, 50)
}
func (v *BrowserView) Focusable() bool   { return true }
func (v *BrowserView) OnFocus()          {}
func (v *BrowserView) OnBlur()           {}
func (v *BrowserView) SetCompact(c bool) { v.compact = c }
func (v *BrowserView) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		Keys.Browser.Up, Keys.Browser.Down, Keys.Browser.Left, Keys.Browser.Right,
		Keys.Browser.Select, Keys.Browser.Filter,
	}
	if len(v.schemas) > 1 {
		bindings = append(bindings, Keys.Browser.Schema)
	}
	bindings = append(bindings,
		Keys.Browser.Editor, Keys.Browser.History, Keys.Browser.Refresh,
	)
	return bindings
}
func (v *ConnectionView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *ConnectionView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ConnectionView) Focusable() bool                 { return true }
func (v *ConnectionView) OnFocus()                        {}
func (v *ConnectionView) OnBlur()                         {}
func (v *ConnectionView) SetCompact(bool)                 {}
func (v *ConnectionView) HelpBindings() []key.Binding     { return nil }
func (v *EditorView) UpdatePane(msg tea.Msg) tea.Cmd { _, cmd := v.Update(msg); return cmd }
func (v *EditorView) SetDimensions(width, height int) {
	v.width = width
	v.height = height
	v.applyWindowSize(width, height)
}
func (v *EditorView) Focusable() bool   { return true }
func (v *EditorView) OnFocus()          { v.textarea.Focus() }
func (v *EditorView) OnBlur()           { v.textarea.Blur() }
func (v *EditorView) SetCompact(c bool) { v.compact = c }
func (v *EditorView) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys(styles.KeyExecute), key.WithHelp(styles.KeyExecute, styles.KeyExecuteDesc)),
		Keys.Editor.Explain,
		Keys.Editor.Format,
		Keys.Editor.OpenEditor,
		Keys.Editor.Bookmarks,
		Keys.Editor.NewTab,
	}
	if len(v.buffers) > 1 {
		bindings = append(bindings, Keys.Editor.PrevTab, Keys.Editor.NextTab, Keys.Editor.CloseTab)
	}
	bindings = append(bindings, Keys.Editor.Clear, Keys.Global.Back)
	return bindings
}
func (v *ResultsView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *ResultsView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ResultsView) Focusable() bool                 { return true }
func (v *ResultsView) OnFocus()                        {}
func (v *ResultsView) OnBlur()                         {}
func (v *ResultsView) SetCompact(c bool)               { v.compact = c }
func (v *ResultsView) HelpBindings() []key.Binding {
	bindings := []key.Binding{
		Keys.Results.NextPage, Keys.Results.ColLeft,
		Keys.Results.ViewCell, Keys.Results.Where, Keys.Results.Columns,
		Keys.Results.Export, Keys.Results.PageSize, Keys.Global.Back,
	}
	if v.tableName != "" {
		bindings = append(bindings, Keys.Results.AddRow, Keys.Results.DeleteRow)
		if v.results != nil && !v.results.DisableUpdate {
			bindings = append(bindings, Keys.Results.EditRow)
		}
	}
	return bindings
}
func (v *HistoryView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *HistoryView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *HistoryView) Focusable() bool                 { return true }
func (v *HistoryView) OnFocus()                        { v.refreshList() }
func (v *HistoryView) OnBlur()                         {}
func (v *HistoryView) SetCompact(bool)                 {}
func (v *HistoryView) HelpBindings() []key.Binding     { return nil }
func (v *ExportView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *ExportView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ExportView) Focusable() bool                 { return true }
func (v *ExportView) OnFocus()                        {}
func (v *ExportView) OnBlur()                         {}
func (v *ExportView) SetCompact(bool)                 {}
func (v *ExportView) HelpBindings() []key.Binding     { return nil }
func (v *WhereView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *WhereView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *WhereView) Focusable() bool                 { return true }
func (v *WhereView) OnFocus()                        {}
func (v *WhereView) OnBlur()                         {}
func (v *WhereView) SetCompact(bool)                 {}
func (v *WhereView) HelpBindings() []key.Binding     { return nil }
func (v *ColumnsView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *ColumnsView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ColumnsView) Focusable() bool                 { return true }
func (v *ColumnsView) OnFocus()                        {}
func (v *ColumnsView) OnBlur()                         {}
func (v *ColumnsView) SetCompact(bool)                 {}
func (v *ColumnsView) HelpBindings() []key.Binding     { return nil }
func (v *ChatView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *ChatView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ChatView) Focusable() bool                 { return true }
func (v *ChatView) OnFocus()                        {}
func (v *ChatView) OnBlur()                         {}
func (v *ChatView) SetCompact(bool)                 {}
func (v *ChatView) HelpBindings() []key.Binding     { return nil }
func (v *SchemaView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *SchemaView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *SchemaView) Focusable() bool                 { return true }
func (v *SchemaView) OnFocus()                        {}
func (v *SchemaView) OnBlur()                         {}
func (v *SchemaView) SetCompact(bool)                 {}
func (v *SchemaView) HelpBindings() []key.Binding     { return nil }
func (v *ImportView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *ImportView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ImportView) Focusable() bool                 { return true }
func (v *ImportView) OnFocus()                        {}
func (v *ImportView) OnBlur()                         {}
func (v *ImportView) SetCompact(bool)                 {}
func (v *ImportView) HelpBindings() []key.Binding     { return nil }
func (v *MockDataView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *MockDataView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *MockDataView) Focusable() bool                 { return true }
func (v *MockDataView) OnFocus()                        {}
func (v *MockDataView) OnBlur()                         {}
func (v *MockDataView) SetCompact(bool)                 {}
func (v *MockDataView) HelpBindings() []key.Binding     { return nil }
func (v *RowWriteView) UpdatePane(msg tea.Msg) tea.Cmd { _, cmd := v.Update(msg); return cmd }
func (v *RowWriteView) SetDimensions(width, height int) {
	v.width = width
	v.height = height
	v.applyDimensions()
}
func (v *RowWriteView) Focusable() bool { return true }
func (v *RowWriteView) OnFocus()        { v.syncAddFocus() }
func (v *RowWriteView) OnBlur() {
	v.textarea.Blur()
	for idx := range v.inputs {
		v.inputs[idx].Blur()
	}
}
func (v *RowWriteView) SetCompact(bool)             {}
func (v *RowWriteView) HelpBindings() []key.Binding { return rowWriteHelpBindings(v.action) }
func (v *JSONViewer) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *JSONViewer) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *JSONViewer) Focusable() bool                 { return true }
func (v *JSONViewer) OnFocus()                        {}
func (v *JSONViewer) OnBlur()                         {}
func (v *JSONViewer) SetCompact(bool)                 {}
func (v *JSONViewer) HelpBindings() []key.Binding     { return nil }
func (v *BookmarksView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *BookmarksView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *BookmarksView) Focusable() bool                 { return true }
func (v *BookmarksView) OnFocus()                        {}
func (v *BookmarksView) OnBlur()                         {}
func (v *BookmarksView) SetCompact(bool)                 {}
func (v *BookmarksView) HelpBindings() []key.Binding     { return nil }
func (v *CmdLogView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *CmdLogView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *CmdLogView) Focusable() bool                 { return true }
func (v *CmdLogView) OnFocus()                        {}
func (v *CmdLogView) OnBlur()                         {}
func (v *CmdLogView) SetCompact(bool)                 {}
func (v *CmdLogView) HelpBindings() []key.Binding     { return nil }
func (v *ExplainView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *ExplainView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ExplainView) Focusable() bool                 { return true }
func (v *ExplainView) OnFocus()                        {}
func (v *ExplainView) OnBlur()                         {}
func (v *ExplainView) SetCompact(bool)                 {}
func (v *ExplainView) HelpBindings() []key.Binding     { return nil }
func (v *SchemaDiffView) UpdatePane(msg tea.Msg) tea.Cmd { _, cmd := v.Update(msg); return cmd }
func (v *SchemaDiffView) SetDimensions(width, height int) {
	v.width = width
	v.height = height
	if v.result != nil && !v.editing {
		v.rebuildViewport()
	}
}
func (v *SchemaDiffView) Focusable() bool { return true }
func (v *SchemaDiffView) OnFocus()        { v.syncFocus() }
func (v *SchemaDiffView) OnBlur() {
	v.fromSchemaInput.Blur()
	v.toSchemaInput.Blur()
}
func (v *SchemaDiffView) SetCompact(bool)             {}
func (v *SchemaDiffView) HelpBindings() []key.Binding { return nil }
func (v *AuditView) UpdatePane(msg tea.Msg) tea.Cmd { _, cmd := v.Update(msg); return cmd }
func (v *AuditView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *AuditView) Focusable() bool { return true }
func (v *AuditView) OnFocus() {}
func (v *AuditView) OnBlur() {}
func (v *AuditView) SetCompact(bool) {}
func (v *AuditView) HelpBindings() []key.Binding { return nil }
func (v *ERDView) UpdatePane(msg tea.Msg) tea.Cmd { _, cmd := v.Update(msg); return cmd }
func (v *ERDView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ERDView) Focusable() bool { return true }
func (v *ERDView) OnFocus() {}
func (v *ERDView) OnBlur() {}
func (v *ERDView) SetCompact(bool) {}
func (v *ERDView) HelpBindings() []key.Binding { return nil }
func (v *ProfilesView) UpdatePane(msg tea.Msg) tea.Cmd  { _, cmd := v.Update(msg); return cmd }
func (v *ProfilesView) SetDimensions(width, height int) { v.width = width; v.height = height }
func (v *ProfilesView) Focusable() bool                 { return true }
func (v *ProfilesView) OnFocus()                        {}
func (v *ProfilesView) OnBlur()                         {}
func (v *ProfilesView) SetCompact(bool)                 {}
func (v *ProfilesView) HelpBindings() []key.Binding     { return nil }