package llm
import (
	"testing"
	"github.com/clidey/whodb/core/src/source"
)
func TestInstanceReturnsCorrectClient(t *testing.T) {
	client := ClientForModel(&source.ExternalModel{Type: string(OpenAI_LLMType), Token: "key1"})
	if client.Type != OpenAI_LLMType {
		t.Fatalf("expected type %s, got %s", OpenAI_LLMType, client.Type)
	}
	if client.APIKey != "key1" {
		t.Fatalf("expected API key 'key1', got %s", client.APIKey)
	}
	client2 := ClientForModel(&source.ExternalModel{Type: string(Ollama_LLMType), Token: "key2"})
	if client2.Type != Ollama_LLMType {
		t.Fatalf("expected type %s, got %s", Ollama_LLMType, client2.Type)
	}
	if client2.APIKey != "key2" {
		t.Fatalf("expected API key 'key2', got %s", client2.APIKey)
	}
}
func TestGetSupportedModelsReturnsErrorForUnsupportedType(t *testing.T) {
	client := LLMClient{Type: LLMType("Unknown")}
	if _, err := client.GetSupportedModels(); err == nil {
		t.Fatalf("expected error for unsupported type")
	}
}