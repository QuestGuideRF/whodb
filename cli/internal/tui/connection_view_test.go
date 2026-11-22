package tui
import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clidey/whodb/cli/internal/config"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/providers"
)
func setupConnectionViewTest(t *testing.T) (*ConnectionView, func()) {
	t.Helper()
	setupTestEnv(t)
	clearedEnv := make(map[string]string)
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		key := parts[0]
		if strings.HasPrefix(key, "WHODB_") && !strings.HasPrefix(key, "WHODB_CLI_") {
			clearedEnv[key] = os.Getenv(key)
			os.Unsetenv(key)
		}
	}
	parent := NewMainModel()
	if parent.err != nil {
		t.Fatalf("Failed to create MainModel: %v", parent.err)
	}
	cleanup := func() {
		for key, value := range clearedEnv {
			os.Setenv(key, value)
		}
	}
	return parent.connectionView, cleanup
}
func requireDBTypeIndex(t *testing.T, v *ConnectionView, dbType string) int {
	t.Helper()
	for i, candidate := range v.dbTypes {
		if candidate == dbType {
			return i
		}
	}
	t.Fatalf("database type %q not found in connection view catalog", dbType)
	return -1
}
type connectionViewTestProvider struct {
	id          string
	name        string
	connections []providers.DiscoveredConnection
}
func (p *connectionViewTestProvider) Type() providers.ProviderType { return providers.ProviderTypeAWS }
func (p *connectionViewTestProvider) ID() string                   { return p.id }
func (p *connectionViewTestProvider) Name() string                 { return p.name }
func (p *connectionViewTestProvider) DiscoverConnections(ctx context.Context) ([]providers.DiscoveredConnection, error) {
	return append([]providers.DiscoveredConnection(nil), p.connections...), nil
}
func (p *connectionViewTestProvider) TestConnection(ctx context.Context) error { return nil }
func (p *connectionViewTestProvider) RefreshConnection(ctx context.Context, connectionID string) (bool, error) {
	return false, nil
}
func (p *connectionViewTestProvider) Close(ctx context.Context) error { return nil }
func TestNewConnectionView_NoConnections(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	if v == nil {
		t.Fatal("NewConnectionView returned nil")
	}
	if v.mode != "form" && v.mode != "list" {
		t.Errorf("Expected mode 'form' or 'list', got '%s'", v.mode)
	}
	if v.focusIndex != focusDBType {
		t.Errorf("Expected focusIndex 7 (db type), got %d", v.focusIndex)
	}
}
func TestNewConnectionView_WithConnections(t *testing.T) {
	setupTestEnv(t)
	cfg, _ := config.LoadConfig()
	cfg.AddConnection(config.Connection{Name: "test", Type: "Postgres", Host: "localhost"})
	cfg.Save()
	parent := NewMainModel()
	if parent.err != nil {
		t.Fatalf("Failed to create MainModel: %v", parent.err)
	}
	v := parent.connectionView
	if v.mode != "list" {
		t.Errorf("Expected mode 'list' with connections, got '%s'", v.mode)
	}
}
func TestConnectionView_LoadsCloudConnectionsAndPrefillsForm(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	originalAWS := env.IsAWSProviderEnabled
	originalAzure := env.IsAzureProviderEnabled
	originalGCP := env.IsGCPProviderEnabled
	t.Cleanup(func() {
		env.IsAWSProviderEnabled = originalAWS
		env.IsAzureProviderEnabled = originalAzure
		env.IsGCPProviderEnabled = originalGCP
		_ = providers.GetDefaultRegistry().Close(context.Background())
	})
	env.IsAWSProviderEnabled = true
	env.IsAzureProviderEnabled = false
	env.IsGCPProviderEnabled = false
	if err := providers.GetDefaultRegistry().Register(&connectionViewTestProvider{
		id:   "aws-prod",
		name: "AWS Prod",
		connections: []providers.DiscoveredConnection{{
			ID:           "aws-prod/prod-db",
			ProviderType: providers.ProviderTypeAWS,
			ProviderID:   "aws-prod",
			Name:         "prod-db",
			DatabaseType: engine.DatabaseType_Postgres,
			Status:       providers.ConnectionStatusAvailable,
			Metadata: map[string]string{
				"endpoint": "prod-db.example.com",
				"port":     "5432",
			},
		}},
	}); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	msg := v.loadCloudConnections()()
	v, _ = v.Update(msg)
	foundCloud := false
	for _, item := range v.list.Items() {
		ci := item.(connectionItem)
		if ci.source == ConnectionSourceCloud {
			foundCloud = true
			v.list.Select(v.list.Index() + 1)
			break
		}
	}
	if !foundCloud {
		t.Fatal("expected a cloud connection item in the list")
	}
	v.list.Select(len(v.list.Items()) - 1)
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if v.mode != "form" {
		t.Fatalf("expected selecting a cloud item to open the form, got %q", v.mode)
	}
	if got := v.inputs[fieldHost].Value(); got != "prod-db.example.com" {
		t.Fatalf("expected discovered host prefill, got %q", got)
	}
	if got := v.advanced["SSL Mode"]; got != "require" {
		t.Fatalf("expected discovered advanced SSL prefill, got %#v", v.advanced)
	}
}
func TestConnectionView_ListMode_Navigation_Tab(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.parent.config.AddConnection(config.Connection{Name: "conn1"})
	v.parent.config.AddConnection(config.Connection{Name: "conn2"})
	_ = v.parent.config.Save()
	v.refreshList()
	msg := tea.KeyMsg{Type: tea.KeyTab}
	v, _ = v.Update(msg)
}
func TestConnectionView_ListMode_Navigation_ShiftTab(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.parent.config.AddConnection(config.Connection{Name: "conn1"})
	v.parent.config.AddConnection(config.Connection{Name: "conn2"})
	_ = v.parent.config.Save()
	v.refreshList()
	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	v, _ = v.Update(msg)
}
func TestConnectionView_ListMode_NewConnection(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	v, _ = v.Update(msg)
	if v.mode != "form" {
		t.Errorf("Expected mode 'form' after 'n', got '%s'", v.mode)
	}
}
func TestConnectionView_ListMode_Delete(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.parent.config.AddConnection(config.Connection{Name: "to-delete", Type: "Postgres"})
	_ = v.parent.config.Save()
	v.refreshList()
	initialCount := len(v.parent.config.Connections)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	v, _ = v.Update(msg)
	if len(v.parent.config.Connections) != initialCount-1 {
		t.Error("Expected connection to be deleted")
	}
}
func TestConnectionView_RenderWrappedSelectableOptions_CachesResult(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	options := []string{"Postgres", "MySQL", "Sqlite3"}
	rendered1, lineCount1 := v.renderWrappedSelectableOptions(options, 1, true, 24)
	rendered2, lineCount2 := v.renderWrappedSelectableOptions(options, 1, true, 24)
	if rendered1 != rendered2 {
		t.Fatalf("expected cached wrapped options render to match, got %q vs %q", rendered1, rendered2)
	}
	if lineCount1 != lineCount2 {
		t.Fatalf("expected cached wrapped option line count to match, got %d vs %d", lineCount1, lineCount2)
	}
	if got := len(v.selectorCache); got != 1 {
		t.Fatalf("expected selector cache size 1, got %d", got)
	}
}
func TestConnectionView_ListMode_EscConfirmation(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.escPressed = false
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	v, cmd := v.Update(msg)
	if !v.escPressed {
		t.Error("Expected escPressed to be true after first ESC")
	}
	if v.escTimeoutSecs != 3 {
		t.Errorf("Expected escTimeoutSecs 3, got %d", v.escTimeoutSecs)
	}
	if cmd == nil {
		t.Error("Expected tick command after first ESC")
	}
}
func TestConnectionView_ListMode_EscConfirmation_SecondEsc(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.escPressed = true
	v.escTimeoutSecs = 2
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := v.Update(msg)
	if cmd == nil {
		t.Error("Expected quit command after second ESC")
	}
}
func TestRenderWrappedSelectableOptions_WrapsLongLists(t *testing.T) {
	options := []string{
		"Postgres", "MySQL", "MariaDB", "CockroachDB", "Sqlite3", "MongoDB",
		"Redis", "ElasticSearch", "ClickHouse", "DuckDB", "Memcached", "TiDB",
	}
	rendered, lineCount := renderWrappedSelectableOptions(options, 0, false, 32)
	if lineCount <= 1 {
		t.Fatalf("expected wrapped options to span multiple lines, got %d", lineCount)
	}
	if !strings.Contains(rendered, "\n  ") {
		t.Fatalf("expected wrapped options to contain an indented newline, got %q", rendered)
	}
}
func TestConnectionView_DbTypeSectionHeight_GrowsWhenWrapped(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.width = 48
	v.parent.width = 48
	height := v.dbTypeSectionHeight()
	if height <= 3 {
		t.Fatalf("expected wrapped database type section height greater than 3, got %d", height)
	}
}
func TestConnectionView_ListMode_EscTimeoutTick(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.escPressed = true
	v.escTimeoutSecs = 2
	msg := escTimeoutTickMsg{}
	v, _ = v.Update(msg)
	if v.escTimeoutSecs != 1 {
		t.Errorf("Expected escTimeoutSecs 1 after tick, got %d", v.escTimeoutSecs)
	}
	v, _ = v.Update(msg)
	if v.escPressed {
		t.Error("Expected escPressed to be false after timeout")
	}
}
func TestConnectionView_FormMode_NextInput(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.focusIndex = 0
	msg := tea.KeyMsg{Type: tea.KeyTab}
	v, _ = v.Update(msg)
	if v.focusIndex != 1 {
		t.Errorf("Expected focusIndex 1 after Tab from 0, got %d", v.focusIndex)
	}
}
func TestConnectionView_FormMode_PrevInput(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.focusIndex = 2
	msg := tea.KeyMsg{Type: tea.KeyShiftTab}
	v, _ = v.Update(msg)
	if v.focusIndex != 1 {
		t.Errorf("Expected focusIndex 1 after Shift+Tab, got %d", v.focusIndex)
	}
}
func TestConnectionView_FormMode_WrapAround(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.focusIndex = focusConnect
	v.nextInput()
	if v.focusIndex != focusDBType {
		t.Errorf("Expected focusIndex 7 (db type) after wrap, got %d", v.focusIndex)
	}
	v.prevInput()
	if v.focusIndex != focusConnect {
		t.Errorf("Expected focusIndex 8 after backward wrap, got %d", v.focusIndex)
	}
}
func TestConnectionView_FormMode_DbTypeSelection(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.focusIndex = focusDBType
	v.dbTypeIndex = 0
	msg := tea.KeyMsg{Type: tea.KeyRight}
	v, _ = v.Update(msg)
	if v.dbTypeIndex != 1 {
		t.Errorf("Expected dbTypeIndex 1 after Right, got %d", v.dbTypeIndex)
	}
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	v, _ = v.Update(msg)
	if v.dbTypeIndex != 0 {
		t.Errorf("Expected dbTypeIndex 0 after Left, got %d", v.dbTypeIndex)
	}
}
func TestConnectionView_FormMode_DbTypeWrapAround(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.focusIndex = focusDBType
	v.dbTypeIndex = 0
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	v, _ = v.Update(msg)
	if v.dbTypeIndex != len(v.dbTypes)-1 {
		t.Errorf("Expected dbTypeIndex %d (last), got %d", len(v.dbTypes)-1, v.dbTypeIndex)
	}
	msg = tea.KeyMsg{Type: tea.KeyRight}
	v, _ = v.Update(msg)
	if v.dbTypeIndex != 0 {
		t.Errorf("Expected dbTypeIndex 0 after wrap, got %d", v.dbTypeIndex)
	}
}
func TestConnectionView_FormMode_EscapeToList(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.parent.config.AddConnection(config.Connection{Name: "test"})
	v.mode = "form"
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	v, _ = v.Update(msg)
	if v.mode != "list" {
		t.Errorf("Expected mode 'list' after Esc, got '%s'", v.mode)
	}
}
func TestConnectionView_FormMode_PasswordPrompt(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.focusIndex = focusConnect
	v.inputs[4].SetValue("")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	v, _ = v.Update(msg)
	if !v.awaitingPassword {
		t.Error("Expected awaitingPassword to be true with empty password")
	}
}
func TestConnectionView_FormMode_PasswordPrompt_Cancel(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.awaitingPassword = true
	v.passwordPrompt.SetValue("secret")
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	v, _ = v.Update(msg)
	if v.awaitingPassword {
		t.Error("Expected awaitingPassword to be false after Esc")
	}
	if v.passwordPrompt.Value() != "" {
		t.Error("Expected password prompt to be cleared")
	}
}
func TestConnectionView_ResetForm(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.inputs[0].SetValue("test connection")
	v.inputs[1].SetValue("myhost")
	v.focusIndex = 5
	v.dbTypeIndex = 3
	v.resetForm()
	for i, input := range v.inputs {
		if input.Value() != "" {
			t.Errorf("Expected input %d to be cleared", i)
		}
	}
	if v.focusIndex != focusDBType {
		t.Errorf("Expected focusIndex 7 (db type) after reset, got %d", v.focusIndex)
	}
	if v.dbTypeIndex != 0 {
		t.Errorf("Expected dbTypeIndex 0 after reset, got %d", v.dbTypeIndex)
	}
}
func TestConnectionView_GetDefaultPort(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	tests := []struct {
		dbType   string
		expected int
	}{
		{"Postgres", 5432},
		{"MySQL", 3306},
		{"MariaDB", 3306},
		{"MongoDB", 27017},
		{"Redis", 6379},
		{"ClickHouse", 9000},
		{"ElasticSearch", 9200},
		{"Sqlite3", 0},
		{"Unknown", 0},
	}
	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := v.getDefaultPort(tt.dbType)
			if result != tt.expected {
				t.Errorf("getDefaultPort(%s) = %d, expected %d", tt.dbType, result, tt.expected)
			}
		})
	}
}
func TestConnectionView_UpdatePortPlaceholder(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.dbTypeIndex = 1
	v.updatePortPlaceholder()
	if v.inputs[2].Placeholder != "3306" {
		t.Errorf("Expected port placeholder '3306' for MySQL, got '%s'", v.inputs[2].Placeholder)
	}
}
func TestConnectionView_WindowSizeMsg(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	v, _ = v.Update(msg)
}
func TestConnectionView_MouseScroll_List(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.parent.config.AddConnection(config.Connection{Name: "conn1"})
	v.parent.config.AddConnection(config.Connection{Name: "conn2"})
	_ = v.parent.config.Save()
	v.refreshList()
	msg := tea.MouseMsg{Button: tea.MouseButtonWheelDown}
	v, _ = v.Update(msg)
	msg = tea.MouseMsg{Button: tea.MouseButtonWheelUp}
	v, _ = v.Update(msg)
}
func TestConnectionView_MouseScroll_Form(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.focusIndex = 0
	v, _ = v.Update(tea.WindowSizeMsg{Width: 80, Height: 40})
	if !v.formReady {
		t.Error("Expected formReady to be true after WindowSizeMsg")
	}
	msg := tea.MouseMsg{Button: tea.MouseButtonWheelDown}
	v, _ = v.Update(msg)
	if v.focusIndex != 0 {
		t.Errorf("Expected focusIndex 0 after wheel down (viewport scroll), got %d", v.focusIndex)
	}
	v, _ = v.Update(tea.KeyMsg{Type: tea.KeyTab})
	if v.focusIndex == 0 {
		t.Error("Expected focusIndex to change after tab")
	}
}
func TestConnectionView_View_List(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.parent.config.AddConnection(config.Connection{Name: "my-db", Type: "Postgres", Host: "localhost"})
	_ = v.parent.config.Save()
	v.refreshList()
	view := v.View()
	if !strings.Contains(view, "Welcome to WhoDB") {
		t.Error("Expected 'Welcome to WhoDB' title")
	}
	if !strings.Contains(view, "new") {
		t.Error("Expected 'new' help text")
	}
	if !strings.Contains(view, "delete") {
		t.Error("Expected 'delete' help text")
	}
}
func TestConnectionView_View_ListWithEscConfirmation(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.escPressed = true
	v.escTimeoutSecs = 2
	view := v.View()
	if !strings.Contains(view, "Press ESC again to quit") {
		t.Error("Expected ESC confirmation message")
	}
	if !strings.Contains(view, "2s") {
		t.Error("Expected countdown in confirmation message")
	}
}
func TestConnectionView_View_Form(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	view := v.View()
	if !strings.Contains(view, "New Database Connection") {
		t.Error("Expected 'New Database Connection' title")
	}
	if !strings.Contains(view, "Connection Name") {
		t.Error("Expected 'Connection Name' field")
	}
	if !strings.Contains(view, "Host") {
		t.Error("Expected 'Host' field")
	}
	if !strings.Contains(view, "Port") {
		t.Error("Expected 'Port' field")
	}
	if !strings.Contains(view, "Database Type") {
		t.Error("Expected 'Database Type' field")
	}
	if !strings.Contains(view, "Connect") {
		t.Error("Expected 'Connect' button")
	}
}
func TestConnectionView_View_FormWithError(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.connError = os.ErrInvalid
	view := v.View()
	if !strings.Contains(view, "invalid") {
		t.Error("Expected error message in view")
	}
}
func TestConnectionView_View_PasswordPrompt(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.awaitingPassword = true
	view := v.View()
	if !strings.Contains(view, "Enter Password") {
		t.Error("Expected 'Enter Password' title")
	}
	if !strings.Contains(view, "confirm") {
		t.Error("Expected 'confirm' help text")
	}
	if !strings.Contains(view, "cancel") {
		t.Error("Expected 'cancel' help text")
	}
}
func TestConnectionItem_Methods(t *testing.T) {
	conn := config.Connection{
		Name: "my-connection",
		Type: "Postgres",
		Host: "db.example.com",
	}
	item := connectionItem{conn: conn}
	if item.Title() != "my-connection" {
		t.Errorf("Expected Title 'my-connection', got '%s'", item.Title())
	}
	if item.Description() != "Postgres@db.example.com" {
		t.Errorf("Expected Description 'Postgres@db.example.com', got '%s'", item.Description())
	}
	if item.FilterValue() != "my-connection" {
		t.Errorf("Expected FilterValue 'my-connection', got '%s'", item.FilterValue())
	}
}
func TestConnectionDelegate_Methods(t *testing.T) {
	d := connectionDelegate{}
	if d.Height() != 2 {
		t.Errorf("Expected Height 2, got %d", d.Height())
	}
	if d.Spacing() != 1 {
		t.Errorf("Expected Spacing 1, got %d", d.Spacing())
	}
}
func TestConnectionView_ConnectionResultMsg_Success(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.connecting = true
	msg := connectionResultMsg{err: nil}
	v, _ = v.Update(msg)
	if v.parent.mode != ViewBrowser {
		t.Errorf("Expected mode ViewBrowser after successful connection, got %v", v.parent.mode)
	}
}
func TestConnectionView_ConnectionResultMsg_UsesProvidedStatusMessage(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.connecting = true
	msg := connectionResultMsg{
		err:           nil,
		statusMessage: "Connected to Postgres@localhost • SSL/TLS: enabled (verify-full)",
	}
	v, _ = v.Update(msg)
	if v.parent.statusMessage != msg.statusMessage {
		t.Fatalf("Expected status message %q, got %q", msg.statusMessage, v.parent.statusMessage)
	}
}
func TestConnectionView_ConnectionResultMsg_Error(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.connecting = true
	msg := connectionResultMsg{err: os.ErrPermission}
	v, _ = v.Update(msg)
	if v.connecting {
		t.Error("Expected connecting to become false after error")
	}
	if v.connError == nil {
		t.Error("Expected connError to be set")
	}
	if v.mode != "form" {
		t.Errorf("Expected mode 'form' after error, got '%s'", v.mode)
	}
}
func TestGetVisibleFields_AllTypes(t *testing.T) {
	tests := []struct {
		dbType   string
		expected []int
	}{
		{"Postgres", []int{0, 1, 2, 3, 4, 5, 6}},
		{"MySQL", []int{0, 1, 2, 3, 4, 5}},
		{"MariaDB", []int{0, 1, 2, 3, 4, 5}},
		{"ClickHouse", []int{0, 1, 2, 3, 4, 5}},
		{"Sqlite3", []int{0, 5}},
		{"MongoDB", []int{0, 1, 2, 3, 4, 5}},
		{"Redis", []int{0, 1, 2, 3, 4}},
		{"ElasticSearch", []int{0, 1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := getVisibleFields(tt.dbType)
			if len(result) != len(tt.expected) {
				t.Errorf("getVisibleFields(%s) returned %d fields, expected %d", tt.dbType, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("getVisibleFields(%s)[%d] = %d, expected %d", tt.dbType, i, v, tt.expected[i])
				}
			}
		})
	}
}
func TestConnectionView_IsFieldVisible(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	for i := 0; i < 7; i++ {
		if !v.isFieldVisible(i) {
			t.Errorf("Expected field %d to be visible for Postgres", i)
		}
	}
	v.visibleFields = getVisibleFields("Sqlite3")
	if !v.isFieldVisible(0) {
		t.Error("Expected field 0 (name) to be visible for SQLite")
	}
	if v.isFieldVisible(1) {
		t.Error("Expected field 1 (host) to be hidden for SQLite")
	}
	if v.isFieldVisible(2) {
		t.Error("Expected field 2 (port) to be hidden for SQLite")
	}
	if v.isFieldVisible(3) {
		t.Error("Expected field 3 (username) to be hidden for SQLite")
	}
	if v.isFieldVisible(4) {
		t.Error("Expected field 4 (password) to be hidden for SQLite")
	}
	if !v.isFieldVisible(5) {
		t.Error("Expected field 5 (database) to be visible for SQLite")
	}
	if v.isFieldVisible(6) {
		t.Error("Expected field 6 (schema) to be hidden for SQLite")
	}
}
func TestConnectionView_OnDbTypeChanged_UpdatesVisibility(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.dbTypeIndex = requireDBTypeIndex(t, v, "Sqlite3")
	v.onDbTypeChanged()
	if len(v.visibleFields) != 2 {
		t.Errorf("Expected 2 visible fields for SQLite, got %d", len(v.visibleFields))
	}
	if v.inputs[5].Placeholder != "/path/to/database.db" {
		t.Errorf("Expected SQLite database placeholder, got '%s'", v.inputs[5].Placeholder)
	}
	v.dbTypeIndex = 0
	v.onDbTypeChanged()
	if len(v.visibleFields) != 7 {
		t.Errorf("Expected 7 visible fields for Postgres, got %d", len(v.visibleFields))
	}
	if v.inputs[5].Placeholder != "mydb" {
		t.Errorf("Expected 'mydb' database placeholder, got '%s'", v.inputs[5].Placeholder)
	}
}
func TestConnectionView_NavigationSkipsHiddenFields_SQLite(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.dbTypeIndex = requireDBTypeIndex(t, v, "Sqlite3")
	v.onDbTypeChanged()
	v.focusIndex = focusDBType
	v.nextInput()
	if v.focusIndex != 0 {
		t.Errorf("Expected focusIndex 0 (name) after Tab from dbType in SQLite, got %d", v.focusIndex)
	}
	v.nextInput()
	if v.focusIndex != 5 {
		t.Errorf("Expected focusIndex 5 (database) after Tab from name in SQLite, got %d", v.focusIndex)
	}
	v.nextInput()
	if v.focusIndex != focusConnect {
		t.Errorf("Expected focusIndex 8 (connect) after Tab from database in SQLite, got %d", v.focusIndex)
	}
	v.nextInput()
	if v.focusIndex != focusDBType {
		t.Errorf("Expected focusIndex 7 (dbType) after wrap, got %d", v.focusIndex)
	}
}
func TestConnectionView_PrevNavigationSkipsHiddenFields_SQLite(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.dbTypeIndex = requireDBTypeIndex(t, v, "Sqlite3")
	v.onDbTypeChanged()
	v.focusIndex = 5
	v.prevInput()
	if v.focusIndex != 0 {
		t.Errorf("Expected focusIndex 0 (name) after Shift+Tab from database in SQLite, got %d", v.focusIndex)
	}
	v.prevInput()
	if v.focusIndex != focusDBType {
		t.Errorf("Expected focusIndex 7 (dbType) after Shift+Tab from name, got %d", v.focusIndex)
	}
	v.prevInput()
	if v.focusIndex != focusConnect {
		t.Errorf("Expected focusIndex 8 (connect) after backward wrap, got %d", v.focusIndex)
	}
}
func TestConnectionView_FormView_SQLiteHidesFields(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.dbTypeIndex = requireDBTypeIndex(t, v, "Sqlite3")
	v.onDbTypeChanged()
	view := v.View()
	if !strings.Contains(view, "Connection Name") {
		t.Error("Expected 'Connection Name' field for SQLite")
	}
	if !strings.Contains(view, "Database") {
		t.Error("Expected 'Database' field for SQLite")
	}
	if strings.Contains(view, "Host:") {
		t.Error("Expected 'Host' field to be hidden for SQLite")
	}
	if strings.Contains(view, "Port:") {
		t.Error("Expected 'Port' field to be hidden for SQLite")
	}
	if strings.Contains(view, "Username:") {
		t.Error("Expected 'Username' field to be hidden for SQLite")
	}
	if strings.Contains(view, "Password:") {
		t.Error("Expected 'Password' field to be hidden for SQLite")
	}
	if strings.Contains(view, "Schema:") {
		t.Error("Expected 'Schema' field to be hidden for SQLite")
	}
}
func TestConnectionView_FormView_PostgresShowsAllFields(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.dbTypeIndex = 0
	v.onDbTypeChanged()
	view := v.View()
	for _, field := range []string{"Connection Name", "Host:", "Port:", "Username:", "Password:", "Database:", "Schema:"} {
		if !strings.Contains(view, field) {
			t.Errorf("Expected '%s' field to be visible for Postgres", field)
		}
	}
	if !strings.Contains(view, "SSL Mode:") {
		t.Error("Expected SSL mode selector to be visible for Postgres")
	}
}
func TestConnectionView_PasswordPromptSkipped_SQLite(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.dbTypeIndex = requireDBTypeIndex(t, v, "Sqlite3")
	v.onDbTypeChanged()
	v.focusIndex = focusConnect
	v.inputs[4].SetValue("")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	v, _ = v.Update(msg)
	if v.awaitingPassword {
		t.Error("Expected awaitingPassword to be false for SQLite (password field hidden)")
	}
}
func TestConnectionView_PasswordPromptSkipped_RedisOptionalPassword(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.dbTypeIndex = requireDBTypeIndex(t, v, "Redis")
	v.onDbTypeChanged()
	v.focusIndex = focusConnect
	v.inputs[fieldPassword].SetValue("")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	v, _ = v.Update(msg)
	if v.awaitingPassword {
		t.Error("Expected awaitingPassword to be false for Redis when password is optional")
	}
}
func TestConnectionView_ResetFormUpdatesVisibility(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "form"
	v.dbTypeIndex = requireDBTypeIndex(t, v, "Sqlite3")
	v.onDbTypeChanged()
	v.resetForm()
	if v.dbTypeIndex != 0 {
		t.Errorf("Expected dbTypeIndex 0 after reset, got %d", v.dbTypeIndex)
	}
	if v.focusIndex != focusDBType {
		t.Errorf("Expected focusIndex 7 (db type) after reset, got %d", v.focusIndex)
	}
	if len(v.visibleFields) != 7 {
		t.Errorf("Expected 7 visible fields after reset (Postgres), got %d", len(v.visibleFields))
	}
}
func TestConnectionView_RefreshList(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.parent.config.AddConnection(config.Connection{Name: "new-conn"})
	_ = v.parent.config.Save()
	v.refreshList()
	items := v.list.Items()
	found := false
	for _, item := range items {
		if ci, ok := item.(connectionItem); ok && ci.conn.Name == "new-conn" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'new-conn' in refreshed list, got %d items", len(items))
	}
}
func TestConnectionView_ListConnect_IsAsync(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	conn := config.Connection{
		Name:     "test-async",
		Type:     "Postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "user",
		Database: "testdb",
	}
	v.parent.config.AddConnection(conn)
	_ = v.parent.config.Save()
	v.refreshList()
	v.mode = "list"
	items := v.list.Items()
	if len(items) == 0 {
		t.Skip("No items in list after refresh")
	}
	if _, ok := v.list.SelectedItem().(connectionItem); !ok {
		t.Skip("No selected connection item")
	}
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	v, cmd := v.Update(msg)
	if !v.connecting {
		t.Error("Expected connecting to be true after Enter on saved connection")
	}
	if cmd == nil {
		t.Error("Expected a tea.Cmd to be returned for async connection")
	}
}
func TestConnectionView_ListConnect_ShowsSpinner(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.connecting = true
	view := v.View()
	if !strings.Contains(view, "Connecting...") {
		t.Error("Expected 'Connecting...' spinner when connecting from list")
	}
}
func TestConnectionView_ListConnect_ConnectionResult_Error(t *testing.T) {
	v, cleanup := setupConnectionViewTest(t)
	defer cleanup()
	v.mode = "list"
	v.connecting = true
	msg := connectionResultMsg{err: fmt.Errorf("connection refused")}
	v, _ = v.Update(msg)
	if v.connecting {
		t.Error("Expected connecting to be false after error")
	}
	if v.parent.err == nil {
		t.Error("Expected parent.err to be set on connection failure")
	}
}