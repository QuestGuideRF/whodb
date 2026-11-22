package mockdata
import (
	"testing"
	"github.com/clidey/whodb/core/src/env"
)
func TestIsMockDataGenerationAllowed(t *testing.T) {
	original := env.DisableMockDataGeneration
	t.Cleanup(func() {
		env.DisableMockDataGeneration = original
	})
	env.DisableMockDataGeneration = ""
	if !IsMockDataGenerationAllowed("users") {
		t.Fatalf("mock data generation should be allowed when unset")
	}
	env.DisableMockDataGeneration = "*"
	if IsMockDataGenerationAllowed("anything") {
		t.Fatalf("mock data generation should be disabled when wildcard is used")
	}
	env.DisableMockDataGeneration = "logs, metrics"
	if IsMockDataGenerationAllowed("logs") {
		t.Fatalf("logs should be disabled when listed")
	}
	if !IsMockDataGenerationAllowed("orders") {
		t.Fatalf("orders should remain enabled when not listed")
	}
}