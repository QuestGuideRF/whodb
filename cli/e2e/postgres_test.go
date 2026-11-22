package e2e_test
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/clidey/whodb/cli/e2e/testharness"
)
func TestPostgres_Schemas(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"schemas",
		"--connection", "test-pg",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "schema", "test_schema")
}
func TestPostgres_Tables(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"tables",
		"--connection", "test-pg",
		"--schema", "test_schema",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "users")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "products")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "orders")
}
func TestPostgres_Columns(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"columns",
		"--connection", "test-pg",
		"--schema", "test_schema",
		"--table", "users",
		"--format", "json",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "id")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "username")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "email")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "password")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "created_at")
}
func TestPostgres_Query_Select(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--connection", "test-pg",
		"--format", "json",
		"--quiet",
		"SELECT * FROM test_schema.users ORDER BY id",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayLength(t, stdout, 3)
	testharness.AssertJSONArrayContainsValue(t, stdout, "username", "john_doe")
	testharness.AssertJSONArrayContainsValue(t, stdout, "username", "jane_smith")
	testharness.AssertJSONArrayContainsValue(t, stdout, "username", "admin_user")
}
func TestPostgres_Query_Where(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--connection", "test-pg",
		"--format", "json",
		"--quiet",
		"SELECT * FROM test_schema.users WHERE username='john_doe'",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayLength(t, stdout, 1)
	testharness.AssertJSONArrayContainsValue(t, stdout, "username", "john_doe")
}
func TestPostgres_Query_Join(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--connection", "test-pg",
		"--format", "json",
		"--quiet",
		"SELECT u.username, o.status FROM test_schema.users u JOIN test_schema.orders o ON u.id = o.user_id ORDER BY u.username",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayMinLength(t, stdout, 2)
	testharness.AssertJSONContains(t, stdout, "username")
	testharness.AssertJSONContains(t, stdout, "status")
}
func TestPostgres_Query_InvalidSQL(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	_, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--connection", "test-pg",
		"--quiet",
		"INVALID SQL SYNTAX HERE",
	)
	testharness.RequireFailure(t, exitCode)
	testharness.AssertContains(t, stderr, "query failed")
}
func TestPostgres_Export_CSV(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "users.csv")
	_, stderr, exitCode := testharness.RunCLI(t,
		"export",
		"--connection", "test-pg",
		"--schema", "test_schema",
		"--table", "users",
		"--output", csvPath,
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertFileExists(t, csvPath)
	testharness.AssertFileNotEmpty(t, csvPath)
	testharness.AssertFileContains(t, csvPath, "id")
	testharness.AssertFileContains(t, csvPath, "username")
	testharness.AssertFileContains(t, csvPath, "email")
}
func TestPostgres_Export_Excel(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	tempDir := t.TempDir()
	xlsxPath := filepath.Join(tempDir, "users.xlsx")
	_, stderr, exitCode := testharness.RunCLI(t,
		"export",
		"--connection", "test-pg",
		"--schema", "test_schema",
		"--table", "users",
		"--format", "excel",
		"--output", xlsxPath,
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertFileExists(t, xlsxPath)
	testharness.AssertFileNotEmpty(t, xlsxPath)
}
func TestPostgres_Export_QueryToCSV(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "query_results.csv")
	_, stderr, exitCode := testharness.RunCLI(t,
		"export",
		"--connection", "test-pg",
		"--query", "SELECT username, email FROM test_schema.users WHERE username LIKE 'john%'",
		"--output", csvPath,
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertFileExists(t, csvPath)
	testharness.AssertFileNotEmpty(t, csvPath)
	testharness.AssertFileContains(t, csvPath, "john_doe")
}
func TestPostgres_Columns_NonexistentTable(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	_, stderr, exitCode := testharness.RunCLI(t,
		"columns",
		"--connection", "test-pg",
		"--schema", "test_schema",
		"--table", "nonexistent_table_xyz",
		"--quiet",
	)
	testharness.RequireFailure(t, exitCode)
	testharness.AssertContains(t, stderr, "failed")
}
func TestPostgres_Query_Products(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--connection", "test-pg",
		"--format", "json",
		"--quiet",
		"SELECT name, price FROM test_schema.products ORDER BY price DESC",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayLength(t, stdout, 4)
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "Laptop")
	testharness.AssertJSONArrayContainsValue(t, stdout, "name", "Smartphone")
}
func TestPostgres_Query_OrderItems(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	stdout, stderr, exitCode := testharness.RunCLI(t,
		"query",
		"--connection", "test-pg",
		"--format", "json",
		"--quiet",
		"SELECT oi.quantity, p.name FROM test_schema.order_items oi JOIN test_schema.products p ON oi.product_id = p.id",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertJSONArrayLength(t, stdout, 3)
}
func TestPostgres_Export_CustomDelimiter(t *testing.T) {
	cfg := testharness.DefaultPostgresConfig()
	cleanup := testharness.SetupEnv(t, cfg)
	defer cleanup()
	tempDir := t.TempDir()
	csvPath := filepath.Join(tempDir, "users_semicolon.csv")
	_, stderr, exitCode := testharness.RunCLI(t,
		"export",
		"--connection", "test-pg",
		"--schema", "test_schema",
		"--table", "users",
		"--output", csvPath,
		"--delimiter", ";",
		"--quiet",
	)
	testharness.RequireSuccess(t, stderr, exitCode)
	testharness.AssertFileExists(t, csvPath)
	testharness.AssertFileNotEmpty(t, csvPath)
	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	testharness.AssertContains(t, string(content), ";")
}