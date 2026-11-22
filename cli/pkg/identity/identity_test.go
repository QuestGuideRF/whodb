package identity
import (
	"path/filepath"
	"strings"
	"testing"
)
func TestSetCurrentSupportsCustomIdentity(t *testing.T) {
	SetCurrent(Config{
		CommandName:    "custom-cli",
		DisplayName:    "Custom CLI",
		VersionName:    "custom-cli special",
		HomeDirName:    ".custom-cli",
		KeyringService: "Custom-CLI",
	})
	defer SetCurrent(CE())
	cfg := Current()
	if cfg.HomeDirName != ".custom-cli" {
		t.Fatalf("HomeDirName = %q, want %q", cfg.HomeDirName, ".custom-cli")
	}
	if cfg.KeyringService != "Custom-CLI" {
		t.Fatalf("KeyringService = %q, want %q", cfg.KeyringService, "Custom-CLI")
	}
}
func TestReplaceTextUsesActiveIdentity(t *testing.T) {
	SetCurrent(Config{
		CommandName: "custom-cli",
		DisplayName: "Custom CLI",
		HomeDirName: ".custom-cli",
	})
	defer SetCurrent(CE())
	text := ReplaceText("WhoDB CLI stores support files under ~/.whodb-cli")
	if !strings.Contains(text, "Custom CLI") {
		t.Fatalf("ReplaceText() = %q, missing custom display name", text)
	}
	if !strings.Contains(text, ".custom-cli") {
		t.Fatalf("ReplaceText() = %q, missing custom home dir name", text)
	}
}
func TestHomePathUsesEditionSpecificDirectory(t *testing.T) {
	SetCurrent(Config{
		HomeDirName: ".custom-cli",
	})
	defer SetCurrent(CE())
	path, err := HomePath("lib")
	if err != nil {
		t.Fatalf("HomePath() error = %v", err)
	}
	expected := filepath.Join(".custom-cli", "lib")
	if !strings.HasSuffix(path, expected) {
		t.Fatalf("HomePath() = %q, want suffix %q", path, expected)
	}
}