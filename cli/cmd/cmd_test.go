package cmd
import (
	"os"
	"path/filepath"
	"testing"
)
func TestRootCmd(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}
	if rootCmd.Use != "whodb-cli" {
		t.Errorf("Expected Use to be 'whodb-cli', got '%s'", rootCmd.Use)
	}
	if rootCmd.Short == "" {
		t.Error("Expected Short description to be non-empty")
	}
}
func TestRootCmd_HasSubcommands(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Error("Expected at least one subcommand")
	}
	hasConnect := false
	for _, cmd := range commands {
		cmdName := cmd.Use
		if cmdName == "connect" {
			hasConnect = true
		}
	}
	if !hasConnect {
		t.Error("Expected 'connect' subcommand")
	}
}
func TestRootCmd_Flags(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	debugFlag := rootCmd.PersistentFlags().Lookup("debug")
	if debugFlag == nil {
		t.Error("Expected 'debug' flag to exist")
	}
	noColorFlag := rootCmd.PersistentFlags().Lookup("no-color")
	if noColorFlag == nil {
		t.Error("Expected 'no-color' flag to exist")
	}
	noAnalyticsFlag := rootCmd.PersistentFlags().Lookup("no-analytics")
	if noAnalyticsFlag == nil {
		t.Error("Expected 'no-analytics' flag to exist")
	}
}
func TestInitConfig(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	initConfig()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to resolve home directory: %v", err)
	}
	configDir := filepath.Join(home, ".whodb-cli")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Config directory was not created: %s", configDir)
	}
}