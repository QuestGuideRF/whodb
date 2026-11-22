package gorm_plugin
import (
	"testing"
	"github.com/clidey/whodb/core/src/engine"
)
func TestParseConnectionConfigUsesExplicitPort(t *testing.T) {
	plugin := newTestPlugin()
	input, err := plugin.ParseConnectionConfig(&engine.PluginConfig{
		Credentials: &engine.Credentials{
			Advanced: []engine.Record{{Key: "Port", Value: "5432"}},
		},
	})
	if err != nil {
		t.Fatalf("expected ParseConnectionConfig to succeed, got %v", err)
	}
	if input.Port != 5432 {
		t.Fatalf("expected explicit port 5432, got %d", input.Port)
	}
}
func TestParseConnectionConfigAllowsMissingPort(t *testing.T) {
	plugin := newTestPlugin()
	input, err := plugin.ParseConnectionConfig(&engine.PluginConfig{
		Credentials: &engine.Credentials{},
	})
	if err != nil {
		t.Fatalf("expected ParseConnectionConfig to allow missing port, got %v", err)
	}
	if input.Port != 0 {
		t.Fatalf("expected missing port to remain unset, got %d", input.Port)
	}
}