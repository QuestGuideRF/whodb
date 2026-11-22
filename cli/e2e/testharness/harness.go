package testharness
import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	Schema   string
}
func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "jio53$*(@nfe)",
		Database: "test_db",
		Schema:   "test_schema",
	}
}
func SetupEnv(t *testing.T, cfg PostgresConfig) func() {
	t.Helper()
	connJSON := fmt.Sprintf(`[{"alias":"test-pg","host":"%s","user":"%s","password":"%s","database":"%s","port":"%s"}]`,
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
	origPostgres := os.Getenv("WHODB_POSTGRES")
	origHome := os.Getenv("HOME")
	tempHome, err := os.MkdirTemp("", "whodb-cli-e2e-")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	os.Setenv("WHODB_POSTGRES", connJSON)
	os.Setenv("HOME", tempHome)
	os.Setenv("USERPROFILE", tempHome)
	os.Setenv("WHODB_CLI_ANALYTICS_DISABLED", "true")
	return func() {
		if origPostgres != "" {
			os.Setenv("WHODB_POSTGRES", origPostgres)
		} else {
			os.Unsetenv("WHODB_POSTGRES")
		}
		if origHome != "" {
			os.Setenv("HOME", origHome)
			os.Setenv("USERPROFILE", origHome)
		}
		os.Unsetenv("WHODB_CLI_ANALYTICS_DISABLED")
		os.RemoveAll(tempHome)
	}
}
func CLIBinaryPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get caller information")
	}
	cliDir := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	binaryName := "whodb-cli"
	if runtime.GOOS == "windows" {
		binaryName = "whodb-cli.exe"
	}
	binaryPath := filepath.Join(cliDir, binaryName)
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("CLI binary not found at %s. Run 'go build -o whodb-cli .' in the cli directory first.", binaryPath)
	}
	return binaryPath
}
func RunCLI(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	binaryPath := CLIBinaryPath(t)
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = os.Environ()
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run CLI: %v", err)
		}
	}
	return stdout, stderr, exitCode
}
func RequireSuccess(t *testing.T, stderr string, exitCode int) {
	t.Helper()
	if exitCode != 0 {
		t.Fatalf("CLI command failed with exit code %d\nStderr: %s", exitCode, stderr)
	}
}
func RequireFailure(t *testing.T, exitCode int) {
	t.Helper()
	if exitCode == 0 {
		t.Fatal("Expected CLI command to fail, but it succeeded")
	}
}
func ParseJSONArray(t *testing.T, stdout string) []map[string]any {
	t.Helper()
	var result []map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON array: %v\nOutput: %s", err, stdout)
	}
	return result
}
func SetupCleanEnv(t *testing.T) func() {
	t.Helper()
	dbPrefixes := []string{
		"WHODB_POSTGRES", "WHODB_MYSQL", "WHODB_MARIADB", "WHODB_MONGODB",
		"WHODB_REDIS", "WHODB_CLICKHOUSE", "WHODB_ELASTICSEARCH", "WHODB_SQLITE3",
	}
	origValues := make(map[string]string)
	origValues["HOME"] = os.Getenv("HOME")
	origValues["USERPROFILE"] = os.Getenv("USERPROFILE")
	for _, env := range os.Environ() {
		for _, prefix := range dbPrefixes {
			if strings.HasPrefix(env, prefix) {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					origValues[parts[0]] = parts[1]
				}
			}
		}
	}
	tempHome, err := os.MkdirTemp("", "whodb-cli-e2e-clean-")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	for _, env := range os.Environ() {
		for _, prefix := range dbPrefixes {
			if strings.HasPrefix(env, prefix) {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) >= 1 {
					os.Unsetenv(parts[0])
				}
			}
		}
	}
	origHome := origValues["HOME"]
	if origHome != "" {
		possibleCacheDirs := []string{
			filepath.Join(origHome, ".cache", "baml", "libs"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "baml", "libs"),
			filepath.Join(origValues["USERPROFILE"], ".cache", "baml", "libs"),
		}
		for _, cacheDir := range possibleCacheDirs {
			if cacheDir == "" {
				continue
			}
			entries, err := os.ReadDir(cacheDir)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if entry.IsDir() {
					versionDir := filepath.Join(cacheDir, entry.Name())
					files, err := os.ReadDir(versionDir)
					if err != nil {
						continue
					}
					for _, f := range files {
						name := f.Name()
						if strings.HasPrefix(name, "libbaml") && (strings.HasSuffix(name, ".so") || strings.HasSuffix(name, ".dll") || strings.HasSuffix(name, ".dylib")) {
							os.Setenv("BAML_LIBRARY_PATH", filepath.Join(versionDir, name))
							break
						}
					}
				}
			}
		}
	}
	os.Setenv("HOME", tempHome)
	os.Setenv("USERPROFILE", tempHome)
	os.Setenv("WHODB_CLI_ANALYTICS_DISABLED", "true")
	return func() {
		for _, env := range os.Environ() {
			for _, prefix := range dbPrefixes {
				if strings.HasPrefix(env, prefix) {
					parts := strings.SplitN(env, "=", 2)
					if len(parts) >= 1 {
						os.Unsetenv(parts[0])
					}
				}
			}
		}
		for env, val := range origValues {
			if val != "" {
				os.Setenv(env, val)
			} else {
				os.Unsetenv(env)
			}
		}
		os.Unsetenv("WHODB_CLI_ANALYTICS_DISABLED")
		os.Unsetenv("BAML_LIBRARY_PATH")
		os.RemoveAll(tempHome)
	}
}
func SetupMultipleConnections(t *testing.T, connections map[string]PostgresConfig) func() {
	t.Helper()
	origPostgres := os.Getenv("WHODB_POSTGRES")
	origHome := os.Getenv("HOME")
	tempHome, err := os.MkdirTemp("", "whodb-cli-e2e-multi-")
	if err != nil {
		t.Fatalf("Failed to create temp home: %v", err)
	}
	var connArray []string
	for alias, cfg := range connections {
		connJSON := fmt.Sprintf(`{"alias":"%s","host":"%s","user":"%s","password":"%s","database":"%s","port":"%s"}`,
			alias, cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
		connArray = append(connArray, connJSON)
	}
	fullJSON := "[" + strings.Join(connArray, ",") + "]"
	os.Setenv("WHODB_POSTGRES", fullJSON)
	os.Setenv("HOME", tempHome)
	os.Setenv("USERPROFILE", tempHome)
	os.Setenv("WHODB_CLI_ANALYTICS_DISABLED", "true")
	return func() {
		if origPostgres != "" {
			os.Setenv("WHODB_POSTGRES", origPostgres)
		} else {
			os.Unsetenv("WHODB_POSTGRES")
		}
		if origHome != "" {
			os.Setenv("HOME", origHome)
			os.Setenv("USERPROFILE", origHome)
		}
		os.Unsetenv("WHODB_CLI_ANALYTICS_DISABLED")
		os.RemoveAll(tempHome)
	}
}
func GetTempHome(t *testing.T) string {
	t.Helper()
	return os.Getenv("HOME")
}