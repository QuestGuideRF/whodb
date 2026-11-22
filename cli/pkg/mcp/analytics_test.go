package mcp
import (
	"context"
	"os"
	"testing"
)
func TestAnalyticsConfig_DisabledByEnv(t *testing.T) {
	os.Setenv("WHODB_MCP_ANALYTICS_DISABLED", "true")
	defer os.Unsetenv("WHODB_MCP_ANALYTICS_DISABLED")
	cfg := &AnalyticsConfig{Enabled: true}
	err := InitializeAnalytics(cfg)
	if err != nil {
	}
	ShutdownAnalytics()
}
func TestAnalyticsConfig_DisabledByFlag(t *testing.T) {
	cfg := &AnalyticsConfig{Enabled: false}
	err := InitializeAnalytics(cfg)
	if err != nil {
	}
	ShutdownAnalytics()
}
func TestTrackToolCall_DoesNotPanicWhenDisabled(t *testing.T) {
	ctx := context.Background()
	TrackToolCall(ctx, "test_tool", "test-123", true, 100, map[string]any{"key": "value"})
}
func TestTrackServerStart_DoesNotPanicWhenDisabled(t *testing.T) {
	ctx := context.Background()
	TrackServerStart(ctx, "stdio", "confirm-writes", map[string]any{"key": "value"})
}
func TestTrackError_DoesNotPanicWhenDisabled(t *testing.T) {
	ctx := context.Background()
	TrackError(ctx, "test_tool", "test-123", "test_operation", "test error message")
}
func TestMcpError_Error(t *testing.T) {
	err := &mcpError{msg: "test error"}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Error())
	}
}