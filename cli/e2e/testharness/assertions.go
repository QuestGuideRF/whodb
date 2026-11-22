package testharness
import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)
func AssertContains(t *testing.T, output, substring string) {
	t.Helper()
	if !strings.Contains(output, substring) {
		t.Errorf("Expected output to contain %q, but it doesn't.\nOutput: %s", substring, output)
	}
}
func AssertNotContains(t *testing.T, output, substring string) {
	t.Helper()
	if strings.Contains(output, substring) {
		t.Errorf("Expected output to NOT contain %q, but it does.\nOutput: %s", substring, output)
	}
}
func AssertJSONContains(t *testing.T, output, key string) {
	t.Helper()
	var data any
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, output)
	}
	if arr, ok := data.([]any); ok {
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				if _, exists := m[key]; exists {
					return
				}
			}
		}
		t.Errorf("Expected JSON array to contain objects with key %q, but none found.\nOutput: %s", key, output)
		return
	}
	if m, ok := data.(map[string]any); ok {
		if _, exists := m[key]; !exists {
			t.Errorf("Expected JSON object to contain key %q, but it doesn't.\nOutput: %s", key, output)
		}
		return
	}
	t.Errorf("Expected JSON to be an array or object, got something else.\nOutput: %s", output)
}
func AssertJSONArrayLength(t *testing.T, output string, expected int) {
	t.Helper()
	var arr []any
	if err := json.Unmarshal([]byte(output), &arr); err != nil {
		t.Fatalf("Failed to parse JSON array: %v\nOutput: %s", err, output)
	}
	if len(arr) != expected {
		t.Errorf("Expected JSON array length %d, got %d.\nOutput: %s", expected, len(arr), output)
	}
}
func AssertJSONArrayMinLength(t *testing.T, output string, minLength int) {
	t.Helper()
	var arr []any
	if err := json.Unmarshal([]byte(output), &arr); err != nil {
		t.Fatalf("Failed to parse JSON array: %v\nOutput: %s", err, output)
	}
	if len(arr) < minLength {
		t.Errorf("Expected JSON array length >= %d, got %d.\nOutput: %s", minLength, len(arr), output)
	}
}
func AssertJSONArrayContainsValue(t *testing.T, output, key, value string) {
	t.Helper()
	var arr []map[string]any
	if err := json.Unmarshal([]byte(output), &arr); err != nil {
		t.Fatalf("Failed to parse JSON array: %v\nOutput: %s", err, output)
	}
	for _, item := range arr {
		if v, ok := item[key]; ok {
			if str, ok := v.(string); ok && str == value {
				return
			}
			if str, ok := v.(float64); ok {
				if strings.Contains(value, ".") {
					if v == str {
						return
					}
				}
			}
		}
	}
	t.Errorf("Expected JSON array to contain an object where %q = %q, but none found.\nOutput: %s", key, value, output)
}
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s, but it doesn't", path)
	}
}
func AssertFileNotEmpty(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s, but it doesn't", path)
		return
	}
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Errorf("Expected file at %s to be non-empty, but it's empty", path)
	}
}
func AssertFileContains(t *testing.T, path, substring string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	if !strings.Contains(string(content), substring) {
		t.Errorf("Expected file %s to contain %q, but it doesn't.\nContent: %s", path, substring, string(content))
	}
}