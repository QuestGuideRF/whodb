package e2e_test
import (
	"path/filepath"
	"testing"
	"github.com/clidey/whodb/cli/e2e/testharness"
)
func TestPostgres_MultipleConnections_List(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	connections := map[string]testharness.PostgresConfig{
		"primary":   cfg,
		"secondary": cfg,
		"readonly":  cfg,
	}
	cleanup := testharness.SetupMultipleConnections(t, connections)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"connections", "list",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "primary")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "secondary")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "readonly")
}
func TestPostgres_MultipleConnections_QueryPrimary(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	connections := map[string]testharness.PostgresConfig{
		"primary":   cfg,
		"secondary": cfg,
	}
	cleanup := testharness.SetupMultipleConnections(t, connections)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--connection", "primary",
		"--format", "json",
		"--quiet",
		"SELECT COUNT(*) as count FROM test_schema.users",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONContains(t, stdout, "count")
}
func TestPostgres_MultipleConnections_QuerySecondary(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	connections := map[string]testharness.PostgresConfig{
		"primary":   cfg,
		"secondary": cfg,
	}
	cleanup := testharness.SetupMultipleConnections(t, connections)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--connection", "secondary",
		"--format", "json",
		"--quiet",
		"SELECT COUNT(*) as count FROM test_schema.products",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONContains(t, stdout, "count")
}
func TestPostgres_MultipleConnections_SwitchBetween(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	connections := map[string]testharness.PostgresConfig{
		"db1": cfg,
		"db2": cfg,
	}
	cleanup := testharness.SetupMultipleConnections(t, connections)
	defer cleanup()
	stdout1, stderr1, exitCode1 := testharness.RunCLI(t,
		"query",
		"--connection", "db1",
		"--format", "json",
		"--quiet",
		"SELECT username FROM test_schema.users WHERE id = 1",
	)
	testharness.RequireSuccess(t, stderr1, exitCode1)
	testharness.AssertJSONArrayContainsValue(t, stdout1, "username", "john_doe")
	stdout2, stderr2, exitCode2 := testharness.RunCLI(t,
		"query",
		"--connection", "db2",
		"--format", "json",
		"--quiet",
		"SELECT name FROM test_schema.products WHERE id = 1",
	)
	testharness.RequireSuccess(t, stderr2, exitCode2)
	testharness.AssertJSONArrayContainsValue(t, stdout2, "name", "Laptop")
	stdout3, stderr3, exitCode3 := testharness.RunCLI(t,
		"query",
		"--connection", "db1",
		"--format", "json",
		"--quiet",
		"SELECT email FROM test_schema.users WHERE id = 2",
	)
	testharness.RequireSuccess(t, stderr3, exitCode3)
	testharness.AssertJSONArrayContainsValue(t, stdout3, "email", "jane@example.com")
}
func TestPostgres_ConnectionTest_Success(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	_, stderr, exitCode := testharness.RunCLI(t,
		"connections", "test", "test-pg",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
}
func TestPostgres_ConnectionTest_NonexistentConnection(t *testing.T) {
	cleanup := testharness.SetupCleanEnv(t)
	defer cleanup()
	_, stderr, exitCode := testharness.RunCLI(t,
		"connections", "test", "nonexistent-connection",
	)
	testharness.RequireFailure(t, exitCode)
	testharness.AssertContains(t, stderr, "connection")
}
func TestPostgres_ConnectionAdd_Remove(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	_, stderr, exitCode := testharness.RunCLI(t,
		"connections", "add",
		"--name", "my-new-connection",
		"--type", "Postgres",
		"--host", cfg.Host,
		"--port", cfg.Port,
		"--user", cfg.User,
		"--password", cfg.Password,
		"--database", cfg.Database,
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"connections", "list",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "my-new-connection")
	_, stderr, exitCode = testharness.RunCLI(t,
		"connections", "remove", "my-new-connection",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	stdout, stderr, exitCode = testharness.RunCLI(t,
		"connections", "list",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertNotContains(t, stdout, "my-new-connection")
}
func TestPostgres_ConnectionAdd_MissingRequired(t *testing.T) {
	cleanup := testharness.SetupCleanEnv(t)
	defer cleanup()
	_, stderr, exitCode := testharness.RunCLI(t,
		"connections", "add",
		"--type", "Postgres",
		"--host", "localhost",
		"--database", "testdb",
	)
	testharness.RequireFailure(t, exitCode)
	testharness.AssertContains(t, stderr, "--name is required")
}
func TestPostgres_ConnectionRemove_EnvConnection(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	_, stderr, exitCode := testharness.RunCLI(t,
		"connections", "remove", "test-pg",
	)
	testharness.RequireFailure(t, exitCode)
	testharness.AssertContains(t, stderr, "environment variables")
}
func TestPostgres_ConnectionsList_ShowsSource(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"connections", "list",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "source", "env")
}
func TestPostgres_Query_UsesFirstConnectionByDefault(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--format", "json",
		"--quiet",
		"SELECT 1 as result",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONContains(t, stdout, "result")
}
func TestPostgres_Schemas_UsesFirstConnectionByDefault(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"schemas",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "schema", "test_schema")
}
func TestPostgres_Export_WithSpecificConnection(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	connections := map[string]testharness.PostgresConfig{
		"export-source": cfg,
		"other-db":      cfg,
	}
	cleanup := testharness.SetupMultipleConnections(t, connections)
	defer cleanup()
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "export_test.csv")
	_, stderr, exitCode := testharness.RunCLI(t,
		"export",
		"--connection", "export-source",
		"--schema", "test_schema",
		"--table", "products",
		"--output", csvPath,
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertFileExists(t, csvPath)
	testharness.AssertFileContains(t, csvPath, "Laptop")
}
func TestPostgres_Tables_MultipleConnections(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	connections := map[string]testharness.PostgresConfig{
		"db-alpha": cfg,
		"db-beta":  cfg,
	}
	cleanup := testharness.SetupMultipleConnections(t, connections)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"tables",
		"--connection", "db-alpha",
		"--schema", "test_schema",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "users")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "orders")
}
func TestPostgres_Columns_MultipleConnections(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	connections := map[string]testharness.PostgresConfig{
		"conn-1": cfg,
		"conn-2": cfg,
	}
	cleanup := testharness.SetupMultipleConnections(t, connections)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"columns",
		"--connection", "conn-2",
		"--schema", "test_schema",
		"--table", "orders",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "user_id")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "status")
}