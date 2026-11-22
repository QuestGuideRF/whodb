package analytics
import (
	"context"
	"errors"
	"os"
	"testing"
)
func TestInitialize_DisabledByEnv(t *testing.T) {
	os.Setenv("WHODB_CLI_ANALYTICS_DISABLED", "true")
	defer os.Unsetenv("WHODB_CLI_ANALYTICS_DISABLED")
	err := Initialize("test-version")
	if err != nil {
		t.Errorf("Initialize should not error when disabled: %v", err)
	}
	if IsEnabled() {
		t.Error("Analytics should be disabled when WHODB_CLI_ANALYTICS_DISABLED=true")
	}
	Shutdown()
}
func TestInitialize_MultipleCallsSafe(t *testing.T) {
	initMu.Lock()
	initialized = false
	initMu.Unlock()
	err1 := Initialize("v1")
	err2 := Initialize("v2")
	if err1 != nil {
		t.Errorf("First Initialize should not error: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second Initialize should not error: %v", err2)
	}
	Shutdown()
}
func TestBaseProps(t *testing.T) {
	cliVersion = "test-1.0.0"
	props := baseProps()
	if props["source"] != "cli" {
		t.Errorf("Expected source='cli', got %v", props["source"])
	}
	if props["cli_version"] != "test-1.0.0" {
		t.Errorf("Expected cli_version='test-1.0.0', got %v", props["cli_version"])
	}
}
func TestMergeProps(t *testing.T) {
	cliVersion = "test-1.0.0"
	additional := map[string]any{
		"custom_key": "custom_value",
		"number":     42,
	}
	merged := mergeProps(additional)
	if merged["source"] != "cli" {
		t.Errorf("Expected source='cli', got %v", merged["source"])
	}
	if merged["custom_key"] != "custom_value" {
		t.Errorf("Expected custom_key='custom_value', got %v", merged["custom_key"])
	}
	if merged["number"] != 42 {
		t.Errorf("Expected number=42, got %v", merged["number"])
	}
}
func TestMergeProps_NilInput(t *testing.T) {
	cliVersion = "test-1.0.0"
	merged := mergeProps(nil)
	if merged["source"] != "cli" {
		t.Errorf("Expected source='cli', got %v", merged["source"])
	}
}
func TestTrackFunctions_NoPanicWhenUninitialized(t *testing.T) {
	ctx := context.Background()
	TrackCLIStartup(ctx, "test")
	TrackSessionStart(ctx, "postgres")
	TrackSessionEnd(ctx, "postgres", 100, 5)
	TrackConnectAttempt(ctx, "postgres")
	TrackConnectSuccess(ctx, "postgres", 100)
	TrackConnectError(ctx, "postgres", "timeout", 100)
	TrackConnectionAdd(ctx, "postgres")
	TrackConnectionRemove(ctx)
	TrackConnectionTest(ctx, "postgres", true, 100)
	TrackQueryExecute(ctx, "postgres", "SELECT", true, 100, 10, nil)
	TrackQueryError(ctx, "postgres", "syntax_error", 100)
	TrackSchemasListed(ctx, "postgres", 5, 100)
	TrackTablesListed(ctx, "postgres", 10, 100)
	TrackColumnsListed(ctx, "postgres", 8, 100)
	TrackExport(ctx, "postgres", "csv", 100, 500)
	TrackMCPServerStart(ctx, "stdio", "safe", nil)
	TrackMCPToolCall(ctx, "query", "req-123", true, 100, nil)
	TrackError(ctx, "test_op", errors.New("test error"), nil)
}
func TestSetEnabled(t *testing.T) {
	SetEnabled(false)
	if IsEnabled() {
		t.Error("Analytics should be disabled after SetEnabled(false)")
	}
	SetEnabled(true)
}