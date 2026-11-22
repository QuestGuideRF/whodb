package mcp
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/clidey/whodb/cli/pkg/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)
type ServerOptions struct {
	Logger *slog.Logger
	Instructions string
	ReadOnly bool
	ConfirmWrites bool
	SecurityLevel SecurityLevel
	QueryTimeout time.Duration
	MaxRows int
	AllowMultiStatement bool
	AllowDrop bool
	EnabledTools []string
	DisabledTools []string
	DefaultConnection string
	AllowedConnections []string
}
type SecurityOptions struct {
	ReadOnly            bool
	ConfirmWrites       bool
	SecurityLevel       SecurityLevel
	QueryTimeout        time.Duration
	MaxRows             int
	AllowMultiStatement bool
	AllowDrop           bool
	DefaultConnection   string
	AllowedConnections  []string
}
func mapWhoDBLogLevel() slog.Level {
	switch strings.ToLower(os.Getenv("WHODB_LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warning", "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "none", "off", "disabled":
		return slog.Level(12)
	default:
		return slog.LevelInfo
	}
}
func NewServer(opts *ServerOptions) *mcp.Server {
	if opts == nil {
		opts = &ServerOptions{ConfirmWrites: true}
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: mapWhoDBLogLevel(),
		}))
	}
	if opts.Instructions == "" {
		opts.Instructions = defaultInstructions
	}
	if opts.SecurityLevel == "" {
		opts.SecurityLevel = SecurityLevelStandard
	}
	if opts.QueryTimeout == 0 {
		opts.QueryTimeout = 30 * time.Second
	}
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "whodb",
		Version: version.Version,
	}, &mcp.ServerOptions{
		Instructions: opts.Instructions,
		Logger:       opts.Logger,
	})
	defaultConn := opts.DefaultConnection
	if defaultConn == "" && len(opts.AllowedConnections) > 0 {
		defaultConn = opts.AllowedConnections[0]
	}
	secOpts := &SecurityOptions{
		ReadOnly:            opts.ReadOnly,
		ConfirmWrites:       opts.ConfirmWrites,
		SecurityLevel:       opts.SecurityLevel,
		QueryTimeout:        opts.QueryTimeout,
		MaxRows:             opts.MaxRows,
		AllowMultiStatement: opts.AllowMultiStatement,
		AllowDrop:           opts.AllowDrop,
		DefaultConnection:   defaultConn,
		AllowedConnections:  opts.AllowedConnections,
	}
	toolEnablement := &ToolEnablement{
		EnabledTools:  opts.EnabledTools,
		DisabledTools: opts.DisabledTools,
	}
	registerTools(server, secOpts, toolEnablement)
	registerPrompts(server)
	registerResources(server)
	return server
}
type ToolEnablement struct {
	EnabledTools  []string
	DisabledTools []string
}
func (te *ToolEnablement) isToolEnabled(toolName string) bool {
	for _, t := range te.DisabledTools {
		if t == toolName {
			return false
		}
	}
	if len(te.EnabledTools) == 0 {
		return true
	}
	for _, t := range te.EnabledTools {
		if t == toolName {
			return true
		}
	}
	return false
}
func boolPtr(b bool) *bool {
	return &b
}
func (so *SecurityOptions) isConnectionAllowed(connection string) bool {
	if len(so.AllowedConnections) == 0 {
		return true
	}
	for _, allowed := range so.AllowedConnections {
		if allowed == connection {
			return true
		}
	}
	return false
}
func registerTools(server *mcp.Server, secOpts *SecurityOptions, toolEnablement *ToolEnablement) {
	if toolEnablement == nil {
		toolEnablement = &ToolEnablement{}
	}
	queryDesc := buildQueryDescription(secOpts)
	if toolEnablement.isToolEnabled("query") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_query",
			Description: queryDesc,
			Annotations: &mcp.ToolAnnotations{
				Title:           "Execute SQL Query",
				ReadOnlyHint:    secOpts.ReadOnly,
				DestructiveHint: boolPtr(!secOpts.ReadOnly && !secOpts.ConfirmWrites),
				IdempotentHint:  false,
				OpenWorldHint:   boolPtr(false),
			},
		}, createQueryHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("schemas") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_schemas",
			Description: descSchemas,
			Annotations: &mcp.ToolAnnotations{
				Title:          "List Database Schemas",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createSchemasHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("tables") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_tables",
			Description: descTables,
			Annotations: &mcp.ToolAnnotations{
				Title:          "List Database Tables",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createTablesHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("columns") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_columns",
			Description: descColumns,
			Annotations: &mcp.ToolAnnotations{
				Title:          "Describe Table Columns",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createColumnsHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("connections") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_connections",
			Description: descConnections,
			Annotations: &mcp.ToolAnnotations{
				Title:          "List Database Connections",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createConnectionsHandler(secOpts))
	}
	if secOpts.ConfirmWrites && toolEnablement.isToolEnabled("confirm") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_confirm",
			Description: descConfirm,
			Annotations: &mcp.ToolAnnotations{
				Title:           "Confirm Write Operation",
				ReadOnlyHint:    false,
				DestructiveHint: boolPtr(true),
				IdempotentHint:  false,
				OpenWorldHint:   boolPtr(false),
			},
		}, createConfirmHandler(secOpts))
	}
	if secOpts.ConfirmWrites && toolEnablement.isToolEnabled("pending") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_pending",
			Description: descPending,
			Annotations: &mcp.ToolAnnotations{
				Title:          "List Pending Confirmations",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createPendingHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("explain") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_explain",
			Description: descExplain,
			Annotations: &mcp.ToolAnnotations{
				Title:          "Explain SQL Query",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createExplainHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("diff") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_diff",
			Description: descDiff,
			Annotations: &mcp.ToolAnnotations{
				Title:          "Compare Schemas",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createSchemaDiffHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("erd") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_erd",
			Description: descERD,
			Annotations: &mcp.ToolAnnotations{
				Title:          "Load ER Diagram Metadata",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createERDHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("audit") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_audit",
			Description: descAudit,
			Annotations: &mcp.ToolAnnotations{
				Title:          "Run Data Quality Audit",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createAuditHandler(secOpts))
	}
	if toolEnablement.isToolEnabled("suggestions") {
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whodb_suggestions",
			Description: descSuggestions,
			Annotations: &mcp.ToolAnnotations{
				Title:          "Load Query Suggestions",
				ReadOnlyHint:   true,
				IdempotentHint: true,
				OpenWorldHint:  boolPtr(false),
			},
		}, createSuggestionsHandler(secOpts))
	}
}
func buildQueryDescription(secOpts *SecurityOptions) string {
	base := `Execute a SQL query against a database connection.
**Best for:** Running SQL SELECT, INSERT, UPDATE, DELETE statements when you need to query or modify data.
**Not recommended for:** Schema exploration (use whodb_schemas, whodb_tables, whodb_columns instead for faster, structured results).
**Common mistakes:** Running queries without specifying connection when multiple exist; using SELECT * instead of specific columns; forgetting LIMIT on large tables.
**Usage Example (simple query):**
` + "```json" + `
{
  "name": "whodb_query",
  "arguments": {
    "connection": "mydb",
    "query": "SELECT id, name, email FROM users WHERE active = true LIMIT 10"
  }
}
` + "```" + `
**Usage Example (parameterized query - RECOMMENDED for user input):**
` + "```json" + `
{
  "name": "whodb_query",
  "arguments": {
    "connection": "mydb",
    "query": "SELECT * FROM users WHERE id = $1 AND status = $2",
    "parameters": [123, "active"]
  }
}
` + "```" + `
**Placeholder syntax by database:** PostgreSQL uses $1, $2, $3; MySQL/SQLite/DuckDB/ClickHouse use ?
**Best practices:**
- **Use parameterized queries** when incorporating user-provided values - this prevents SQL injection
- Always use LIMIT for exploratory queries to avoid overwhelming results
- Prefer specific column selection over SELECT *
- Check schema structure with whodb_columns before writing complex queries`
	if secOpts.ReadOnly {
		return base + `
**Security Mode: READ-ONLY**
Only SELECT, SHOW, DESCRIBE, and EXPLAIN queries are allowed. Write operations (INSERT, UPDATE, DELETE, DROP, etc.) are blocked.`
	}
	if secOpts.ConfirmWrites {
		return base + `
**Security Mode: CONFIRM-WRITES (Default)**
Write operations (INSERT, UPDATE, DELETE, etc.) require user confirmation. When you submit a write query:
1. The query is validated but NOT executed
2. You receive a confirmation_token
3. Explain to the user what the query will do
4. Call whodb_confirm with the token after user approves
5. The query executes and returns results`
	}
	return base + `
**Security Mode: ALLOW-WRITE**
Full write access enabled. All queries execute immediately. Use with caution in production.`
}
func createQueryHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input QueryInput) (*mcp.CallToolResult, QueryOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input QueryInput) (*mcp.CallToolResult, QueryOutput, error) {
		return HandleQuery(ctx, req, input, secOpts)
	}
}
func createConfirmHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input ConfirmInput) (*mcp.CallToolResult, ConfirmOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ConfirmInput) (*mcp.CallToolResult, ConfirmOutput, error) {
		return HandleConfirm(ctx, req, input, secOpts)
	}
}
func createPendingHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input PendingInput) (*mcp.CallToolResult, PendingOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input PendingInput) (*mcp.CallToolResult, PendingOutput, error) {
		return HandlePending(ctx, req, input, secOpts)
	}
}
func createSchemasHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input SchemasInput) (*mcp.CallToolResult, SchemasOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SchemasInput) (*mcp.CallToolResult, SchemasOutput, error) {
		if input.Connection == "" && secOpts.DefaultConnection != "" {
			input.Connection = secOpts.DefaultConnection
		}
		if !secOpts.isConnectionAllowed(input.Connection) {
			return nil, SchemasOutput{Error: fmt.Sprintf("connection %q is not allowed", input.Connection)}, nil
		}
		return HandleSchemas(ctx, req, input)
	}
}
func createTablesHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input TablesInput) (*mcp.CallToolResult, TablesOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input TablesInput) (*mcp.CallToolResult, TablesOutput, error) {
		if input.Connection == "" && secOpts.DefaultConnection != "" {
			input.Connection = secOpts.DefaultConnection
		}
		if !secOpts.isConnectionAllowed(input.Connection) {
			return nil, TablesOutput{Error: fmt.Sprintf("connection %q is not allowed", input.Connection)}, nil
		}
		return HandleTables(ctx, req, input)
	}
}
func createColumnsHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input ColumnsInput) (*mcp.CallToolResult, ColumnsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ColumnsInput) (*mcp.CallToolResult, ColumnsOutput, error) {
		if input.Connection == "" && secOpts.DefaultConnection != "" {
			input.Connection = secOpts.DefaultConnection
		}
		if !secOpts.isConnectionAllowed(input.Connection) {
			return nil, ColumnsOutput{Error: fmt.Sprintf("connection %q is not allowed", input.Connection)}, nil
		}
		return HandleColumns(ctx, req, input)
	}
}
func createConnectionsHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input ConnectionsInput) (*mcp.CallToolResult, ConnectionsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ConnectionsInput) (*mcp.CallToolResult, ConnectionsOutput, error) {
		result, output, err := HandleConnections(ctx, req, input)
		if err != nil || output.Error != "" {
			return result, output, err
		}
		if len(secOpts.AllowedConnections) > 0 {
			filtered := make([]ConnectionInfo, 0)
			for _, conn := range output.Connections {
				if secOpts.isConnectionAllowed(conn.Name) {
					filtered = append(filtered, conn)
				}
			}
			output.Connections = filtered
		}
		return result, output, err
	}
}
func createExplainHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input ExplainInput) (*mcp.CallToolResult, ExplainOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ExplainInput) (*mcp.CallToolResult, ExplainOutput, error) {
		if input.Connection == "" && secOpts.DefaultConnection != "" {
			input.Connection = secOpts.DefaultConnection
		}
		if !secOpts.isConnectionAllowed(input.Connection) {
			return nil, ExplainOutput{Error: fmt.Sprintf("connection %q is not allowed", input.Connection)}, nil
		}
		return HandleExplain(ctx, req, input)
	}
}
func createSchemaDiffHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input SchemaDiffInput) (*mcp.CallToolResult, SchemaDiffOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SchemaDiffInput) (*mcp.CallToolResult, SchemaDiffOutput, error) {
		if input.FromConnection == "" && secOpts.DefaultConnection != "" {
			input.FromConnection = secOpts.DefaultConnection
		}
		if input.ToConnection == "" && secOpts.DefaultConnection != "" {
			input.ToConnection = secOpts.DefaultConnection
		}
		if !secOpts.isConnectionAllowed(input.FromConnection) {
			return nil, SchemaDiffOutput{Error: fmt.Sprintf("connection %q is not allowed", input.FromConnection)}, nil
		}
		if !secOpts.isConnectionAllowed(input.ToConnection) {
			return nil, SchemaDiffOutput{Error: fmt.Sprintf("connection %q is not allowed", input.ToConnection)}, nil
		}
		return HandleSchemaDiff(ctx, req, input)
	}
}
func createERDHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input ERDInput) (*mcp.CallToolResult, ERDOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ERDInput) (*mcp.CallToolResult, ERDOutput, error) {
		if input.Connection == "" && secOpts.DefaultConnection != "" {
			input.Connection = secOpts.DefaultConnection
		}
		if !secOpts.isConnectionAllowed(input.Connection) {
			return nil, ERDOutput{Error: fmt.Sprintf("connection %q is not allowed", input.Connection)}, nil
		}
		return HandleERD(ctx, req, input)
	}
}
func createAuditHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input AuditInput) (*mcp.CallToolResult, AuditOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AuditInput) (*mcp.CallToolResult, AuditOutput, error) {
		if input.Connection == "" && secOpts.DefaultConnection != "" {
			input.Connection = secOpts.DefaultConnection
		}
		if !secOpts.isConnectionAllowed(input.Connection) {
			return nil, AuditOutput{Error: fmt.Sprintf("connection %q is not allowed", input.Connection)}, nil
		}
		return HandleAudit(ctx, req, input)
	}
}
func createSuggestionsHandler(secOpts *SecurityOptions) func(ctx context.Context, req *mcp.CallToolRequest, input SuggestionsInput) (*mcp.CallToolResult, SuggestionsOutput, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SuggestionsInput) (*mcp.CallToolResult, SuggestionsOutput, error) {
		if input.Connection == "" && secOpts.DefaultConnection != "" {
			input.Connection = secOpts.DefaultConnection
		}
		if !secOpts.isConnectionAllowed(input.Connection) {
			return nil, SuggestionsOutput{Error: fmt.Sprintf("connection %q is not allowed", input.Connection)}, nil
		}
		return HandleSuggestions(ctx, req, input)
	}
}
func registerResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:        "connections",
		URI:         "whodb://connections",
		Description: "List of available database connections",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		conns, err := ListAvailableConnections()
		if err != nil {
			return nil, err
		}
		data, _ := json.MarshalIndent(conns, "", "  ")
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{Text: string(data)},
			},
		}, nil
	})
}
func registerPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "query_help",
		Title:       "SQL Query Help",
		Description: "Get guidance on writing SQL queries with WhoDB, including best practices and examples.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "database_type",
				Title:       "Database Type",
				Description: "The type of database (postgres, mysql, sqlite, duckdb, etc.) for dialect-specific help.",
				Required:    false,
			},
			{
				Name:        "query_type",
				Title:       "Query Type",
				Description: "The type of query you need help with (select, insert, update, delete, join, aggregate).",
				Required:    false,
			},
		},
	}, handleQueryHelpPrompt)
	server.AddPrompt(&mcp.Prompt{
		Name:        "schema_exploration_help",
		Title:       "Schema Exploration Help",
		Description: "Learn how to effectively explore database schemas, tables, and relationships using WhoDB tools.",
	}, handleSchemaExplorationHelpPrompt)
	server.AddPrompt(&mcp.Prompt{
		Name:        "workflow_help",
		Title:       "Database Workflow Help",
		Description: "Get guidance on common database workflows like data analysis, debugging queries, or understanding table relationships.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "workflow",
				Title:       "Workflow Type",
				Description: "The workflow you need help with (analysis, debugging, relationships, migration).",
				Required:    false,
			},
		},
	}, handleWorkflowHelpPrompt)
}
func handleQueryHelpPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	dbType := ""
	queryType := ""
	if req.Params.Arguments != nil {
		if v, ok := req.Params.Arguments["database_type"]; ok {
			dbType = v
		}
		if v, ok := req.Params.Arguments["query_type"]; ok {
			queryType = v
		}
	}
	content := buildQueryHelpContent(dbType, queryType)
	return &mcp.GetPromptResult{
		Description: "SQL query guidance for WhoDB",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}
func handleSchemaExplorationHelpPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	content := `I need help exploring a database schema. Please guide me through the WhoDB workflow:
## Recommended Exploration Workflow
1. **Start with connections** - Use whodb_connections to see available databases
2. **Get tables with columns** - Use whodb_tables(include_columns=true) to get all tables AND their column details in one call
3. **Query data** - Use whodb_query for actual data exploration (results include column_types)
For multi-schema databases, use whodb_schemas(include_tables=true) first to see all schemas and their tables.
## Tips for Effective Exploration
- Use include_columns=true and include_tables=true to minimize round-trips
- Query results include column_types alongside column names — no separate whodb_columns call needed
- Always check foreign key relationships in column details
- Look for naming patterns (e.g., *_id columns often indicate relationships)
- Check for audit columns (created_at, updated_at) to understand data lifecycle
## Example Exploration Session
` + "```" + `
# Step 1: What databases are available?
whodb_connections → [{name: "mydb", type: "postgres", ...}]
# Step 2: Get all tables and their columns in one call
whodb_tables(connection="mydb", include_columns=true) → [
  {name: "users", columns: [{name: "id", type: "integer", is_primary: true}, ...]},
  {name: "orders", columns: [{name: "user_id", type: "integer", is_foreign_key: true, referenced_table: "users"}, ...]}
]
# Step 3: Query with full type awareness
whodb_query(connection="mydb", query="SELECT * FROM users LIMIT 5")
→ {columns: ["id", "name"], column_types: ["integer", "varchar"], rows: [...]}
` + "```" + `
Please help me explore my database using this workflow.`
	return &mcp.GetPromptResult{
		Description: "Database schema exploration guidance",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}
