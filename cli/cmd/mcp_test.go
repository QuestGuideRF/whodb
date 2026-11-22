package cmd
import "testing"
func TestMcpCmd_Exists(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "mcp" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'mcp' command to be registered")
	}
}
func TestMcpCmd_HasServeSubcommand(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()
	found := false
	for _, cmd := range mcpCmd.Commands() {
		if cmd.Name() == "serve" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'mcp serve' subcommand")
	}
}
func TestMcpServeCmd_HasDescription(t *testing.T) {
	if mcpServeCmd.Short == "" {
		t.Error("Expected 'mcp serve' to have a short description")
	}
	if mcpServeCmd.Long == "" {
		t.Error("Expected 'mcp serve' to have a long description")
	}
}
func TestMcpServeCmd_HasExample(t *testing.T) {
	if mcpServeCmd.Example == "" {
		t.Error("Expected 'mcp serve' to have usage examples")
	}
}
func TestMcpServeCmd_SilencesErrors(t *testing.T) {
	if !mcpServeCmd.SilenceUsage {
		t.Error("Expected SilenceUsage to be true")
	}
	if !mcpServeCmd.SilenceErrors {
		t.Error("Expected SilenceErrors to be true")
	}
}