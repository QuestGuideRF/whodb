package settings
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"github.com/clidey/whodb/core/src/common/config"
)
func TestPersistedProviderConfig_DoesNotContainCredentials(t *testing.T) {
	cfg := persistedProviderConfig{
		ID:                  "test-id",
		Name:                "Test",
		Region:              "us-west-2",
		AuthMethod:          "default",
		ProfileName:         "default",
		DiscoverRDS:         true,
		DiscoverElastiCache: true,
		DiscoverDocumentDB:  true,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	str := string(data)
	if stringContains(str, "accessKeyID") || stringContains(str, "AccessKeyID") {
		t.Error("persistedProviderConfig should not contain accessKeyID field")
	}
	if stringContains(str, "secretAccessKey") || stringContains(str, "SecretAccessKey") {
		t.Error("persistedProviderConfig should not contain secretAccessKey field")
	}
	if stringContains(str, "sessionToken") || stringContains(str, "SessionToken") {
		t.Error("persistedProviderConfig should not contain sessionToken field")
	}
}
func TestAwsSection_JSONFormat(t *testing.T) {
	section := awsSection{
		Providers: []persistedProviderConfig{
			{
				ID:         "provider-1",
				Name:       "Provider One",
				Region:     "us-west-2",
				AuthMethod: "default",
			},
			{
				ID:          "provider-2",
				Name:        "Provider Two",
				Region:      "eu-west-1",
				AuthMethod:  "profile",
				ProfileName: "production",
			},
		},
	}
	data, err := json.MarshalIndent(section, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var parsed awsSection
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(parsed.Providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(parsed.Providers))
	}
	if parsed.Providers[0].ID != "provider-1" {
		t.Errorf("expected first provider ID 'provider-1', got %s", parsed.Providers[0].ID)
	}
	if parsed.Providers[1].ProfileName != "production" {
		t.Errorf("expected profile name 'production', got %s", parsed.Providers[1].ProfileName)
	}
}
func TestGetConfigOptions(t *testing.T) {
	opts := getConfigOptions()
	if opts.AppName != "whodb" {
		t.Errorf("expected AppName 'whodb', got %s", opts.AppName)
	}
}
func TestLoadProvidersFromFile_EmptyConfig(t *testing.T) {
	resetProviders()
	config.ResetConfigPath()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	emptyConfig := `{"aws": {"providers": []}}`
	if err := os.WriteFile(configPath, []byte(emptyConfig), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}
	err := LoadProvidersFromFile()
	if err != nil {
		t.Logf("LoadProvidersFromFile returned expected error: %v", err)
	}
}
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}