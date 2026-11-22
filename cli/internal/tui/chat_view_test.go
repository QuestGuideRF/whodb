package tui
import (
	"context"
	"errors"
	"strings"
	"testing"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/clidey/whodb/cli/internal/config"
	"github.com/clidey/whodb/cli/internal/database"
	"github.com/clidey/whodb/core/src/engine"
)
func setupChatViewTest(t *testing.T) (*ChatView, func()) {
	t.Helper()
	setupTestEnv(t)
	parent := NewMainModel()
	if parent.err != nil {
		t.Fatalf("Failed to create MainModel: %v", parent.err)
	}
	cleanup := func() {}
	return parent.chatView, cleanup
}
func TestNewChatView(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	if v == nil {
		t.Fatal("NewChatView returned nil")
	}
	if v.selectedProvider < 0 {
		t.Error("Expected selectedProvider >= 0")
	}
	if len(v.messages) != 0 {
		t.Error("Expected empty messages initially")
	}
	if v.sending {
		t.Error("Expected sending to be false initially")
	}
	if v.selectedMessage != -1 {
		t.Errorf("Expected selectedMessage -1, got %d", v.selectedMessage)
	}
	if v.focusField != focusFieldMessage {
		t.Errorf("Expected focusField focusFieldMessage, got %d", v.focusField)
	}
}
func TestChatView_ConsentAccept(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = false
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	v, _ = v.Update(msg)
	if !v.consented {
		t.Error("Expected consented to be true after 'a'")
	}
}
func TestChatView_ConsentCancel_Esc(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = false
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	v, _ = v.Update(msg)
	if v.consented {
		t.Error("Expected consented to remain false after Esc")
	}
	if v.parent.mode != ViewBrowser {
		t.Errorf("Expected mode ViewBrowser after Esc, got %v", v.parent.mode)
	}
}
func TestChatView_ConsentCancel_Q(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = false
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	v, _ = v.Update(msg)
	if v.parent.mode != ViewBrowser {
		t.Errorf("Expected mode ViewBrowser after 'q', got %v", v.parent.mode)
	}
}
func TestChatView_ConsentCancel_D(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = false
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	v, _ = v.Update(msg)
	if v.parent.mode != ViewBrowser {
		t.Errorf("Expected mode ViewBrowser after 'd', got %v", v.parent.mode)
	}
}
func TestChatView_RevokeConsent(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	msg := tea.KeyMsg{Type: tea.KeyCtrlR}
	v, _ = v.Update(msg)
	if v.consented {
		t.Error("Expected consented to be false after Ctrl+R")
	}
}
func TestChatView_Escape_GoesBack(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	v, _ = v.Update(msg)
	if v.parent.mode != ViewBrowser {
		t.Errorf("Expected mode ViewBrowser after Esc, got %v", v.parent.mode)
	}
}
func TestChatView_FieldNavigation_Up(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.focusField = focusFieldMessage
	msg := tea.KeyMsg{Type: tea.KeyUp}
	v, _ = v.Update(msg)
	if v.focusField != focusFieldModel {
		t.Errorf("Expected focusField focusFieldModel after Up, got %d", v.focusField)
	}
	v, _ = v.Update(msg)
	if v.focusField != focusFieldProvider {
		t.Errorf("Expected focusField focusFieldProvider after second Up, got %d", v.focusField)
	}
	v, _ = v.Update(msg)
	if v.focusField != focusFieldProvider {
		t.Errorf("Expected focusField to stay focusFieldProvider, got %d", v.focusField)
	}
}
func TestChatView_FieldNavigation_Down(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.focusField = focusFieldProvider
	msg := tea.KeyMsg{Type: tea.KeyDown}
	v, _ = v.Update(msg)
	if v.focusField != focusFieldModel {
		t.Errorf("Expected focusField focusFieldModel after Down, got %d", v.focusField)
	}
	v, _ = v.Update(msg)
	if v.focusField != focusFieldMessage {
		t.Errorf("Expected focusField focusFieldMessage after second Down, got %d", v.focusField)
	}
	v, _ = v.Update(msg)
	if v.focusField != focusFieldMessage {
		t.Errorf("Expected focusField to stay focusFieldMessage, got %d", v.focusField)
	}
}
func TestChatView_ProviderSelection_LeftRight(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.focusField = focusFieldProvider
	if len(v.providers) < 2 {
		t.Skip("Need at least 2 providers to test navigation")
	}
	initialProvider := v.selectedProvider
	msg := tea.KeyMsg{Type: tea.KeyRight}
	v, _ = v.Update(msg)
	if v.selectedProvider == initialProvider {
	}
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	v, _ = v.Update(msg)
}
func TestChatView_ModelSelection_LeftRight(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.focusField = focusFieldModel
	v.models = []string{"model1", "model2", "model3"}
	v.selectedModel = 0
	msg := tea.KeyMsg{Type: tea.KeyRight}
	v, _ = v.Update(msg)
	if v.selectedModel != 1 {
		t.Errorf("Expected selectedModel 1 after Right, got %d", v.selectedModel)
	}
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	v, _ = v.Update(msg)
	if v.selectedModel != 0 {
		t.Errorf("Expected selectedModel 0 after Left, got %d", v.selectedModel)
	}
}
func TestChatView_ModelSelection_WrapAround(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.focusField = focusFieldModel
	v.models = []string{"model1", "model2"}
	v.selectedModel = 0
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	v, _ = v.Update(msg)
	if v.selectedModel != 1 {
		t.Errorf("Expected selectedModel 1 after wrap, got %d", v.selectedModel)
	}
	msg = tea.KeyMsg{Type: tea.KeyRight}
	v, _ = v.Update(msg)
	if v.selectedModel != 0 {
		t.Errorf("Expected selectedModel 0 after wrap, got %d", v.selectedModel)
	}
}
func TestChatView_MessageNavigation_CtrlP(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.messages = []chatMessage{
		{Role: "user", Content: "Hello"},
		{Role: "system", Content: "Hi"},
	}
	v.selectedMessage = -1
	msg := tea.KeyMsg{Type: tea.KeyCtrlP}
	v, _ = v.Update(msg)
	if v.selectedMessage != 1 {
		t.Errorf("Expected selectedMessage 1 after Ctrl+P, got %d", v.selectedMessage)
	}
	v, _ = v.Update(msg)
	if v.selectedMessage != 0 {
		t.Errorf("Expected selectedMessage 0 after second Ctrl+P, got %d", v.selectedMessage)
	}
}
func TestChatView_MessageNavigation_CtrlN(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.messages = []chatMessage{
		{Role: "user", Content: "Hello"},
		{Role: "system", Content: "Hi"},
	}
	v.selectedMessage = -1
	msg := tea.KeyMsg{Type: tea.KeyCtrlN}
	v, _ = v.Update(msg)
	if v.selectedMessage != 0 {
		t.Errorf("Expected selectedMessage 0 after Ctrl+N, got %d", v.selectedMessage)
	}
	v, _ = v.Update(msg)
	if v.selectedMessage != 1 {
		t.Errorf("Expected selectedMessage 1 after second Ctrl+N, got %d", v.selectedMessage)
	}
}
func TestChatView_FocusInput(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.focusField = focusFieldModel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	v, _ = v.Update(msg)
	if v.focusField != focusFieldMessage {
		t.Errorf("Expected focusField focusFieldMessage after '/', got %d", v.focusField)
	}
}
func TestChatView_FocusInput_Slash(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.focusField = focusFieldProvider
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	v, _ = v.Update(msg)
	if v.focusField != focusFieldMessage {
		t.Errorf("Expected focusField focusFieldMessage after '/', got %d", v.focusField)
	}
}
func TestChatView_MouseScroll(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.focusField = focusFieldMessage
	v.messages = make([]chatMessage, 20)
	v.scrollOffset = 5
	v.height = 30
	msg := tea.MouseMsg{Type: tea.MouseWheelUp}
	v, _ = v.Update(msg)
	if v.scrollOffset != 4 {
		t.Errorf("Expected scrollOffset 4 after wheel up, got %d", v.scrollOffset)
	}
	msg = tea.MouseMsg{Type: tea.MouseWheelDown}
	v, _ = v.Update(msg)
	if v.scrollOffset != 5 {
		t.Errorf("Expected scrollOffset 5 after wheel down, got %d", v.scrollOffset)
	}
}
func TestChatView_MouseScroll_WorksFromAnyField(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.messages = make([]chatMessage, 20)
	v.scrollOffset = 5
	v.height = 30
	v.focusField = focusFieldProvider
	msg := tea.MouseMsg{Type: tea.MouseWheelUp}
	v, _ = v.Update(msg)
	if v.scrollOffset != 4 {
		t.Errorf("Expected scrollOffset 4 after wheel up from provider field, got %d", v.scrollOffset)
	}
	v.focusField = focusFieldModel
	v, _ = v.Update(msg)
	if v.scrollOffset != 3 {
		t.Errorf("Expected scrollOffset 3 after wheel up from model field, got %d", v.scrollOffset)
	}
}
func TestChatView_MouseScrollUp_AtTop(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.scrollOffset = 0
	msg := tea.MouseMsg{Type: tea.MouseWheelUp}
	v, _ = v.Update(msg)
	if v.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset to stay 0 at top, got %d", v.scrollOffset)
	}
}
func TestChatView_EnsureMessageVisible(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.height = 30
	v.messages = make([]chatMessage, 10)
	v.scrollOffset = 0
	maxVisible := v.maxVisibleMessages()
	v.selectedMessage = maxVisible + 2
	v.ensureMessageVisible()
	if v.scrollOffset == 0 {
		t.Error("Expected scrollOffset to increase to show selected message")
	}
	if v.selectedMessage < v.scrollOffset || v.selectedMessage >= v.scrollOffset+maxVisible {
		t.Errorf("Selected message %d not visible in window [%d, %d)",
			v.selectedMessage, v.scrollOffset, v.scrollOffset+maxVisible)
	}
	v.selectedMessage = 0
	v.ensureMessageVisible()
	if v.scrollOffset != 0 {
		t.Errorf("Expected scrollOffset 0 to show message 0, got %d", v.scrollOffset)
	}
}
func TestChatView_CtrlP_AdjustsScroll(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.height = 30
	v.messages = make([]chatMessage, 10)
	maxVisible := v.maxVisibleMessages()
	v.scrollOffset = 5
	v.selectedMessage = 5
	msg := tea.KeyMsg{Type: tea.KeyCtrlP}
	for i := 0; i < maxVisible+2; i++ {
		v, _ = v.Update(msg)
	}
	if v.selectedMessage < v.scrollOffset || v.selectedMessage >= v.scrollOffset+maxVisible {
		t.Errorf("After Ctrl+P, selected message %d not in visible range [%d, %d)",
			v.selectedMessage, v.scrollOffset, v.scrollOffset+maxVisible)
	}
}
func TestChatView_WindowSizeMsg(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
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
func TestChatView_WindowSizeMsg_ConsentScreen(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = false
	msg := tea.WindowSizeMsg{Width: 100, Height: 30}
	v, _ = v.Update(msg)
	if v.width != 100 {
		t.Errorf("Expected width 100, got %d", v.width)
	}
}
func TestChatView_ModelsLoadedMsg(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.loadingModels = true
	msg := modelsLoadedMsg{models: []string{"gpt-4", "gpt-3.5"}, err: nil}
	v, _ = v.Update(msg)
	if v.loadingModels {
		t.Error("Expected loadingModels to be false after models loaded")
	}
	if len(v.models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(v.models))
	}
	if v.selectedModel != 0 {
		t.Errorf("Expected selectedModel 0, got %d", v.selectedModel)
	}
}
func TestChatView_ModelsLoadedMsg_Error(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.loadingModels = true
	msg := modelsLoadedMsg{models: nil, err: errors.New("failed to load")}
	v, _ = v.Update(msg)
	if v.loadingModels {
		t.Error("Expected loadingModels to be false after error")
	}
	if v.err == nil {
		t.Error("Expected error to be set")
	}
}
func TestChatView_ChatResponseMsg(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.sending = true
	msg := chatResponseMsg{
		messages: nil,
		err:      nil,
	}
	v, _ = v.Update(msg)
	if v.sending {
		t.Error("Expected sending to be false after response")
	}
}
func TestChatView_ChatResponseMsg_Error(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.sending = true
	initialMsgCount := len(v.messages)
	msg := chatResponseMsg{
		messages: nil,
		err:      errors.New("chat failed"),
	}
	v, _ = v.Update(msg)
	if v.sending {
		t.Error("Expected sending to be false after error")
	}
	if v.err == nil {
		t.Error("Expected error to be set")
	}
	if len(v.messages) != initialMsgCount+1 {
		t.Error("Expected error message to be added")
	}
}
func TestChatView_View_ConsentScreen(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = false
	view := v.View()
	if !strings.Contains(view, "AI Chat") {
		t.Error("Expected 'AI Chat' title")
	}
	if !strings.Contains(view, "accept") {
		t.Error("Expected 'accept' option on consent screen")
	}
	if !strings.Contains(view, "cancel") {
		t.Error("Expected 'cancel' option on consent screen")
	}
}
func TestChatView_View_ChatScreen(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	view := v.View()
	if !strings.Contains(view, "AI Chat") {
		t.Error("Expected 'AI Chat' title")
	}
	if !strings.Contains(view, "Provider:") {
		t.Error("Expected 'Provider:' label")
	}
	if !strings.Contains(view, "Model:") {
		t.Error("Expected 'Model:' label")
	}
	if !strings.Contains(view, "Message:") {
		t.Error("Expected 'Message:' label")
	}
}
func TestChatView_View_WithMessages(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.messages = []chatMessage{
		{Role: "user", Content: "Hello", Type: "message"},
		{Role: "system", Content: "Hi there", Type: "message"},
	}
	view := v.View()
	if !strings.Contains(view, "Conversation") {
		t.Error("Expected 'Conversation' subtitle")
	}
	if !strings.Contains(view, "You:") {
		t.Error("Expected 'You:' label for user message")
	}
	if !strings.Contains(view, "Assistant:") {
		t.Error("Expected 'Assistant:' label for system message")
	}
}
func TestChatView_View_WithError(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.err = errors.New("test error")
	v.sending = false
	view := v.View()
	if !strings.Contains(view, "test error") {
		t.Error("Expected error message in view")
	}
}
func TestChatView_View_Sending(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.sending = true
	view := v.View()
	if !strings.Contains(view, "Thinking...") {
		t.Error("Expected 'Thinking...' when sending")
	}
}
func TestChatView_View_LoadingModels(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.loadingModels = true
	view := v.View()
	if !strings.Contains(view, "Loading models...") {
		t.Error("Expected 'Loading models...' when loading")
	}
}
func TestChatView_View_NoModels(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.models = []string{}
	v.loadingModels = false
	view := v.View()
	if !strings.Contains(view, "Ctrl+L") || !strings.Contains(view, "load models") {
		t.Error("Expected hint to load models")
	}
}
func TestChatView_WrapText(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.width = 80
	result := v.wrapText("Hello world", 5)
	if result == "" {
		t.Error("Expected non-empty wrapped text")
	}
}
func TestChatView_WrapText_LongText(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.width = 50
	longText := strings.Repeat("This is a long line of text. ", 10)
	result := v.wrapText(longText, 5)
	lines := strings.Split(result, "\n")
	if len(lines) > 4 {
		t.Errorf("Expected max 4 lines, got %d", len(lines))
	}
}
func TestChatView_RenderTableSummary(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	result := v.renderTableSummary(nil)
	if !strings.Contains(result, "No results") {
		t.Error("Expected 'No results' for nil result")
	}
	result = v.renderTableSummary(&engine.GetRowsResult{
		Columns: []engine.Column{},
		Rows:    [][]string{},
	})
	if !strings.Contains(result, "No results") {
		t.Error("Expected 'No results' for empty columns")
	}
	result = v.renderTableSummary(&engine.GetRowsResult{
		Columns: []engine.Column{{Name: "id"}, {Name: "name"}},
		Rows:    [][]string{{"1", "Alice"}, {"2", "Bob"}},
	})
	if !strings.Contains(result, "2 rows") {
		t.Error("Expected row count in summary")
	}
	if !strings.Contains(result, "2 columns") {
		t.Error("Expected column count in summary")
	}
}
func TestChatView_Init_WithConsent(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	if len(v.providers) > 0 {
		cmd := v.Init()
		if cmd == nil {
			t.Error("Expected Init to return command when consented with providers")
		}
	}
}
func TestChatView_Init_WithoutConsent(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = false
	cmd := v.Init()
	if cmd != nil {
		t.Error("Expected Init to return nil when not consented")
	}
}
func TestChatMessage_Fields(t *testing.T) {
	msg := chatMessage{
		Role:    "user",
		Content: "Hello",
		Type:    "message",
		Result:  nil,
	}
	if msg.Role != "user" {
		t.Errorf("Expected Role 'user', got '%s'", msg.Role)
	}
	if msg.Content != "Hello" {
		t.Errorf("Expected Content 'Hello', got '%s'", msg.Content)
	}
	if msg.Type != "message" {
		t.Errorf("Expected Type 'message', got '%s'", msg.Type)
	}
}
func TestFocusField_Constants(t *testing.T) {
	if focusFieldProvider >= focusFieldModel {
		t.Error("Expected focusFieldProvider < focusFieldModel")
	}
	if focusFieldModel >= focusFieldMessage {
		t.Error("Expected focusFieldModel < focusFieldMessage")
	}
}
func TestChatView_RetryPrompt_EscCancels(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.retryPrompt.Show("tell me about the users table")
	v.err = errors.New("request timed out")
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	v, _ = v.Update(msg)
	if v.retryPrompt.IsActive() {
		t.Error("Expected retryPrompt to be false after ESC")
	}
	if v.retryPrompt.TimedOutQuery() != "" {
		t.Errorf("Expected timedOutQuery to be empty, got '%s'", v.retryPrompt.TimedOutQuery())
	}
}
func TestChatView_RetryPrompt_KeyHandling(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"option_1", "1"},
		{"option_2", "2"},
		{"option_3", "3"},
		{"option_4", "4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, cleanup := setupChatViewTest(t)
			defer cleanup()
			v.consented = true
			v.retryPrompt.Show("tell me about the users table")
			v.err = errors.New("request timed out")
			v.models = []string{"test-model"}
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			v, cmd := v.Update(msg)
			if v.retryPrompt.IsActive() {
				t.Error("Expected retryPrompt to be false after selecting retry option")
			}
			if v.err != nil {
				t.Error("Expected err to be nil after retry")
			}
			if cmd == nil {
				t.Error("Expected a command to be returned for retry")
			}
		})
	}
}
func TestChatView_RetryPrompt_IgnoresOtherKeys(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.retryPrompt.Show("tell me about the users table")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}
	v, _ = v.Update(msg)
	if !v.retryPrompt.IsActive() {
		t.Error("Expected retryPrompt to still be true after unrecognized key")
	}
	if v.retryPrompt.TimedOutQuery() == "" {
		t.Error("Expected timedOutQuery to still be set")
	}
}
func TestChatView_RestoreProvider(t *testing.T) {
	setupTestEnv(t)
	cfg, _ := config.LoadConfig()
	cfg.SetLastAIProvider("OpenAI")
	cfg.Save()
	parent := NewMainModel()
	if parent.err != nil {
		t.Fatalf("Failed to create MainModel: %v", parent.err)
	}
	v := parent.chatView
	if len(v.providers) > 0 {
		found := false
		for i, p := range v.providers {
			if p.Type == "OpenAI" && v.selectedProvider == i {
				found = true
				break
			}
		}
		if !found {
			if v.selectedProvider < 0 || v.selectedProvider >= len(v.providers) {
				t.Errorf("selectedProvider out of range: %d", v.selectedProvider)
			}
		}
	}
}
func TestChatView_RestoreModel_AutoFocusMessage(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.parent.config.SetLastAIModel("test-model")
	v.parent.config.Save()
	v.consented = true
	v.loadingModels = true
	v.focusField = focusFieldProvider
	msg := modelsLoadedMsg{models: []string{"other-model", "test-model", "third-model"}, err: nil}
	v, _ = v.Update(msg)
	if v.selectedModel != 1 {
		t.Errorf("Expected selectedModel 1 (test-model), got %d", v.selectedModel)
	}
	if v.focusField != focusFieldMessage {
		t.Errorf("Expected focusField focusFieldMessage after model restore, got %d", v.focusField)
	}
}
func TestChatView_RestoreModel_NotFound_FallsBack(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.parent.config.SetLastAIModel("nonexistent-model")
	v.parent.config.Save()
	v.consented = true
	v.loadingModels = true
	msg := modelsLoadedMsg{models: []string{"model-a", "model-b"}, err: nil}
	v, _ = v.Update(msg)
	if v.selectedModel != 0 {
		t.Errorf("Expected selectedModel 0 (fallback), got %d", v.selectedModel)
	}
}
func TestChatView_SavesProviderModelOnSuccess(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.sending = true
	v.providers = v.parent.dbManager.GetAIProviders()
	if len(v.providers) == 0 {
		t.Skip("No providers available")
	}
	v.selectedProvider = 0
	v.models = []string{"test-model-a", "test-model-b"}
	v.selectedModel = 1
	msg := chatResponseMsg{
		messages: []*database.ChatMessage{{Text: "Hello!", Type: "message"}},
		err:      nil,
	}
	v, _ = v.Update(msg)
	if v.parent.config.GetLastAIProvider() != v.providers[0].Type {
		t.Errorf("Expected saved provider '%s', got '%s'", v.providers[0].Type, v.parent.config.GetLastAIProvider())
	}
	if v.parent.config.GetLastAIModel() != "test-model-b" {
		t.Errorf("Expected saved model 'test-model-b', got '%s'", v.parent.config.GetLastAIModel())
	}
}
func TestChatView_TimeoutAutoRetry_WithPreference(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.sending = true
	v.models = []string{"test-model"}
	v.retryPrompt.SetAutoRetried(false)
	v.parent.config.SetPreferredTimeout(60)
	msg := chatResponseMsg{
		messages: nil,
		query:    "test query",
		err:      context.DeadlineExceeded,
	}
	v, cmd := v.Update(msg)
	if v.retryPrompt.IsActive() {
		t.Error("Expected retryPrompt to be false (auto-retry should happen)")
	}
	if !v.retryPrompt.AutoRetried() {
		t.Error("Expected autoRetried to be true")
	}
	if cmd == nil {
		t.Error("Expected a command to be returned for auto-retry")
	}
}
func TestChatView_TimeoutShowsMenu_AfterAutoRetry(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.sending = true
	v.retryPrompt.SetAutoRetried(true)
	v.parent.config.SetPreferredTimeout(60)
	msg := chatResponseMsg{
		messages: nil,
		query:    "test query",
		err:      context.DeadlineExceeded,
	}
	v, _ = v.Update(msg)
	if !v.retryPrompt.IsActive() {
		t.Error("Expected retryPrompt to be true after auto-retry failed")
	}
}
func TestChatView_TimeoutShowsMenu_NoPreference(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.sending = true
	v.retryPrompt.SetAutoRetried(false)
	v.parent.config.SetPreferredTimeout(0)
	msg := chatResponseMsg{
		messages: nil,
		query:    "test query",
		err:      context.DeadlineExceeded,
	}
	v, _ = v.Update(msg)
	if !v.retryPrompt.IsActive() {
		t.Error("Expected retryPrompt to be true with no preferred timeout")
	}
}
func TestChatView_RetryMenuSavesPreference(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		expectedTimeout int
	}{
		{"option_1_saves_60", "1", 60},
		{"option_2_saves_120", "2", 120},
		{"option_3_saves_300", "3", 300},
		{"option_4_no_save", "4", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, cleanup := setupChatViewTest(t)
			defer cleanup()
			v.consented = true
			v.retryPrompt.Show("test query")
			v.models = []string{"test-model"}
			v.parent.config.SetPreferredTimeout(0)
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			v, _ = v.Update(msg)
			saved := v.parent.config.GetPreferredTimeout()
			if saved != tt.expectedTimeout {
				t.Errorf("Expected preferred timeout %d after key '%s', got %d", tt.expectedTimeout, tt.key, saved)
			}
		})
	}
}
func TestChatView_AutoRetriedResetOnNewMessage(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.retryPrompt.SetAutoRetried(true)
	v.focusField = focusFieldMessage
	v.models = []string{"test-model"}
	v.input.SetValue("new question")
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	v, _ = v.Update(msg)
	if v.retryPrompt.AutoRetried() {
		t.Error("Expected autoRetried to be reset on new message")
	}
}
func TestChatView_RetryPrompt_View(t *testing.T) {
	v, cleanup := setupChatViewTest(t)
	defer cleanup()
	v.consented = true
	v.retryPrompt.Show("tell me about the users table")
	view := v.View()
	if !strings.Contains(view, "timed out") {
		t.Error("Expected 'timed out' in view")
	}
	if !strings.Contains(view, "60 seconds") {
		t.Error("Expected '60 seconds' option in view")
	}
	if !strings.Contains(view, "2 minutes") {
		t.Error("Expected '2 minutes' option in view")
	}
	if !strings.Contains(view, "5 minutes") {
		t.Error("Expected '5 minutes' option in view")
	}
	if !strings.Contains(view, "No limit") {
		t.Error("Expected 'No limit' option in view")
	}
}