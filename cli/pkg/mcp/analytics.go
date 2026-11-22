package mcp
import (
	"context"
	"os"
	"github.com/clidey/whodb/cli/pkg/analytics"
)
type AnalyticsConfig struct {
	Enabled bool
	AppVersion string
}
func InitializeAnalytics(cfg *AnalyticsConfig) error {
	if cfg == nil {
		cfg = &AnalyticsConfig{Enabled: true}
	}
	if os.Getenv("WHODB_MCP_ANALYTICS_DISABLED") == "true" {
		cfg.Enabled = false
	}
	if err := analytics.Initialize(cfg.AppVersion); err != nil {
		return err
	}
	if !cfg.Enabled {
		analytics.SetEnabled(false)
	}
	return nil
}
func ShutdownAnalytics() {
	analytics.Shutdown()
}
func IsAnalyticsEnabled() bool {
	return analytics.IsEnabled()
}
func TrackToolCall(ctx context.Context, toolName, requestID string, success bool, durationMs int64, props map[string]any) {
	analytics.TrackMCPToolCall(ctx, toolName, requestID, success, durationMs, props)
}
func TrackServerStart(ctx context.Context, transport string, securityMode string, props map[string]any) {
	analytics.TrackMCPServerStart(ctx, transport, securityMode, props)
}
func TrackError(ctx context.Context, toolName, requestID, operation string, errMsg string) {
	analytics.TrackError(ctx, operation, &mcpError{msg: errMsg}, map[string]any{
		"tool_name":  toolName,
		"request_id": requestID,
	})
}
type mcpError struct {
	msg string
}
func (e *mcpError) Error() string {
	return e.msg
}