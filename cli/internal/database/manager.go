package database
import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"github.com/clidey/whodb/cli/internal/baml"
	"github.com/clidey/whodb/cli/internal/bootstrap"
	"github.com/clidey/whodb/cli/internal/config"
	connresolver "github.com/clidey/whodb/cli/internal/connections"
	"github.com/clidey/whodb/cli/internal/sourcetypes"
	tunnelpkg "github.com/clidey/whodb/cli/internal/ssh"
	"github.com/clidey/whodb/core/graph/model"
	"github.com/clidey/whodb/core/src"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/envconfig"
	"github.com/clidey/whodb/core/src/llm"
	queryast "github.com/clidey/whodb/core/src/query"
	"github.com/clidey/whodb/core/src/source"
	"github.com/clidey/whodb/core/src/sourcecatalog"
	"github.com/xuri/excelize/v2"
	"golang.org/x/sync/errgroup"
)
const MaxQueryLogEntries = 100
type QueryLogEntry struct {
	Query     string
	Timestamp time.Time
	Duration  time.Duration
	Success   bool
	Error     string
	RowCount  int
}
type Connection = config.Connection
const ConnectionSourceSaved = connresolver.ConnectionSourceSaved
const ConnectionSourceEnv = connresolver.ConnectionSourceEnv
type ConnectionSourceInfo = connresolver.ConnectionSourceInfo
type QuerySuggestion = source.QuerySuggestion
const DefaultCacheTTL = 5 * time.Minute
var ErrReadOnly = fmt.Errorf("query blocked: read-only mode is enabled")
var mutationKeywords = map[string]bool{
	"INSERT":   true,
	"UPDATE":   true,
	"DELETE":   true,
	"DROP":     true,
	"ALTER":    true,
	"CREATE":   true,
	"TRUNCATE": true,
	"GRANT":    true,
	"REVOKE":   true,
}
func IsMutationQuery(query string) bool {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return false
	}
	end := strings.IndexAny(trimmed, " \t\n\r;(")
	var firstWord string
	if end == -1 {
		firstWord = trimmed
	} else {
		firstWord = trimmed[:end]
	}
	return mutationKeywords[strings.ToUpper(firstWord)]
}
type MetadataCache struct {
	mu sync.RWMutex
	schemas     []string
	schemasTime time.Time
	tables     map[string][]engine.StorageUnit
	tablesTime map[string]time.Time
	columns     map[string][]engine.Column
	columnsTime map[string]time.Time
	ttl time.Duration
}
func NewMetadataCache(ttl time.Duration) *MetadataCache {
	return &MetadataCache{
		tables:      make(map[string][]engine.StorageUnit),
		tablesTime:  make(map[string]time.Time),
		columns:     make(map[string][]engine.Column),
		columnsTime: make(map[string]time.Time),
		ttl:         ttl,
	}
}
func (c *MetadataCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.schemas = nil
	c.schemasTime = time.Time{}
	c.tables = make(map[string][]engine.StorageUnit)
	c.tablesTime = make(map[string]time.Time)
	c.columns = make(map[string][]engine.Column)
	c.columnsTime = make(map[string]time.Time)
}
func (c *MetadataCache) GetSchemas() ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.schemas == nil || time.Since(c.schemasTime) > c.ttl {
		return nil, false
	}
	return c.schemas, true
}
func (c *MetadataCache) SetSchemas(schemas []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.schemas = schemas
	c.schemasTime = time.Now()
}
func (c *MetadataCache) GetTables(schema string) ([]engine.StorageUnit, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	tables, ok := c.tables[schema]
	if !ok {
		return nil, false
	}
	cacheTime, ok := c.tablesTime[schema]
	if !ok || time.Since(cacheTime) > c.ttl {
		return nil, false
	}
	return tables, true
}
func (c *MetadataCache) SetTables(schema string, tables []engine.StorageUnit) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tables[schema] = tables
	c.tablesTime[schema] = time.Now()
}
func (c *MetadataCache) GetColumns(schema, table string) ([]engine.Column, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := schema + "." + table
	columns, ok := c.columns[key]
	if !ok {
		return nil, false
	}
	cacheTime, ok := c.columnsTime[key]
	if !ok || time.Since(cacheTime) > c.ttl {
		return nil, false
	}
	return columns, true
}
func (c *MetadataCache) SetColumns(schema, table string, columns []engine.Column) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := schema + "." + table
	c.columns[key] = columns
	c.columnsTime[key] = time.Now()
}
type Manager struct {
	currentConnection *Connection
	config            *config.Config
	cache             *MetadataCache
	queryLog          []QueryLogEntry
	queryLogMu        sync.Mutex
	tunnel            *tunnelpkg.Tunnel
}
func (m *Manager) buildCredentials(conn *Connection) *engine.Credentials {
	credentials := &engine.Credentials{
		Type:      conn.Type,
		Hostname:  conn.Host,
		Username:  conn.Username,
		Password:  conn.Password,
		Database:  conn.Database,
		IsProfile: conn.IsProfile,
	}
	var advanced []engine.Record
	if conn.Port > 0 {
		advanced = append(advanced, engine.Record{
			Key:   "Port",
			Value: fmt.Sprintf("%d", conn.Port),
		})
	}
	for key, value := range conn.Advanced {
		if key == "Port" && conn.Port > 0 {
			continue
		}
		advanced = append(advanced, engine.Record{
			Key:   key,
			Value: value,
		})
	}
	if len(advanced) > 0 {
		credentials.Advanced = advanced
	}
	return credentials
}
func NewManagerWithConfig(cfg *config.Config) (*Manager, error) {
	bootstrap.Ensure()
	if cfg == nil {
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("error loading config: %w", err)
		}
	}
	src.InitializeEngine()
	return &Manager{
		config: cfg,
		cache:  NewMetadataCache(DefaultCacheTTL),
	}, nil
}
func NewManager() (*Manager, error) {
	return NewManagerWithConfig(nil)
}
func (m *Manager) ListConnections() []Connection {
	return m.config.Connections
}
func (m *Manager) ListConnectionsWithSource() []ConnectionSourceInfo {
	return connresolver.NewResolverWithConfig(m.config).ListWithSource()
}
func (m *Manager) ListAvailableConnections() []Connection {
	infos := m.ListConnectionsWithSource()
	conns := make([]Connection, len(infos))
	for i, info := range infos {
		conns[i] = info.Connection
	}
	return conns
}
func (m *Manager) GetConnection(name string) (*Connection, error) {
	return m.config.GetConnection(name)
}
func (m *Manager) ResolveConnection(name string) (*Connection, string, error) {
	return connresolver.NewResolverWithConfig(m.config).Resolve(name)
}
func (m *Manager) Connect(conn *Connection) error {
	if m.tunnel != nil {
		m.tunnel.Stop()
		m.tunnel = nil
	}
	if conn.SSHHost != "" {
		sshPort := conn.SSHPort
		if sshPort == 0 {
			sshPort = 22
		}
		tunnel, err := tunnelpkg.NewTunnel(
			conn.SSHHost, sshPort, conn.SSHUser,
			conn.SSHKeyFile, conn.SSHPassword,
			conn.Host, conn.Port,
		)
		if err != nil {
			return fmt.Errorf("SSH tunnel setup failed: %w", err)
		}
		if err := tunnel.Start(); err != nil {
			tunnel.Stop()
			return fmt.Errorf("SSH tunnel failed to start: %w", err)
		}
		m.tunnel = tunnel
		conn.Host = "127.0.0.1"
		conn.Port = tunnel.LocalPort()
	}
	_, session, err := m.openSourceSession(context.Background(), conn)
	if err != nil {
		if m.tunnel != nil {
			m.tunnel.Stop()
			m.tunnel = nil
		}
		return fmt.Errorf("cannot connect to database. please check your credentials and ensure the database is accessible")
	}
	checker, ok := session.(source.AvailabilityChecker)
	if !ok || !checker.IsAvailable(context.Background()) {
		if m.tunnel != nil {
			m.tunnel.Stop()
			m.tunnel = nil
		}
		return fmt.Errorf("cannot connect to database. please check your credentials and ensure the database is accessible")
	}
	m.cache.Clear()
	m.currentConnection = conn
	return nil
}
func (m *Manager) Ping(conn *Connection) bool {
	_, session, err := m.openSourceSession(context.Background(), conn)
	if err != nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	checker, ok := session.(source.AvailabilityChecker)
	if !ok {
		return false
	}
	return checker.IsAvailable(ctx)
}
func (m *Manager) Disconnect() error {
	m.cache.Clear()
	m.currentConnection = nil
	if m.tunnel != nil {
		m.tunnel.Stop()
		m.tunnel = nil
	}
	return nil
}
func (m *Manager) InvalidateCache() {
	m.cache.Clear()
}
func (m *Manager) GetCache() *MetadataCache {
	return m.cache
}
func (m *Manager) logQuery(query string, start time.Time, result *engine.GetRowsResult, err error) {
	entry := QueryLogEntry{
		Query:     query,
		Timestamp: start,
		Duration:  time.Since(start),
		Success:   err == nil,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	if result != nil {
		entry.RowCount = len(result.Rows)
	}
	m.queryLogMu.Lock()
	defer m.queryLogMu.Unlock()
	m.queryLog = append(m.queryLog, entry)
	if len(m.queryLog) > MaxQueryLogEntries {
		m.queryLog = m.queryLog[len(m.queryLog)-MaxQueryLogEntries:]
	}
}
func (m *Manager) logStreamedQuery(query string, start time.Time, rowCount int, err error) {
	entry := QueryLogEntry{
		Query:     query,
		Timestamp: start,
		Duration:  time.Since(start),
		Success:   err == nil,
		RowCount:  rowCount,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	m.queryLogMu.Lock()
	defer m.queryLogMu.Unlock()
	m.queryLog = append(m.queryLog, entry)
	if len(m.queryLog) > MaxQueryLogEntries {
		m.queryLog = m.queryLog[len(m.queryLog)-MaxQueryLogEntries:]
	}
}
func (m *Manager) logOperation(operation string, start time.Time, count int, err error) {
	entry := QueryLogEntry{
		Query:     operation,
		Timestamp: start,
		Duration:  time.Since(start),
		Success:   err == nil,
		RowCount:  count,
	}
	if err != nil {
		entry.Error = err.Error()
	}
	m.queryLogMu.Lock()
	defer m.queryLogMu.Unlock()
	m.queryLog = append(m.queryLog, entry)
	if len(m.queryLog) > MaxQueryLogEntries {
		m.queryLog = m.queryLog[len(m.queryLog)-MaxQueryLogEntries:]
	}
}
func (m *Manager) GetQueryLog() []QueryLogEntry {
	m.queryLogMu.Lock()
	defer m.queryLogMu.Unlock()
	out := make([]QueryLogEntry, len(m.queryLog))
	copy(out, m.queryLog)
	return out
}
func (m *Manager) GetCurrentConnection() *Connection {
	return m.currentConnection
}
func (m *Manager) ResolveSnapshotSchema(conn *Connection, explicitSchema string) (string, error) {
	if strings.TrimSpace(explicitSchema) != "" {
		return explicitSchema, nil
	}
	if conn == nil {
		conn = m.currentConnection
	}
	if conn == nil {
		return "", fmt.Errorf("not connected to any database")
	}
	if sourcecatalog.UsesDatabaseInsteadOfSchema(conn.Type) && strings.TrimSpace(conn.Database) != "" {
		return conn.Database, nil
	}
	if strings.TrimSpace(conn.Schema) != "" {
		return conn.Schema, nil
	}
	schemas, err := m.GetSchemas()
	if err != nil || len(schemas) == 0 {
		return "", nil
	}
	return schemas[0], nil
}
func (m *Manager) GetQuerySuggestions(schema string) ([]QuerySuggestion, error) {
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	suggester, ok := session.(source.QuerySuggester)
	if !ok {
		return nil, fmt.Errorf("query suggestions are not supported for %s", spec.Label)
	}
	scopeRef, err := m.sourceScopeRef(spec, schema)
	if err != nil {
		return nil, err
	}
	return suggester.QuerySuggestions(context.Background(), scopeRef)
}
func (m *Manager) GetSSLStatus() (*engine.SSLStatus, error) {
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	reader, ok := session.(source.SecurityReader)
	if !ok {
		return nil, fmt.Errorf("SSL status is not supported for %s", spec.Label)
	}
	return reader.SSLStatus(context.Background())
}
func (m *Manager) GetSSLStatusSummary() (string, error) {
	status, err := m.GetSSLStatus()
	if err != nil {
		return "", err
	}
	return formatSSLStatusSummary(status), nil
}
func formatSSLStatusSummary(status *engine.SSLStatus) string {
	if status == nil {
		return ""
	}
	mode := strings.TrimSpace(status.Mode)
	if status.IsEnabled {
		if mode == "" || strings.EqualFold(mode, "enabled") {
			return "SSL/TLS: enabled"
		}
		return fmt.Sprintf("SSL/TLS: enabled (%s)", mode)
	}
	if mode == "" {
		mode = "disabled"
	}
	return fmt.Sprintf("SSL/TLS: %s", mode)
}
func (m *Manager) loadSavedConnections() []Connection {
	if m.config == nil {
		return nil
	}
	return m.config.Connections
}
func (m *Manager) getEnvConnections() []Connection {
	return connresolver.EnvConnections()
}
func (m *Manager) GetSchemas() ([]string, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if cached, ok := m.cache.GetSchemas(); ok {
		return cached, nil
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	start := time.Now()
	objects, err := m.listNamespaceObjects(context.Background(), spec, session)
	schemas := make([]string, 0, len(objects))
	for _, object := range objects {
		schemas = append(schemas, object.Name)
	}
	m.logOperation("GetSchemas()", start, len(schemas), err)
	if err != nil {
		return nil, err
	}
	m.cache.SetSchemas(schemas)
	return schemas, nil
}
func (m *Manager) GetStorageUnits(schema string) ([]engine.StorageUnit, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if cached, ok := m.cache.GetTables(schema); ok {
		return cached, nil
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	start := time.Now()
	objects, err := m.listStorageUnitObjects(context.Background(), spec, session, schema)
	tables := storageUnitsFromSourceObjects(objects)
	m.logOperation(fmt.Sprintf("GetStorageUnits(%s)", schema), start, len(tables), err)
	if err != nil {
		return nil, err
	}
	m.cache.SetTables(schema, tables)
	return tables, nil
}
func (m *Manager) GetGraph(schema string) ([]engine.GraphUnit, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	reader, ok := session.(source.GraphReader)
	if !ok {
		return nil, fmt.Errorf("graph views are not supported for %s", spec.Label)
	}
	scopeRef, err := m.sourceScopeRef(spec, schema)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	graphUnits, err := reader.ReadGraph(context.Background(), scopeRef)
	m.logOperation(fmt.Sprintf("GetGraph(%s)", schema), start, len(graphUnits), err)
	if err != nil {
		return nil, err
	}
	return graphUnits, nil
}
func (m *Manager) ExecuteQuery(query string) (*engine.GetRowsResult, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if m.config.GetReadOnly() && IsMutationQuery(query) {
		return nil, ErrReadOnly
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	runner, ok := session.(source.QueryRunner)
	if !ok {
		return nil, fmt.Errorf("querying is not supported for %s", spec.Label)
	}
	start := time.Now()
	result, err := runner.RunQuery(context.Background(), query)
	m.logQuery(query, start, result, err)
	return result, err
}
func (m *Manager) GetRows(schema, storageUnit string, where *model.WhereCondition, pageSize, pageOffset int) (*engine.GetRowsResult, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	reader, ok := session.(source.TabularReader)
	if !ok {
		return nil, fmt.Errorf("viewing rows is not supported for %s", spec.Label)
	}
	ref, err := m.storageUnitRef(spec, schema, storageUnit)
	if err != nil {
		return nil, err
	}
	query := fmt.Sprintf("GetRows(%s.%s, page=%d, offset=%d)", schema, storageUnit, pageSize, pageOffset)
	start := time.Now()
	result, err := reader.ReadRows(context.Background(), ref, queryWhereConditionFromModel(where), nil, pageSize, pageOffset)
	m.logQuery(query, start, result, err)
	return result, err
}
func runWithContext[T any](ctx context.Context, fn func() (T, error)) (T, error) {
	type result struct {
		data T
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		data, err := fn()
		ch <- result{data: data, err: err}
	}()
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case res := <-ch:
		return res.data, res.err
	}
}
func queryWhereConditionFromModel(condition *model.WhereCondition) *queryast.WhereCondition {
	if condition == nil {
		return nil
	}
	return &queryast.WhereCondition{
		Type:   queryast.WhereConditionType(condition.Type),
		Atomic: queryAtomicWhereConditionFromModel(condition.Atomic),
		And:    queryOperationWhereConditionFromModel(condition.And),
		Or:     queryOperationWhereConditionFromModel(condition.Or),
	}
}
func queryAtomicWhereConditionFromModel(condition *model.AtomicWhereCondition) *queryast.AtomicWhereCondition {
	if condition == nil {
		return nil
	}
	return &queryast.AtomicWhereCondition{
		ColumnType: condition.ColumnType,
		Key:        condition.Key,
		Operator:   condition.Operator,
		Value:      condition.Value,
	}
}
func queryOperationWhereConditionFromModel(condition *model.OperationWhereCondition) *queryast.OperationWhereCondition {
	if condition == nil {
		return nil
	}
	children := make([]*queryast.WhereCondition, 0, len(condition.Children))
	for _, child := range condition.Children {
		children = append(children, queryWhereConditionFromModel(child))
	}
	return &queryast.OperationWhereCondition{
		Children: children,
	}
}
type countingQueryStreamWriter struct {
	writer   engine.QueryStreamWriter
	rowCount int
}
func (w *countingQueryStreamWriter) WriteColumns(columns []engine.Column) error {
	return w.writer.WriteColumns(columns)
}
func (w *countingQueryStreamWriter) WriteRow(row []string) error {
	w.rowCount++
	return w.writer.WriteRow(row)
}
func (m *Manager) ExecuteExplain(query string) (*engine.GetRowsResult, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	explainMode, ok := sourcetypes.ExplainMode(m.currentConnection.Type)
	if !ok {
		return nil, fmt.Errorf("explain is not supported for source type %s", m.currentConnection.Type)
	}
	var explainQuery string
	switch explainMode {
	case source.QueryExplainModeExplainAnalyze:
		explainQuery = "EXPLAIN ANALYZE " + query
	case source.QueryExplainModeExplainPipeline:
		explainQuery = "EXPLAIN PIPELINE " + query
	case source.QueryExplainModeExplain:
		explainQuery = "EXPLAIN " + query
	default:
		return nil, fmt.Errorf("unsupported explain mode %s for source type %s", explainMode, m.currentConnection.Type)
	}
	return m.ExecuteQuery(explainQuery)
}
func (m *Manager) ExecuteQueryWithContext(ctx context.Context, query string) (*engine.GetRowsResult, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if m.config.GetReadOnly() && IsMutationQuery(query) {
		return nil, ErrReadOnly
	}
	spec, session, err := m.currentSourceSession(ctx)
	if err != nil {
		return nil, err
	}
	runner, ok := session.(source.QueryRunner)
	if !ok {
		return nil, fmt.Errorf("querying is not supported for %s", spec.Label)
	}
	start := time.Now()
	result, err := runWithContext(ctx, func() (*engine.GetRowsResult, error) {
		return runner.RunQuery(ctx, query)
	})
	m.logQuery(query, start, result, err)
	return result, err
}
func (m *Manager) ExecuteQueryStream(ctx context.Context, query string, writer engine.QueryStreamWriter) (int, error) {
	if m.currentConnection == nil {
		return 0, fmt.Errorf("not connected to any database")
	}
	if writer == nil {
		return 0, fmt.Errorf("stream writer is required")
	}
	if m.config.GetReadOnly() && IsMutationQuery(query) {
		return 0, ErrReadOnly
	}
	spec, session, err := m.currentSourceSession(ctx)
	if err != nil {
		return 0, err
	}
	streamer, ok := session.(source.StreamQueryRunner)
	if !ok {
		return 0, fmt.Errorf("streaming queries are not supported for %s", spec.Label)
	}
	countingWriter := &countingQueryStreamWriter{writer: writer}
	start := time.Now()
	_, err = runWithContext(ctx, func() (int, error) {
		if err := streamer.RunQueryStream(ctx, query, &sourceQueryStreamWriterAdapter{writer: countingWriter}); err != nil {
			return 0, err
		}
		return countingWriter.rowCount, nil
	})
	m.logStreamedQuery(query, start, countingWriter.rowCount, err)
	return countingWriter.rowCount, err
}
func (m *Manager) ExecuteQueryWithParams(query string, params []any) (*engine.GetRowsResult, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if m.config.GetReadOnly() && IsMutationQuery(query) {
		return nil, ErrReadOnly
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	runner, ok := session.(source.QueryRunner)
	if !ok {
		return nil, fmt.Errorf("querying is not supported for %s", spec.Label)
	}
	start := time.Now()
	result, err := runner.RunQuery(context.Background(), query, params...)
	m.logQuery(query, start, result, err)
	return result, err
}
func (m *Manager) ExecuteQueryWithContextAndParams(ctx context.Context, query string, params []any) (*engine.GetRowsResult, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if m.config.GetReadOnly() && IsMutationQuery(query) {
		return nil, ErrReadOnly
	}
	spec, session, err := m.currentSourceSession(ctx)
	if err != nil {
		return nil, err
	}
	runner, ok := session.(source.QueryRunner)
	if !ok {
		return nil, fmt.Errorf("querying is not supported for %s", spec.Label)
	}
	start := time.Now()
	result, err := runWithContext(ctx, func() (*engine.GetRowsResult, error) {
		return runner.RunQuery(ctx, query, params...)
	})
	m.logQuery(query, start, result, err)
	return result, err
}
func (m *Manager) GetRowsWithContext(ctx context.Context, schema, storageUnit string, where *model.WhereCondition, pageSize, pageOffset int) (*engine.GetRowsResult, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	spec, session, err := m.currentSourceSession(ctx)
	if err != nil {
		return nil, err
	}
	reader, ok := session.(source.TabularReader)
	if !ok {
		return nil, fmt.Errorf("viewing rows is not supported for %s", spec.Label)
	}
	ref, err := m.storageUnitRef(spec, schema, storageUnit)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	result, err := runWithContext(ctx, func() (*engine.GetRowsResult, error) {
		return reader.ReadRows(ctx, ref, queryWhereConditionFromModel(where), nil, pageSize, pageOffset)
	})
	m.logQuery(fmt.Sprintf("GetRows(%s.%s, page=%d, offset=%d)", schema, storageUnit, pageSize, pageOffset), start, result, err)
	return result, err
}
func (m *Manager) GetSchemasWithContext(ctx context.Context) ([]string, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if cached, ok := m.cache.GetSchemas(); ok {
		return cached, nil
	}
	spec, session, err := m.currentSourceSession(ctx)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	result, err := runWithContext(ctx, func() ([]string, error) {
		objects, err := m.listNamespaceObjects(ctx, spec, session)
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(objects))
		for _, object := range objects {
			names = append(names, object.Name)
		}
		return names, nil
	})
	m.logOperation("GetSchemas()", start, len(result), err)
	if err == nil {
		m.cache.SetSchemas(result)
	}
	return result, err
}
func (m *Manager) GetStorageUnitsWithContext(ctx context.Context, schema string) ([]engine.StorageUnit, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if cached, ok := m.cache.GetTables(schema); ok {
		return cached, nil
	}
	spec, session, err := m.currentSourceSession(ctx)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	result, err := runWithContext(ctx, func() ([]engine.StorageUnit, error) {
		objects, err := m.listStorageUnitObjects(ctx, spec, session, schema)
		if err != nil {
			return nil, err
		}
		return storageUnitsFromSourceObjects(objects), nil
	})
	m.logOperation(fmt.Sprintf("GetStorageUnits(%s)", schema), start, len(result), err)
	if err == nil {
		m.cache.SetTables(schema, result)
	}
	return result, err
}
func (m *Manager) GetConfig() *config.Config {
	return m.config
}
func (m *Manager) GetColumnsWithContext(ctx context.Context, schema, storageUnit string) ([]engine.Column, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if cached, ok := m.cache.GetColumns(schema, storageUnit); ok {
		return cached, nil
	}
	spec, session, err := m.currentSourceSession(ctx)
	if err != nil {
		return nil, err
	}
	reader, ok := session.(source.TabularReader)
	if !ok {
		return nil, fmt.Errorf("inspecting objects is not supported for %s", spec.Label)
	}
	ref, err := m.storageUnitRef(spec, schema, storageUnit)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	result, err := runWithContext(ctx, func() ([]engine.Column, error) {
		return reader.Columns(ctx, ref)
	})
	m.logOperation(fmt.Sprintf("GetColumns(%s.%s)", schema, storageUnit), start, len(result), err)
	if err == nil {
		m.cache.SetColumns(schema, storageUnit, result)
	}
	return result, err
}
func (m *Manager) GetColumnsForStorageUnits(schema string, storageUnitNames []string) (map[string][]engine.Column, error) {
	return loadStorageUnitMetadata(
		storageUnitNames,
		func(name string) ([]engine.Column, error) {
			return m.GetColumns(schema, name)
		},
	)
}
func (m *Manager) GetColumns(schema, storageUnit string) ([]engine.Column, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	if cached, ok := m.cache.GetColumns(schema, storageUnit); ok {
		return cached, nil
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	reader, ok := session.(source.TabularReader)
	if !ok {
		return nil, fmt.Errorf("inspecting objects is not supported for %s", spec.Label)
	}
	ref, err := m.storageUnitRef(spec, schema, storageUnit)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	columns, err := reader.Columns(context.Background(), ref)
	m.logOperation(fmt.Sprintf("GetColumns(%s.%s)", schema, storageUnit), start, len(columns), err)
	if err != nil {
		return nil, err
	}
	m.cache.SetColumns(schema, storageUnit, columns)
	return columns, nil
}
func (m *Manager) GetColumnConstraints(schema, storageUnit string) (map[string]map[string]any, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	reader, ok := session.(source.ColumnConstraintReader)
	if !ok {
		return nil, fmt.Errorf("column constraints are not supported for %s", spec.Label)
	}
	ref, err := m.storageUnitRef(spec, schema, storageUnit)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	constraints, err := reader.ColumnConstraints(context.Background(), ref)
	m.logOperation(fmt.Sprintf("GetColumnConstraints(%s.%s)", schema, storageUnit), start, len(constraints), err)
	if err != nil {
		return nil, err
	}
	return constraints, nil
}
func (m *Manager) GetColumnConstraintsForStorageUnits(schema string, storageUnitNames []string) (map[string]map[string]map[string]any, error) {
	return loadStorageUnitMetadata(
		storageUnitNames,
		func(name string) (map[string]map[string]any, error) {
			return m.GetColumnConstraints(schema, name)
		},
	)
}
func loadStorageUnitMetadata[T any](storageUnitNames []string, load func(string) (T, error)) (map[string]T, error) {
	const maxConcurrentMetadataLoads = 6
	results := make(map[string]T, len(storageUnitNames))
	if len(storageUnitNames) == 0 {
		return results, nil
	}
	var mu sync.Mutex
	group := new(errgroup.Group)
	group.SetLimit(maxConcurrentMetadataLoads)
	for _, storageUnitName := range storageUnitNames {
		storageUnitName := storageUnitName
		group.Go(func() error {
			value, err := load(storageUnitName)
			if err != nil {
				return fmt.Errorf("%s: %w", storageUnitName, err)
			}
			mu.Lock()
			results[storageUnitName] = value
			mu.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}
func (m *Manager) ExportToCSV(schema, storageUnit, filename, delimiter string) error {
	start := time.Now()
	defer func() {
		m.logOperation(fmt.Sprintf("ExportToCSV(%s.%s → %s)", schema, storageUnit, filepath.Base(filename)), start, 0, nil)
	}()
	if m.currentConnection == nil {
		return fmt.Errorf("not connected to any database")
	}
	result, err := m.GetRows(schema, storageUnit, nil, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}
	dir := filepath.Dir(filename)
	tmp, err := os.CreateTemp(dir, ".whodb-export-*.csv")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	writer := csv.NewWriter(tmp)
	delimRune := rune(delimiter[0])
	writer.Comma = delimRune
	headers := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		headers[i] = col.Name
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}
	for _, row := range result.Rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}
	_ = tmp.Sync()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, filename); err != nil {
		_ = os.Remove(filename)
		if err2 := os.Rename(tmpPath, filename); err2 != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to save file: %w", err2)
		}
	}
	syncDir(filepath.Dir(filename))
	_ = os.Chmod(filename, 0600)
	return nil
}
func (m *Manager) ExportDataStream(schema, storageUnit string, writer func([]string) error) (int, error) {
	start := time.Now()
	if m.currentConnection == nil {
		return 0, fmt.Errorf("not connected to any database")
	}
	if writer == nil {
		return 0, fmt.Errorf("writer is required")
	}
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return 0, err
	}
	exporter, ok := session.(source.TabularExporter)
	if !ok {
		return 0, fmt.Errorf("exporting rows is not supported for %s", spec.Label)
	}
	ref, err := m.storageUnitRef(spec, schema, storageUnit)
	if err != nil {
		return 0, err
	}
	headerWritten := false
	rowCount := 0
	err = exporter.ExportRows(context.Background(), ref, func(record []string) error {
		if headerWritten {
			rowCount++
		} else {
			headerWritten = true
		}
		return writer(record)
	}, nil)
	m.logOperation(fmt.Sprintf("ExportDataStream(%s.%s)", schema, storageUnit), start, rowCount, err)
	if err != nil {
		return 0, err
	}
	return rowCount, nil
}
func (m *Manager) ExportToExcel(schema, storageUnit, filename string) error {
	start := time.Now()
	defer func() {
		m.logOperation(fmt.Sprintf("ExportToExcel(%s.%s → %s)", schema, storageUnit, filepath.Base(filename)), start, 0, nil)
	}()
	if m.currentConnection == nil {
		return fmt.Errorf("not connected to any database")
	}
	result, err := m.GetRows(schema, storageUnit, nil, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}
	f := excelize.NewFile()
	defer f.Close()
	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}
	for i, col := range result.Columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, col.Name)
	}
	for rowIdx, row := range result.Rows {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}
	f.SetActiveSheet(index)
	dir := filepath.Dir(filename)
	tmp, err := os.CreateTemp(dir, ".whodb-export-*.xlsx")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	if err := f.SaveAs(tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to save file: %w", err)
	}
	if tf, err := os.OpenFile(tmpPath, os.O_RDWR, 0); err == nil {
		_ = tf.Sync()
		_ = tf.Close()
	}
	if err := os.Rename(tmpPath, filename); err != nil {
		_ = os.Remove(filename)
		if err2 := os.Rename(tmpPath, filename); err2 != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to save file: %w", err2)
		}
	}
	syncDir(filepath.Dir(filename))
	_ = os.Chmod(filename, 0600)
	return nil
}
func (m *Manager) ExportResultsToCSV(result *engine.GetRowsResult, filename, delimiter string) error {
	start := time.Now()
	defer func() {
		m.logOperation(fmt.Sprintf("ExportResultsToCSV(→ %s)", filepath.Base(filename)), start, 0, nil)
	}()
	if result == nil {
		return fmt.Errorf("no results to export")
	}
	dir := filepath.Dir(filename)
	tmp, err := os.CreateTemp(dir, ".whodb-export-*.csv")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	writer := csv.NewWriter(tmp)
	delimRune := rune(delimiter[0])
	writer.Comma = delimRune
	headers := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		headers[i] = col.Name
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}
	for _, row := range result.Rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to flush CSV writer: %w", err)
	}
	_ = tmp.Sync()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Rename(tmpPath, filename); err != nil {
		_ = os.Remove(filename)
		if err2 := os.Rename(tmpPath, filename); err2 != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to save file: %w", err2)
		}
	}
	syncDir(filepath.Dir(filename))
	_ = os.Chmod(filename, 0600)
	return nil
}
func (m *Manager) ExportResultsToExcel(result *engine.GetRowsResult, filename string) error {
	start := time.Now()
	defer func() {
		m.logOperation(fmt.Sprintf("ExportResultsToExcel(→ %s)", filepath.Base(filename)), start, 0, nil)
	}()
	if result == nil {
		return fmt.Errorf("no results to export")
	}
	f := excelize.NewFile()
	defer f.Close()
	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create sheet: %w", err)
	}
	for i, col := range result.Columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, col.Name)
	}
	for rowIdx, row := range result.Rows {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}
	f.SetActiveSheet(index)
	dir := filepath.Dir(filename)
	tmp, err := os.CreateTemp(dir, ".whodb-export-*.xlsx")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	if err := f.SaveAs(tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to save file: %w", err)
	}
	if tf, err := os.OpenFile(tmpPath, os.O_RDWR, 0); err == nil {
		_ = tf.Sync()
		_ = tf.Close()
	}
	if err := os.Rename(tmpPath, filename); err != nil {
		_ = os.Remove(filename)
		if err2 := os.Rename(tmpPath, filename); err2 != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to save file: %w", err2)
		}
	}
	syncDir(filepath.Dir(filename))
	_ = os.Chmod(filename, 0600)
	return nil
}
func syncDir(dir string) {
	if dir == "" || dir == "." {
		return
	}
	if f, err := os.Open(dir); err == nil {
		_ = f.Sync()
		_ = f.Close()
	}
}
type AIProvider struct {
	Type       string
	ProviderId string
}
func (m *Manager) GetAIProviders() []AIProvider {
	providers := envconfig.GetConfiguredChatProviders()
	aiProviders := []AIProvider{}
	for _, provider := range providers {
		aiProviders = append(aiProviders, AIProvider{
			Type:       provider.Type,
			ProviderId: provider.ProviderId,
		})
	}
	return aiProviders
}
func (m *Manager) GetAIModels(providerID, modelType, token string) ([]string, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	externalModel := &engine.ExternalModel{
		Type: modelType,
	}
	if providerID != "" {
		providers := envconfig.GetConfiguredChatProviders()
		for _, provider := range providers {
			if provider.ProviderId == providerID {
				externalModel.Token = provider.APIKey
				break
			}
		}
	} else if token != "" {
		externalModel.Token = token
	}
	return llm.ClientForModel(externalModel).GetSupportedModels()
}
func (m *Manager) GetAIModelsWithContext(ctx context.Context, providerID, modelType, token string) ([]string, error) {
	return runWithContext(ctx, func() ([]string, error) {
		return m.GetAIModels(providerID, modelType, token)
	})
}
type ChatMessage struct {
	Type                 string
	Result               *engine.GetRowsResult
	Text                 string
	RequiresConfirmation bool
}
type StreamChunk struct {
	Text    string
	IsFinal bool
	Final   []*ChatMessage
	Err     error
}
func resolveExternalModel(providerID, modelType, token, model string) *engine.ExternalModel {
	externalModel := &engine.ExternalModel{
		Type:  modelType,
		Model: model,
	}
	if providerID != "" {
		providers := envconfig.GetConfiguredChatProviders()
		for _, provider := range providers {
			if provider.ProviderId == providerID {
				externalModel.Token = provider.APIKey
				return externalModel
			}
		}
	}
	if token != "" {
		externalModel.Token = token
	}
	return externalModel
}
func (m *Manager) SendAIChat(providerID, modelType, token, schema, model, previousConversation, query string) ([]*ChatMessage, error) {
	if m.currentConnection == nil {
		return nil, fmt.Errorf("not connected to any database")
	}
	baml.Ensure()
	spec, session, err := m.currentSourceSession(context.Background())
	if err != nil {
		return nil, err
	}
	scopeRef, err := m.sourceScopeRef(spec, schema)
	if err != nil {
		return nil, err
	}
	if modelSelection := resolveExternalModel(providerID, modelType, token, model); modelSelection != nil && modelSelection.Model != "" {
		if assistant, ok := session.(source.ModelAwareSourceAssistant); ok {
			messages, err := assistant.ReplyWithModel(context.Background(), scopeRef, previousConversation, query, modelSelection)
			if err != nil {
				return nil, err
			}
			return cliChatMessagesFromSource(messages), nil
		}
	}
	assistant, ok := session.(source.SourceAssistant)
	if !ok {
		return nil, fmt.Errorf("chat is not supported for %s", spec.Label)
	}
	messages, err := assistant.Reply(context.Background(), scopeRef, previousConversation, query)
	if err != nil {
		return nil, err
	}
	return cliChatMessagesFromSource(messages), nil
}
func (m *Manager) SendAIChatWithContext(ctx context.Context, providerID, modelType, token, schema, model, previousConversation, query string) ([]*ChatMessage, error) {
	return runWithContext(ctx, func() ([]*ChatMessage, error) {
		return m.SendAIChat(providerID, modelType, token, schema, model, previousConversation, query)
	})
}