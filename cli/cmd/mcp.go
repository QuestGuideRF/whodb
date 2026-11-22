package cmd
import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	whodbmcp "github.com/clidey/whodb/cli/pkg/mcp"
	"github.com/clidey/whodb/cli/pkg/version"
	"github.com/spf13/cobra"
)
var (
	mcpReadOnly            bool
	mcpConfirmWrites       bool
	mcpAllowWrite          bool
	mcpAllowDrop           bool
	mcpSecurity            string
	mcpTimeout             time.Duration
	mcpMaxRows             int
	mcpAllowMultiStatement bool
	mcpSafeMode            bool
)
var (
	mcpTransport string
	mcpHost      string
	mcpPort      int
)
var (
	mcpEnabledTools  []string
	mcpDisabledTools []string
)
var mcpConnection string
var mcpAllowedConnections []string
var mcpNoAnalytics bool
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Model Context Protocol server",
	Long: `Run WhoDB as an MCP (Model Context Protocol) server.
MCP enables AI assistants like Claude to interact with your databases
through a standardized protocol.`,
}
var mcpServeCmd = &cobra.Command{
	Use:           "serve",
	Short:         "Start the MCP server",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `Start WhoDB as an MCP server.
TRANSPORT:
  --transport stdio  (default) Communicate via stdin/stdout for CLI integration
  --transport http   Run as HTTP service for cloud/shared deployments
  HTTP mode options:
    --host HOST      Bind address (default: localhost)
    --port PORT      Listen port (default: 3000)
  HTTP mode exposes:
    /mcp      - MCP endpoint (streaming HTTP)
    /health   - Health check endpoint (returns {"status":"ok"})
SECURITY:
  Write operations require user confirmation by default. This keeps you in
  control while allowing full database functionality.
  Permission Modes (controls whether writes are allowed):
    (default)        - Confirm-writes: All writes require user confirmation
    --safe-mode      - Safe mode: read-only + strict security
    --read-only      - Read-only: SELECT, SHOW, DESCRIBE, EXPLAIN only
    --allow-write    - Full write access without confirmation (use with caution)
  Security Levels (additional validation, does NOT override permission mode):
    strict   - Blocks dangerous functions (pg_read_file, COPY, LOAD_FILE, etc.)
    standard - Basic validation (default)
    minimal  - Only blocks DELETE without WHERE (when writes allowed)
  Note: Permission mode takes priority. --read-only blocks all writes regardless of
  security level. Multi-statement queries are blocked by default (--allow-multi-statement).
Available tools:
  whodb_query       - Execute SQL queries (security-validated)
  whodb_schemas     - List database schemas
  whodb_tables      - List tables in a schema
  whodb_columns     - Describe table columns
  whodb_connections - List available connections
  whodb_explain     - Run EXPLAIN for a query
  whodb_diff        - Compare schema metadata between two connections
  whodb_erd         - Load graph/relationship metadata
  whodb_audit       - Run data quality audits
  whodb_suggestions - Load backend query suggestions
  whodb_confirm     - Confirm pending writes (only with --confirm-writes)
  whodb_pending     - List pending confirmation tokens
TOOL SELECTION:
  --tools           - Comma-separated list of tools to enable (default: all)
                      Valid: query, schemas, tables, columns, connections, confirm, pending, explain, diff, erd, audit, suggestions
  --disable-tools   - Comma-separated list of tools to disable (takes precedence)
ANALYTICS:
  Anonymous usage analytics are enabled by default to help improve WhoDB.
  No query content, database credentials, or personal data is ever collected.
  Only tool usage patterns and error rates are tracked.
  To disable analytics:
    --no-analytics                        Flag to disable analytics
    WHODB_MCP_ANALYTICS_DISABLED=true     Environment variable to disable
CONNECTION SCOPING:
  --default-connection NAME     Set default connection (no access restriction)
  --allowed-connections A,B,C   Restrict access to listed connections only
  With --allowed-connections:
  - whodb_connections only shows allowed connections
  - Queries to other connections are rejected
  - First allowed connection becomes the default (unless --default-connection set)
  With --default-connection only:
  - Sets the default, but all connections remain accessible
Connection Resolution:
  Tools accept a 'connection' parameter that references either:
  1. Environment profiles, for example:
     - WHODB_POSTGRES='[{"alias":"prod","host":"localhost","user":"user","password":"pass","database":"db","port":"5432"}]'
     - WHODB_MYSQL_1='{"alias":"staging","host":"localhost","user":"user","password":"pass","database":"db","port":"3306"}'
  2. Saved connection from 'whodb-cli connections add'
  Saved connections take precedence when names collide.`,
	Example: `  # Start MCP server (confirm-writes by default - you approve each write)
  whodb-cli mcp serve
  # Safe mode for demos/playgrounds (read-only + strict security)
  whodb-cli mcp serve --safe-mode
  # Read-only mode (no writes at all)
  whodb-cli mcp serve --read-only
  # Allow full write access without confirmation (use with caution)
  whodb-cli mcp serve --allow-write
  # Strict security mode (blocks dangerous functions like pg_read_file)
  whodb-cli mcp serve --security=strict
  # Custom timeout and row limit
  whodb-cli mcp serve --timeout=60s --max-rows=500
  # Run as HTTP service
  whodb-cli mcp serve --transport=http --port=3000
  # Endpoint: http:
  # Health:   http:
  # HTTP mode bound to all interfaces (can be used in Docker/Kubernetes)
  whodb-cli mcp serve --transport=http --host=0.0.0.0 --port=8080
  # Enable only specific tools (minimal surface for read-only exploration)
  whodb-cli mcp serve --tools=schemas,tables,columns,connections,suggestions
  # Disable query tool (only schema exploration)
  whodb-cli mcp serve --disable-tools=query,confirm
  # Restrict to specific connections (first becomes default)
  whodb-cli mcp serve --allowed-connections=prod,staging
  # Set default connection without restricting access
  whodb-cli mcp serve --default-connection=prod
  # Claude Desktop / Claude Code configuration (stdio):
  {
    "mcpServers": {
      "whodb": {
        "command": "whodb-cli",
        "args": ["mcp", "serve"],
        "env": {
          "WHODB_POSTGRES_1": "{\"alias\":\"prod\",\"host\":\"localhost\",\"user\":\"user\",\"password\":\"pass\",\"database\":\"db\"}"
        }
      }
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()
		if err := whodbmcp.InitializeAnalytics(&whodbmcp.AnalyticsConfig{
			Enabled:    !mcpNoAnalytics,
			AppVersion: version.Version,
		}); err != nil {
		}
		defer whodbmcp.ShutdownAnalytics()
		readOnly := false
		confirmWrites := true
		securityLevel := mcpSecurity
		if mcpSafeMode {
			readOnly = true
			confirmWrites = false
			securityLevel = "strict"
		} else if mcpAllowWrite {
			confirmWrites = false
		} else if mcpReadOnly {
			readOnly = true
			confirmWrites = false
		}
		opts := &whodbmcp.ServerOptions{
			ReadOnly:            readOnly,
			ConfirmWrites:       confirmWrites,
			SecurityLevel:       whodbmcp.SecurityLevel(securityLevel),
			QueryTimeout:        mcpTimeout,
			MaxRows:             mcpMaxRows,
			AllowMultiStatement: mcpAllowMultiStatement,
			AllowDrop:           mcpAllowDrop,
			EnabledTools:        mcpEnabledTools,
			DisabledTools:       mcpDisabledTools,
			DefaultConnection:   mcpConnection,
			AllowedConnections:  mcpAllowedConnections,
		}
		server := whodbmcp.NewServer(opts)
		securityModeName := "confirm-writes"
		if mcpSafeMode {
			securityModeName = "safe-mode"
		} else if mcpReadOnly {
			securityModeName = "read-only"
		} else if mcpAllowWrite {
			securityModeName = "allow-write"
		}
		whodbmcp.TrackServerStart(ctx, mcpTransport, securityModeName, map[string]any{
			"enabled_tools":  mcpEnabledTools,
			"disabled_tools": mcpDisabledTools,
			"security_level": securityLevel,
		})
		switch whodbmcp.TransportType(mcpTransport) {
		case whodbmcp.TransportHTTP:
			return whodbmcp.RunHTTP(ctx, server, &whodbmcp.HTTPOptions{
				Host: mcpHost,
				Port: mcpPort,
			}, opts.Logger)
		default:
			return whodbmcp.Run(ctx, server)
		}
	},
}
func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpServeCmd)
	mcpServeCmd.Flags().BoolVar(&mcpSafeMode, "safe-mode", false,
		"Enable safe mode (read-only + strict security) for demos and playgrounds")
	mcpServeCmd.Flags().BoolVar(&mcpReadOnly, "read-only", false,
		"Enable read-only mode (blocks all write operations)")
	mcpServeCmd.Flags().BoolVar(&mcpConfirmWrites, "confirm-writes", false,
		"Enable human-in-the-loop write confirmation (this is the default)")
	mcpServeCmd.Flags().BoolVar(&mcpAllowWrite, "allow-write", false,
		"Allow all write operations without confirmation (use with caution)")
	mcpServeCmd.Flags().BoolVar(&mcpAllowDrop, "allow-drop", false,
		"Allow DROP/TRUNCATE even with --allow-write (requires explicit opt-in)")
	mcpServeCmd.Flags().StringVar(&mcpSecurity, "security", "standard",
		"Security level: strict, standard, or minimal")
	mcpServeCmd.Flags().DurationVar(&mcpTimeout, "timeout", 30*time.Second,
		"Query timeout duration")
	mcpServeCmd.Flags().IntVar(&mcpMaxRows, "max-rows", 0,
		"Limit rows returned per query (0 = unlimited, default)")
	mcpServeCmd.Flags().BoolVar(&mcpAllowMultiStatement, "allow-multi-statement", false,
		"Allow multiple SQL statements in one query (security risk)")
	mcpServeCmd.Flags().StringVar(&mcpTransport, "transport", "stdio",
		"Transport type: stdio (default) or http")
	mcpServeCmd.Flags().StringVar(&mcpHost, "host", "localhost",
		"Host to bind to (only used with --transport=http)")
	mcpServeCmd.Flags().IntVar(&mcpPort, "port", 3000,
		"Port to listen on (only used with --transport=http)")
	mcpServeCmd.Flags().StringSliceVar(&mcpEnabledTools, "tools", nil,
		"Comma-separated list of tools to enable (default: all). Valid: query, schemas, tables, columns, connections, confirm")
	mcpServeCmd.Flags().StringSliceVar(&mcpDisabledTools, "disable-tools", nil,
		"Comma-separated list of tools to disable (takes precedence over --tools)")
	mcpServeCmd.Flags().StringVar(&mcpConnection, "default-connection", "",
		"Default connection to use when not specified (does not restrict access)")
	mcpServeCmd.Flags().StringSliceVar(&mcpAllowedConnections, "allowed-connections", nil,
		"Comma-separated list of connections to allow (restricts access, first one becomes default)")
	mcpServeCmd.Flags().BoolVar(&mcpNoAnalytics, "no-analytics", false,
		"Disable anonymous usage analytics (can also set WHODB_MCP_ANALYTICS_DISABLED=true)")
	mcpServeCmd.MarkFlagsMutuallyExclusive("read-only", "allow-write")
	mcpServeCmd.RegisterFlagCompletionFunc("default-connection", completeConnectionNames)
	mcpServeCmd.RegisterFlagCompletionFunc("allowed-connections", completeConnectionNames)
	mcpServeCmd.RegisterFlagCompletionFunc("transport", completeMCPTransports)
	mcpServeCmd.RegisterFlagCompletionFunc("security", completeMCPSecurityLevels)
}