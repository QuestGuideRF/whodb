package providers
import (
	"testing"
)
func TestGetBAMLConfig_RoutesToBuiltInProviders(t *testing.T) {
	RegisterProvider(NewOpenAIProvider())
	RegisterProvider(NewAnthropicProvider())
	RegisterProvider(NewOllamaProvider())
	tests := []struct {
		providerType     string
		expectedBAMLType string
	}{
		{"OpenAI", "openai"},
		{"Anthropic", "anthropic"},
		{"Ollama", "openai-generic"},
	}
	for _, tc := range tests {
		t.Run(tc.providerType, func(t *testing.T) {
			bamlType, opts, err := GetBAMLConfig(tc.providerType, "test-key", "https://example.com", "test-model")
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.providerType, err)
			}
			if bamlType != tc.expectedBAMLType {
				t.Fatalf("expected BAML type %q for %s, got %q", tc.expectedBAMLType, tc.providerType, bamlType)
			}
			if opts["model"] != "test-model" {
				t.Fatalf("expected model 'test-model' in opts, got %v", opts["model"])
			}
		})
	}
}
func TestGetBAMLConfig_RoutesToGenericProvider(t *testing.T) {
	provider := NewGenericProvider("lmstudio", "LM Studio", []string{"model-1"}, "openai-generic")
	RegisterProvider(provider)
	bamlType, opts, err := GetBAMLConfig("lmstudio", "key", "http://localhost:1234/v1", "model-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bamlType != "openai-generic" {
		t.Fatalf("expected BAML type 'openai-generic', got %q", bamlType)
	}
	if opts["base_url"] != "http://localhost:1234/v1" {
		t.Fatalf("expected base_url in opts, got %v", opts["base_url"])
	}
	if opts["default_role"] != "user" {
		t.Fatalf("expected default_role 'user' for generic provider, got %v", opts["default_role"])
	}
	if opts["request_timeout_ms"] != 60000 {
		t.Fatalf("expected request_timeout_ms 60000 for generic provider, got %v", opts["request_timeout_ms"])
	}
}
func TestGetBAMLConfig_ErrorsForUnregisteredProvider(t *testing.T) {
	_, _, err := GetBAMLConfig("nonexistent-provider", "", "", "model")
	if err == nil {
		t.Fatalf("expected error for unregistered provider type")
	}
}
func TestGenericProvider_CreateBAMLClient_IncludesDefaults(t *testing.T) {
	p := NewGenericProvider("custom", "Custom Provider", nil, "openai-generic")
	config := &ProviderConfig{
		Endpoint: "http://custom.api/v1",
		APIKey:   "custom-key",
	}
	clientType, opts, err := p.CreateBAMLClient(config, "custom-model")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clientType != "openai-generic" {
		t.Fatalf("expected client type 'openai-generic', got %q", clientType)
	}
	if opts["default_role"] != "user" {
		t.Fatalf("expected default_role 'user', got %v", opts["default_role"])
	}
	if opts["request_timeout_ms"] != 60000 {
		t.Fatalf("expected request_timeout_ms 60000, got %v", opts["request_timeout_ms"])
	}
	if opts["base_url"] != "http://custom.api/v1" {
		t.Fatalf("expected base_url, got %v", opts["base_url"])
	}
	if opts["api_key"] != "custom-key" {
		t.Fatalf("expected api_key, got %v", opts["api_key"])
	}
}