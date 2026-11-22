package database
import (
	"context"
	"errors"
	"testing"
)
func TestIsMutationQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{"insert", "INSERT INTO users (name) VALUES ('alice')", true},
		{"insert lowercase", "insert into users (name) values ('alice')", true},
		{"insert mixed case", "Insert Into users (name) values ('alice')", true},
		{"insert with leading spaces", "  INSERT INTO users (name) VALUES ('alice')", true},
		{"insert with leading newline", "\n INSERT INTO users (name) VALUES ('alice')", true},
		{"insert with tab", "\tINSERT INTO users (name) VALUES ('alice')", true},
		{"update", "UPDATE users SET name = 'bob'", true},
		{"delete", "DELETE FROM users WHERE id = 1", true},
		{"drop table", "DROP TABLE users", true},
		{"drop database", "DROP DATABASE test", true},
		{"alter table", "ALTER TABLE users ADD COLUMN age INT", true},
		{"create table", "CREATE TABLE users (id INT)", true},
		{"create index", "CREATE INDEX idx ON users(name)", true},
		{"truncate", "TRUNCATE TABLE users", true},
		{"grant", "GRANT SELECT ON users TO reader", true},
		{"revoke", "REVOKE SELECT ON users FROM reader", true},
		{"select", "SELECT * FROM users", false},
		{"select lowercase", "select * from users", false},
		{"select with leading spaces", "  SELECT * FROM users", false},
		{"show", "SHOW TABLES", false},
		{"describe", "DESCRIBE users", false},
		{"explain", "EXPLAIN SELECT * FROM users", false},
		{"pragma", "PRAGMA table_info(users)", false},
		{"with cte", "WITH cte AS (SELECT 1) SELECT * FROM cte", false},
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"single word select", "SELECT", false},
		{"semicolon only keyword", "INSERT;", true},
		{"paren after keyword", "CREATE(", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMutationQuery(tt.query)
			if got != tt.expected {
				t.Errorf("IsMutationQuery(%q) = %v, want %v", tt.query, got, tt.expected)
			}
		})
	}
}
func TestExecuteQuery_ReadOnlyBlocks(t *testing.T) {
	setupTestEnv(t)
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	mgr.currentConnection = &Connection{
		Name: "test",
		Type: "postgres",
		Host: "localhost",
	}
	mgr.config.SetReadOnly(true)
	mutations := []string{
		"INSERT INTO users (name) VALUES ('test')",
		"UPDATE users SET name = 'test'",
		"DELETE FROM users WHERE id = 1",
		"DROP TABLE users",
		"ALTER TABLE users ADD COLUMN age INT",
		"CREATE TABLE test (id INT)",
		"TRUNCATE TABLE users",
		"GRANT SELECT ON users TO reader",
		"REVOKE SELECT ON users FROM reader",
	}
	for _, query := range mutations {
		_, err := mgr.ExecuteQuery(query)
		if !errors.Is(err, ErrReadOnly) {
			t.Errorf("ExecuteQuery(%q) with read-only ON: got err=%v, want ErrReadOnly", query, err)
		}
	}
}
func TestExecuteQuery_ReadOnlyAllowsSelects(t *testing.T) {
	setupTestEnv(t)
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	mgr.currentConnection = &Connection{
		Name: "test",
		Type: "postgres",
		Host: "localhost",
	}
	mgr.config.SetReadOnly(true)
	_, err = mgr.ExecuteQuery("SELECT 1")
	if errors.Is(err, ErrReadOnly) {
		t.Error("ExecuteQuery(SELECT) should not be blocked by read-only mode")
	}
}
func TestExecuteQuery_ReadOnlyOffAllowsMutations(t *testing.T) {
	setupTestEnv(t)
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	mgr.currentConnection = &Connection{
		Name: "test",
		Type: "postgres",
		Host: "localhost",
	}
	mgr.config.SetReadOnly(false)
	_, err = mgr.ExecuteQuery("INSERT INTO users (name) VALUES ('test')")
	if errors.Is(err, ErrReadOnly) {
		t.Error("ExecuteQuery(INSERT) should not be blocked when read-only is OFF")
	}
}
func TestExecuteQueryWithContext_ReadOnlyBlocks(t *testing.T) {
	setupTestEnv(t)
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	mgr.currentConnection = &Connection{
		Name: "test",
		Type: "postgres",
		Host: "localhost",
	}
	mgr.config.SetReadOnly(true)
	ctx := context.Background()
	_, err = mgr.ExecuteQueryWithContext(ctx, "DELETE FROM users")
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("ExecuteQueryWithContext(DELETE) with read-only ON: got err=%v, want ErrReadOnly", err)
	}
}
func TestExecuteQueryWithParams_ReadOnlyBlocks(t *testing.T) {
	setupTestEnv(t)
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	mgr.currentConnection = &Connection{
		Name: "test",
		Type: "postgres",
		Host: "localhost",
	}
	mgr.config.SetReadOnly(true)
	_, err = mgr.ExecuteQueryWithParams("UPDATE users SET name = $1", []any{"test"})
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("ExecuteQueryWithParams(UPDATE) with read-only ON: got err=%v, want ErrReadOnly", err)
	}
}
func TestExecuteQueryWithContextAndParams_ReadOnlyBlocks(t *testing.T) {
	setupTestEnv(t)
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	mgr.currentConnection = &Connection{
		Name: "test",
		Type: "postgres",
		Host: "localhost",
	}
	mgr.config.SetReadOnly(true)
	ctx := context.Background()
	_, err = mgr.ExecuteQueryWithContextAndParams(ctx, "DROP TABLE users", nil)
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("ExecuteQueryWithContextAndParams(DROP) with read-only ON: got err=%v, want ErrReadOnly", err)
	}
}
func TestImportData_ReadOnlyBlocks(t *testing.T) {
	setupTestEnv(t)
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	mgr.currentConnection = &Connection{
		Name: "test",
		Type: "postgres",
		Host: "localhost",
	}
	mgr.config.SetReadOnly(true)
	_, err = mgr.ImportData("public", "users", []string{"name"}, [][]string{{"alice"}}, ImportOptions{})
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("ImportData with read-only ON: got err=%v, want ErrReadOnly", err)
	}
}
func TestGenerateMockData_ReadOnlyBlocks(t *testing.T) {
	setupTestEnv(t)
	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	mgr.currentConnection = &Connection{
		Name: "test",
		Type: "postgres",
		Host: "localhost",
	}
	mgr.config.SetReadOnly(true)
	_, err = mgr.GenerateMockData("public", "users", 10, false, 0)
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("GenerateMockData with read-only ON: got err=%v, want ErrReadOnly", err)
	}
}