func handleWorkflowHelpPrompt(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	workflow := "general"
	if req.Params.Arguments != nil {
		if v, ok := req.Params.Arguments["workflow"]; ok {
			workflow = v
		}
	}
	content := buildWorkflowHelpContent(workflow)
	return &mcp.GetPromptResult{
		Description: "Database workflow guidance",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: content},
			},
		},
	}, nil
}
func buildQueryHelpContent(dbType, queryType string) string {
	base := `I need help writing SQL queries with WhoDB. Please provide guidance on:
## Query Best Practices
1. **Always use LIMIT** for exploratory queries to avoid overwhelming results
2. **Select specific columns** instead of SELECT * for better performance
3. **Check schema first** with whodb_columns before writing complex queries
4. **Use parameterized patterns** when constructing dynamic queries
## Security Notes
- WhoDB validates queries before execution
- Multi-statement queries (separated by ;) are blocked by default
- Dangerous functions may be blocked depending on security level
- Write operations may require confirmation depending on server mode
`
	if dbType != "" {
		base += fmt.Sprintf("\n## %s-Specific Tips\n\n", dbType)
		switch dbType {
		case "postgres":
			base += `- Use ILIKE for case-insensitive pattern matching
- Use jsonb operators (->>, @>) for JSON columns
- Use CTEs (WITH clause) for complex queries
- Use EXPLAIN ANALYZE to understand query performance
`
		case "mysql":
			base += `- Use LIKE with BINARY for case-sensitive matching
- Use JSON_EXTRACT() for JSON columns
- Use SHOW CREATE TABLE to see table definitions
- Use EXPLAIN to understand query execution
`
		case "sqlite":
			base += `- SQLite is type-flexible, but be consistent
- Use json_extract() for JSON columns
- Use .schema command equivalent via whodb_columns
- VACUUM periodically for performance
`
		case "duckdb":
			base += `- DuckDB uses ? for query placeholders
- Use information_schema for metadata queries
- DuckDB supports INTERVAL, HUGEINT, and JSON types natively
- Use duckdb_constraints() for constraint introspection
`
		}
	}
	if queryType != "" {
		base += fmt.Sprintf("\n## %s Query Examples\n\n", queryType)
		switch queryType {
		case "select":
			base += "```sql\n-- Basic select with filtering\nSELECT id, name, email FROM users WHERE active = true LIMIT 10;\n\n-- With ordering\nSELECT * FROM orders ORDER BY created_at DESC LIMIT 20;\n```\n"
		case "insert":
			base += "```sql\n-- Single row insert\nINSERT INTO users (name, email) VALUES ('John', 'john@example.com');\n\n-- Multiple rows\nINSERT INTO users (name, email) VALUES ('A', 'a@x.com'), ('B', 'b@x.com');\n```\n"
		case "update":
			base += "```sql\n-- Update with WHERE clause (always use WHERE!)\nUPDATE users SET active = false WHERE last_login < '2024-01-01';\n\n-- Update single row by ID\nUPDATE orders SET status = 'shipped' WHERE id = 123;\n```\n"
		case "delete":
			base += "```sql\n-- Delete with WHERE clause (always use WHERE!)\nDELETE FROM sessions WHERE expires_at < NOW();\n\n-- Delete by ID\nDELETE FROM users WHERE id = 456;\n```\n"
		case "join":
			base += "```sql\n-- Inner join\nSELECT u.name, o.total FROM users u\nJOIN orders o ON u.id = o.user_id\nWHERE o.status = 'completed' LIMIT 10;\n\n-- Left join to include users without orders\nSELECT u.name, COUNT(o.id) as order_count\nFROM users u LEFT JOIN orders o ON u.id = o.user_id\nGROUP BY u.id, u.name;\n```\n"
		case "aggregate":
			base += "```sql\n-- Count with grouping\nSELECT status, COUNT(*) as count FROM orders GROUP BY status;\n\n-- Sum with filtering\nSELECT user_id, SUM(total) as total_spent\nFROM orders WHERE created_at > '2024-01-01'\nGROUP BY user_id HAVING SUM(total) > 100;\n```\n"
		}
	}
	base += "\nPlease help me write effective SQL queries following these guidelines."
	return base
}
func buildWorkflowHelpContent(workflow string) string {
	switch workflow {
	case "analysis":
		return `I need help with data analysis using WhoDB. Guide me through:
## Data Analysis Workflow
1. **Understand the data** - Explore schema to find relevant tables
2. **Sample the data** - Run small queries to understand data patterns
3. **Aggregate and summarize** - Use GROUP BY, COUNT, SUM, AVG
4. **Filter and refine** - Add WHERE clauses based on findings
5. **Export or report** - Format results for further use
## Useful Analysis Patterns
` + "```sql" + `
-- Distribution analysis
SELECT column, COUNT(*) as count
FROM table GROUP BY column ORDER BY count DESC;
-- Time-based analysis
SELECT DATE(created_at) as date, COUNT(*) as daily_count
FROM events GROUP BY DATE(created_at) ORDER BY date;
-- Percentile/quantile analysis (PostgreSQL)
SELECT percentile_cont(0.5) WITHIN GROUP (ORDER BY value) as median
FROM measurements;
` + "```" + `
Please help me analyze my data using these patterns.`
	case "debugging":
		return `I need help debugging database queries with WhoDB. Guide me through:
## Query Debugging Workflow
1. **Start simple** - Run the simplest version of your query
2. **Add complexity gradually** - Add one clause at a time
3. **Check data types** - Use whodb_columns to verify column types
4. **Verify relationships** - Check foreign keys before JOINs
5. **Inspect intermediate results** - Break complex queries into steps
## Common Issues to Check
- NULL handling (use IS NULL, not = NULL)
- Data type mismatches in comparisons
- Missing indexes causing slow queries
- Incorrect JOIN conditions causing duplicates
- Case sensitivity in string comparisons
## Debugging Steps
` + "```" + `
# Step 1: Check table structure
whodb_columns(table="problematic_table")
# Step 2: Sample raw data
SELECT * FROM table LIMIT 5;
# Step 3: Check for NULLs
SELECT COUNT(*), COUNT(column) FROM table;
# Step 4: Verify JOIN data exists
SELECT COUNT(*) FROM table1 t1
JOIN table2 t2 ON t1.id = t2.foreign_id;
` + "```" + `
Please help me debug my query using this approach.`
	case "relationships":
		return `I need help understanding table relationships with WhoDB.
## Finding Relationships
1. **Check foreign keys** - whodb_columns shows is_foreign_key and referenced_table
2. **Look for naming patterns** - Columns like user_id, order_id suggest relationships
3. **Verify with data** - Query to confirm relationships exist
## Relationship Discovery Workflow
` + "```" + `
# Step 1: Get columns with FK info
whodb_columns(table="orders")
# Look for: is_foreign_key: true, referenced_table: "users"
# Step 2: Verify the relationship
SELECT o.id, u.name
FROM orders o JOIN users u ON o.user_id = u.id LIMIT 5;
# Step 3: Check cardinality
SELECT user_id, COUNT(*) as order_count
FROM orders GROUP BY user_id ORDER BY order_count DESC LIMIT 10;
` + "```" + `
## Common Relationship Types
- **One-to-Many**: user_id in orders (one user has many orders)
- **Many-to-Many**: Usually via junction table (user_roles with user_id and role_id)
- **One-to-One**: Rare, usually same primary key in both tables
Please help me map out the relationships in my database.`
	default:
		return `I need help with database operations using WhoDB.
## Available Workflows
1. **Schema Exploration** - Discover tables, columns, and relationships
2. **Data Analysis** - Aggregate, summarize, and understand data
3. **Query Debugging** - Troubleshoot slow or incorrect queries
4. **Relationship Mapping** - Understand how tables connect
## General Best Practices
- Always start with whodb_connections to see available databases
- Use whodb_tables(include_columns=true) to get tables and columns in one call
- Query results include column_types — no separate whodb_columns call needed
- Always use LIMIT for exploratory queries
- Check column types and relationships before writing JOINs
Please guide me through the appropriate workflow for my task.`
	}
}
type TransportType string
const (
	TransportStdio TransportType = "stdio"
	TransportHTTP TransportType = "http"
)
type HTTPOptions struct {
	Host string
	Port int
}
func Run(ctx context.Context, server *mcp.Server) error {
	return server.Run(ctx, &mcp.StdioTransport{})
}
func RunHTTP(ctx context.Context, server *mcp.Server, opts *HTTPOptions, logger *slog.Logger) error {
	if opts == nil {
		opts = &HTTPOptions{}
	}
	if opts.Host == "" {
		opts.Host = "localhost"
	}
	if opts.Port == 0 {
		opts.Port = 3000
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	if opts.Host != "localhost" && opts.Host != "127.0.0.1" {
		logger.Warn("MCP server binding to network interface - no authentication is configured",
			"host", opts.Host,
			"recommendation", "Use --host=localhost for local development, or add authentication for production")
	}
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/mcp", handler)
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()
	logger.Info("MCP server listening", "transport", "http", "address", addr, "endpoint", "/mcp", "health", "/health")
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server error: %w", err)
	}
	return nil
}
const descSchemas = `List all schemas (namespaces) in a database.
**Best for:** Discovering what schemas exist in a database; understanding database organization before exploring tables.
**Not recommended for:** When you already know the schema name (proceed directly to whodb_tables).
**Common mistakes:** Calling this repeatedly - schema lists rarely change during a session.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_schemas",
  "arguments": {
    "connection": "mydb"
  }
}
` + "```" + `
**Returns:** Array of schema names (e.g., ["public", "analytics", "audit"]).
**Typical workflow:** whodb_schemas → whodb_tables(include_columns=true) → whodb_query
**Optional parameter:** Set "include_tables": true to also return all tables within each schema in a single call. This populates a "details" array with schema names and their tables, saving you a separate whodb_tables call per schema.`
const descTables = `List all tables in a database schema.
**Best for:** Discovering what tables exist in a schema; getting table metadata like row counts.
**Not recommended for:** When you already know the table name (proceed directly to whodb_columns or whodb_query).
**Common mistakes:** Not specifying schema when the database has multiple schemas with same-named tables.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_tables",
  "arguments": {
    "connection": "mydb",
    "schema": "public"
  }
}
` + "```" + `
**Returns:** Array of table objects with name and attributes (row count, size, etc.).
**Note:** If schema is omitted, uses the connection's default schema or the first available schema.
**Optional parameter:** Set "include_columns": true to also return column details (name, type, primary key, foreign keys) for each table. This saves you separate whodb_columns calls and gives you everything needed to write queries in a single round-trip.`
const descColumns = `Describe the columns in a database table.
**Best for:** Understanding table structure before writing queries; discovering primary keys and foreign key relationships.
**Not recommended for:** When you need actual data (use whodb_query with SELECT).
**Common mistakes:** Forgetting to specify the table name; not using this before writing INSERT/UPDATE queries.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_columns",
  "arguments": {
    "connection": "mydb",
    "table": "users",
    "schema": "public"
  }
}
` + "```" + `
**Returns:** Array of column objects with:
- name: Column name
- type: Data type (varchar, integer, timestamp, etc.)
- is_primary: Whether this is a primary key
- is_foreign_key: Whether this references another table
- referenced_table/referenced_column: Foreign key target (if applicable)
**Pro tip:** Always call this before writing INSERT queries to ensure correct column names and types.`
const descConnections = `List all available database connections.
**Best for:** Discovering what databases are configured; choosing which connection to use.
**Not recommended for:** When you already know the connection name.
**Common mistakes:** Not calling this first when connection name is unknown.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_connections",
  "arguments": {}
}
` + "```" + `
**Returns:** Array of connection objects with:
- name: Connection identifier to use in other tools
- type: Database type (postgres, mysql, sqlite, duckdb, etc.)
- host/port/database: Connection details (passwords are never exposed)
- source: "saved" (from CLI config) or "env" (from environment variables)
**Note:** If only one connection exists, other tools will use it automatically when connection is omitted.`
const descConfirm = `Confirm and execute a pending write operation.
**Best for:** Executing write queries after user approval in confirm-writes mode.
**Not recommended for:** Read queries (they execute immediately without confirmation).
**Common mistakes:** Using an expired token (tokens expire after 5 minutes); not explaining the query to the user before confirming.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_confirm",
  "arguments": {
    "token": "550e8400-e29b-41d4-a716-446655440000"
  }
}
` + "```" + `
**Workflow:**
1. Call whodb_query with a write operation (INSERT, UPDATE, DELETE, etc.)
2. Receive confirmation_required=true, a confirmation_token, and confirmation_expiry
3. Explain to the user what the query will do in plain language
4. After user approves, call whodb_confirm with the token
5. Query executes and returns results
**Token behavior:** Tokens are valid for 5 minutes (expiry time is in the response). If confirmation fails due to a connection error or timeout, you can retry with the same token — it is only consumed after successful execution. Use whodb_pending to list active tokens if you lose track.`
const descPending = `List all pending write confirmations that are waiting for approval.
**Best for:** Recovering lost confirmation tokens; checking what operations are pending.
**Not recommended for:** Anything else — this is a utility tool for the confirm-writes workflow.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_pending",
  "arguments": {}
}
` + "```" + `
**Returns:** Array of pending confirmations with token, query, connection, and expiry time.
**Important:** Tokens are single-use and expire after 60 seconds. If expired, re-submit the original query to get a new token.`
const descExplain = `Run EXPLAIN for a SQL query using the database's native explain mode.
**Best for:** Understanding query plans; checking whether a query will scan too much data before you run the real query.
**Not recommended for:** Fetching actual data (use whodb_query for that).
**Common mistakes:** Passing a non-SQL string; forgetting that EXPLAIN output is database-specific.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_explain",
  "arguments": {
    "connection": "mydb",
    "query": "SELECT * FROM users WHERE email LIKE '%@example.com' LIMIT 10"
  }
}
` + "```" + `
**Returns:** The database-native EXPLAIN output with columns and rows, ready for follow-up analysis.`
const descDiff = `Compare schema metadata between two database connections.
**Best for:** Spotting drift between environments; comparing staging vs production; reviewing storage-unit, column, and relationship changes.
**Not recommended for:** Row-level data comparison.
**Common mistakes:** Forgetting to specify both connections; comparing the same connection and schema without overrides.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_diff",
  "arguments": {
    "from_connection": "staging",
    "to_connection": "prod",
    "from_schema": "public",
    "to_schema": "public"
  }
}
` + "```" + `
**Returns:** A structured schema diff with storage-unit, column, and relationship summaries plus per-object changes.`
const descERD = `Load backend graph metadata for a schema or database.
**Best for:** Understanding how tables relate before writing joins; inspecting primary/foreign key relationships programmatically.
**Not recommended for:** Query execution.
**Common mistakes:** Expecting row data instead of metadata.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_erd",
  "arguments": {
    "connection": "mydb",
    "schema": "public"
  }
}
` + "```" + `
**Returns:** Storage units with columns plus normalized relationship edges sourced from backend graph metadata.`
const descAudit = `Run data-quality checks on one schema or table.
**Best for:** Finding null-rate spikes, missing primary keys, low-cardinality issues, duplicate rows, and orphaned foreign keys.
**Not recommended for:** Replacing a full observability or data-governance system.
**Common mistakes:** Forgetting to scope the audit to one table when you only need one table.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_audit",
  "arguments": {
    "connection": "mydb",
    "schema": "public",
    "table": "orders",
    "null_warning": 15,
    "null_error": 60
  }
}
` + "```" + `
**Returns:** Audit results per table, including issue summaries and the underlying table/column findings.`
const descSuggestions = `Load backend-generated starter queries for a schema or database.
**Best for:** Quickly orienting yourself in an unfamiliar database; suggesting first queries for exploration.
**Not recommended for:** Exhaustive SQL tutoring.
**Common mistakes:** Treating the suggestions as guaranteed-valid business logic rather than onboarding hints.
**Usage Example:**
` + "```json" + `
{
  "name": "whodb_suggestions",
  "arguments": {
    "connection": "mydb",
    "schema": "public"
  }
}
` + "```" + `
**Returns:** A short list of backend-generated query suggestions derived from the actual storage units in the resolved schema.`
const defaultInstructions = `WhoDB MCP Server - Database Management Tools
Available tools:
- whodb_query: Execute SQL queries against a database
- whodb_schemas: List database schemas/namespaces
- whodb_tables: List tables in a schema
- whodb_columns: Describe columns in a table
- whodb_connections: List available database connections
- whodb_confirm: Confirm pending write operations (enabled by default)
- whodb_pending: List pending write confirmations awaiting approval
- whodb_explain: Run EXPLAIN for a SQL query
- whodb_diff: Compare schema metadata between two connections
- whodb_erd: Load graph/relationship metadata for a schema
- whodb_audit: Run data quality checks on one schema or table
- whodb_suggestions: Load backend-generated starter queries
Available prompts (for guidance):
- query_help: Get SQL query writing guidance (supports database_type and query_type args)
- schema_exploration_help: Learn how to explore database structure effectively
- workflow_help: Get guidance on common database workflows (analysis, debugging, relationships)
SECURITY MODE:
This server runs with confirm-writes enabled by default. When you execute write operations
(INSERT, UPDATE, DELETE, etc.), the query will be shown to the user for approval before
executing. This keeps the user in control while giving you full database functionality.
IMPORTANT: When a write operation requires confirmation, inform the user clearly:
- Tell them what query you're proposing to run
- Explain what it will do in plain language
- Let them know they'll see a confirmation prompt
If the user wants different security settings, they can restart with:
- --read-only: No writes allowed at all
- --allow-write: Full write access without confirmation (not recommended for production)
Connection names reference either:
1. Environment profiles, for example:
   - WHODB_POSTGRES='[{"alias":"prod","host":"localhost","user":"user","password":"pass","database":"db","port":"5432"}]'
   - WHODB_MYSQL_1='{"alias":"staging","host":"localhost","user":"user","password":"pass","database":"db","port":"3306"}'
2. Saved connections from WhoDB CLI configuration
Saved connections take precedence when names collide.
Example workflow:
1. List connections: whodb_connections
2. Get tables with columns: whodb_tables(connection="mydb", include_columns=true)
3. Query data: whodb_query(connection="mydb", query="SELECT * FROM users LIMIT 10")
   → Results include column_types alongside column names
4. Inspect a plan: whodb_explain(connection="mydb", query="SELECT ...")
5. Compare environments: whodb_diff(from_connection="staging", to_connection="prod")
6. Explore relationships: whodb_erd(connection="mydb")
7. Audit one schema: whodb_audit(connection="mydb")
8. Get starter queries: whodb_suggestions(connection="mydb")
9. Write data: whodb_query(connection="mydb", query="INSERT INTO...") -> user confirms -> whodb_confirm(token="...")
   → Tokens are valid for 5 minutes and retryable. Use whodb_pending to recover lost tokens.
Best practices:
- Send ONE query at a time (multi-statement queries are blocked for security)
- Always use LIMIT for exploratory queries
- Check schema structure before writing queries
- Prefer specific column selection over SELECT *
- For writes, explain to the user what will happen before proposing the query
- Use parameterized queries when incorporating user input: whodb_query(query="SELECT * FROM users WHERE id = $1", parameters=[42])
  Placeholders: PostgreSQL uses $1, $2; MySQL/SQLite/DuckDB/ClickHouse use ?
`