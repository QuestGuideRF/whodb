package env
import (
	"fmt"
	"strings"
	"testing"
)
func TestGetBasePathNormalizesInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "root", input: "/", want: ""},
		{name: "prefix without leading slash", input: "whodb", want: "/whodb"},
		{name: "prefix with trailing slash", input: "/whodb/", want: "/whodb"},
		{name: "trim whitespace", input: " /nested/path/ ", want: "/nested/path"},
		{name: "allows safe punctuation", input: "/v1.2/api_gateway/", want: "/v1.2/api_gateway"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WHODB_BASE_PATH", tt.input)
			if got := getBasePath(); got != tt.want {
				t.Fatalf("expected normalized base path %q, got %q", tt.want, got)
			}
		})
	}
}
func TestGetBasePathRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "double slash", input: "/whodb//admin"},
		{name: "dot segment", input: "/whodb/./admin"},
		{name: "dot dot segment", input: "/whodb/../admin"},
		{name: "space in segment", input: "/who db"},
		{name: "query delimiter", input: "/whodb?next=/"},
		{name: "fragment delimiter", input: "/whodb#frag"},
		{name: "quote", input: `/who"db`},
		{name: "angle bracket", input: "/who<db"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("WHODB_BASE_PATH", tt.input)
			defer func() {
				panicValue := recover()
				if panicValue == nil {
					t.Fatalf("expected invalid base path %q to panic", tt.input)
				}
				if !strings.Contains(fmt.Sprint(panicValue), "invalid WHODB_BASE_PATH") {
					t.Fatalf("expected panic to mention invalid WHODB_BASE_PATH, got %v", panicValue)
				}
			}()
			_ = getBasePath()
		})
	}
}
func TestGetOllamaEndpointRespectsOverrides(t *testing.T) {
	origHost, origPort := OllamaHost, OllamaPort
	t.Cleanup(func() { OllamaHost, OllamaPort = origHost, origPort })
	OllamaHost = "ollama.example.com"
	OllamaPort = "9999"
	endpoint := GetOllamaEndpoint()
	if endpoint != "http://ollama.example.com:9999/api" {
		t.Fatalf("expected custom ollama endpoint to be honored, got %s", endpoint)
	}
}