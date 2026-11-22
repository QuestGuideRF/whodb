package analytics
import (
	"context"
	"os"
	"sync"
	"time"
	"github.com/clidey/whodb/cli/pkg/identity"
	"github.com/clidey/whodb/core/src/analytics"
)
const (
	posthogAPIKey = "phc_hbXcCoPTdxm5ADL8PmLSYTIUvS6oRWFM2JAK8SMbfnH"
	posthogHost   = "https://us.i.posthog.com"
)
var (
	initialized bool
	initMu      sync.Mutex
	cliVersion  string
)
func Initialize(version string) error {
	initMu.Lock()
	defer initMu.Unlock()
	if initialized {
		return nil
	}
	cliVersion = version
	if os.Getenv(identity.Current().AnalyticsDisabledEnv) == "true" {
		return nil
	}
	if err := analytics.Initialize(analytics.Config{
		APIKey:             posthogAPIKey,
		Host:               posthogHost,
		Environment:        "cli",
		AppVersion:         version,
		SuppressClientLogs: true,
	}); err != nil {
		return err
	}
	analytics.SetEnabled(true)
	initialized = true
	return nil
}
func Shutdown() {
	analytics.Shutdown()
}
func SetEnabled(enabled bool) {
	analytics.SetEnabled(enabled)
}
func IsEnabled() bool {
	return analytics.Enabled()
}
func baseProps() map[string]any {
	return map[string]any{
		"source":      "cli",
		"cli_version": cliVersion,
		"edition":     identity.Current().Edition,
	}
}
func mergeProps(additional map[string]any) map[string]any {
	props := baseProps()
	for k, v := range additional {
		props[k] = v
	}
	return props
}
func TrackCLIStartup(ctx context.Context, command string) {
	analytics.Capture(ctx, "cli.startup", mergeProps(map[string]any{
		"command":   command,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}))
}
func TrackSessionStart(ctx context.Context, dbType string) {
	analytics.Capture(ctx, "cli.session.start", mergeProps(map[string]any{
		"db_type":   dbType,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}))
}
func TrackSessionEnd(ctx context.Context, dbType string, durationSec int64, queryCount int) {
	analytics.Capture(ctx, "cli.session.end", mergeProps(map[string]any{
		"db_type":      dbType,
		"duration_sec": durationSec,
		"query_count":  queryCount,
	}))
}
func TrackConnectAttempt(ctx context.Context, dbType string) {
	analytics.Capture(ctx, "cli.connect.attempt", mergeProps(map[string]any{
		"db_type": dbType,
	}))
}
func TrackConnectSuccess(ctx context.Context, dbType string, durationMs int64) {
	analytics.Capture(ctx, "cli.connect.success", mergeProps(map[string]any{
		"db_type":     dbType,
		"duration_ms": durationMs,
	}))
}
func TrackConnectError(ctx context.Context, dbType string, errorType string, durationMs int64) {
	analytics.Capture(ctx, "cli.connect.error", mergeProps(map[string]any{
		"db_type":     dbType,
		"error_type":  errorType,
		"duration_ms": durationMs,
	}))
}
func TrackConnectionAdd(ctx context.Context, dbType string) {
	analytics.Capture(ctx, "cli.connections.add", mergeProps(map[string]any{
		"db_type": dbType,
	}))
}
func TrackConnectionRemove(ctx context.Context) {
	analytics.Capture(ctx, "cli.connections.remove", mergeProps(nil))
}
func TrackConnectionTest(ctx context.Context, dbType string, success bool, durationMs int64) {
	analytics.Capture(ctx, "cli.connections.test", mergeProps(map[string]any{
		"db_type":     dbType,
		"success":     success,
		"duration_ms": durationMs,
	}))
}
func TrackQueryExecute(ctx context.Context, dbType string, statementType string, success bool, durationMs int64, rowCount int, props map[string]any) {
	eventProps := mergeProps(map[string]any{
		"db_type":        dbType,
		"statement_type": statementType,
		"success":        success,
		"duration_ms":    durationMs,
		"row_count":      rowCount,
	})
	for k, v := range props {
		eventProps[k] = v
	}
	analytics.Capture(ctx, "cli.query.execute", eventProps)
}
func TrackQueryError(ctx context.Context, dbType string, errorType string, durationMs int64) {
	analytics.Capture(ctx, "cli.query.error", mergeProps(map[string]any{
		"db_type":     dbType,
		"error_type":  errorType,
		"duration_ms": durationMs,
	}))
}
func TrackSchemasListed(ctx context.Context, dbType string, schemaCount int, durationMs int64) {
	analytics.Capture(ctx, "cli.schemas.list", mergeProps(map[string]any{
		"db_type":      dbType,
		"schema_count": schemaCount,
		"duration_ms":  durationMs,
	}))
}
func TrackTablesListed(ctx context.Context, dbType string, tableCount int, durationMs int64) {
	analytics.Capture(ctx, "cli.tables.list", mergeProps(map[string]any{
		"db_type":     dbType,
		"table_count": tableCount,
		"duration_ms": durationMs,
	}))
}
func TrackColumnsListed(ctx context.Context, dbType string, columnCount int, durationMs int64) {
	analytics.Capture(ctx, "cli.columns.list", mergeProps(map[string]any{
		"db_type":      dbType,
		"column_count": columnCount,
		"duration_ms":  durationMs,
	}))
}
func TrackExport(ctx context.Context, dbType string, format string, rowCount int, durationMs int64) {
	analytics.Capture(ctx, "cli.export.execute", mergeProps(map[string]any{
		"db_type":     dbType,
		"format":      format,
		"row_count":   rowCount,
		"duration_ms": durationMs,
	}))
}
func TrackMCPServerStart(ctx context.Context, transport string, securityMode string, props map[string]any) {
	eventProps := mergeProps(map[string]any{
		"transport":     transport,
		"security_mode": securityMode,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	})
	for k, v := range props {
		eventProps[k] = v
	}
	analytics.Capture(ctx, "mcp.server.start", eventProps)
}
func TrackMCPToolCall(ctx context.Context, toolName string, requestID string, success bool, durationMs int64, props map[string]any) {
	eventProps := mergeProps(map[string]any{
		"tool_name":   toolName,
		"request_id":  requestID,
		"success":     success,
		"duration_ms": durationMs,
	})
	for k, v := range props {
		eventProps[k] = v
	}
	analytics.Capture(ctx, "mcp.tool.call", eventProps)
}
func TrackError(ctx context.Context, operation string, err error, props map[string]any) {
	analytics.CaptureError(ctx, operation, err, mergeProps(props))
}