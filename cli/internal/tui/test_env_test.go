package tui
import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"github.com/clidey/whodb/cli/internal/config"
)
var (
	testHomeOnce sync.Once
	testHome     string
)
func setupTestEnv(t *testing.T) {
	t.Helper()
	testHomeOnce.Do(func() {
		dir, err := os.MkdirTemp("", "whodb-cli-test-home-")
		if err != nil {
			t.Fatalf("Failed to create test home: %v", err)
		}
		testHome = dir
	})
	if err := os.Setenv("HOME", testHome); err != nil {
		t.Fatalf("Failed to set HOME: %v", err)
	}
	if err := os.Setenv("USERPROFILE", testHome); err != nil {
		t.Fatalf("Failed to set USERPROFILE: %v", err)
	}
	if err := os.Setenv("XDG_DATA_HOME", testHome); err != nil {
		t.Fatalf("Failed to set XDG_DATA_HOME: %v", err)
	}
	if err := os.Setenv("APPDATA", testHome); err != nil {
		t.Fatalf("Failed to set APPDATA: %v", err)
	}
	config.ResetPathsForTesting()
	cleanupConfigFiles(t)
}
func cleanupConfigFiles(t *testing.T) {
	t.Helper()
	configPath, err := config.GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Remove config file failed: %v", err)
	}
	configDir, err := config.GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}
	historyPath := filepath.Join(configDir, "history.json")
	if err := os.Remove(historyPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Remove history file failed: %v", err)
	}
}