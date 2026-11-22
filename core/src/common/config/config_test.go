package config
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"github.com/clidey/whodb/core/src/common/datadir"
)
func setupTestDir(t *testing.T) (datadir.Options, func()) {
	t.Helper()
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, ".local", "share"))
	t.Setenv("APPDATA", filepath.Join(tempDir, "AppData", "Roaming"))
	ResetConfigPath()
	cleanup := func() {
		ResetConfigPath()
	}
	return datadir.Options{AppName: "whodb-test"}, cleanup
}
func TestWriteSection_CreatesConfigFile(t *testing.T) {
	opts, cleanup := setupTestDir(t)
	defer cleanup()
	type TestSection struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	section := TestSection{Name: "test", Value: 42}
	err := WriteSection("test", section, opts)
	if err != nil {
		t.Fatalf("WriteSection failed: %v", err)
	}
	path, err := GetPath(opts)
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	var raw RawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}
	if _, exists := raw["test"]; !exists {
		t.Error("test section should exist in config")
	}
}
func TestWriteSection_PreservesOtherSections(t *testing.T) {
	opts, cleanup := setupTestDir(t)
	defer cleanup()
	type Section struct {
		Data string `json:"data"`
	}
	section1 := Section{Data: "first"}
	if err := WriteSection("section1", section1, opts); err != nil {
		t.Fatalf("WriteSection (1) failed: %v", err)
	}
	section2 := Section{Data: "second"}
	if err := WriteSection("section2", section2, opts); err != nil {
		t.Fatalf("WriteSection (2) failed: %v", err)
	}
	var read1, read2 Section
	if err := ReadSection("section1", &read1, opts); err != nil {
		t.Fatalf("ReadSection (1) failed: %v", err)
	}
	if err := ReadSection("section2", &read2, opts); err != nil {
		t.Fatalf("ReadSection (2) failed: %v", err)
	}
	if read1.Data != "first" {
		t.Errorf("expected section1 data 'first', got '%s'", read1.Data)
	}
	if read2.Data != "second" {
		t.Errorf("expected section2 data 'second', got '%s'", read2.Data)
	}
}
func TestReadSection_MissingFile(t *testing.T) {
	opts, cleanup := setupTestDir(t)
	defer cleanup()
	type Section struct {
		Data string `json:"data"`
	}
	var section Section
	section.Data = "default"
	err := ReadSection("test", &section, opts)
	if err != nil {
		t.Fatalf("ReadSection should not error for missing file: %v", err)
	}
	if section.Data != "default" {
		t.Errorf("expected default value to be preserved, got '%s'", section.Data)
	}
}
func TestReadSection_MissingSection(t *testing.T) {
	opts, cleanup := setupTestDir(t)
	defer cleanup()
	type Section struct {
		Data string `json:"data"`
	}
	section1 := Section{Data: "exists"}
	if err := WriteSection("section1", section1, opts); err != nil {
		t.Fatalf("WriteSection failed: %v", err)
	}
	var section2 Section
	section2.Data = "default"
	err := ReadSection("nonexistent", &section2, opts)
	if err != nil {
		t.Fatalf("ReadSection should not error for missing section: %v", err)
	}
	if section2.Data != "default" {
		t.Errorf("expected default value to be preserved, got '%s'", section2.Data)
	}
}
func TestReadSection_CorruptedJSON(t *testing.T) {
	opts, cleanup := setupTestDir(t)
	defer cleanup()
	path, err := GetPath(opts)
	if err != nil {
		t.Fatalf("GetPath failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("failed to write corrupted config: %v", err)
	}
	type Section struct {
		Data string `json:"data"`
	}
	var section Section
	err = ReadSection("test", &section, opts)
	if err == nil {
		t.Error("ReadSection should error for corrupted JSON")
	}
}
func TestWriteSection_AtomicWrite(t *testing.T) {
	opts, cleanup := setupTestDir(t)
	defer cleanup()
	type Section struct {
		Data string `json:"data"`
	}
	section := Section{Data: "initial"}
	if err := WriteSection("test", section, opts); err != nil {
		t.Fatalf("WriteSection failed: %v", err)
	}
	section.Data = "updated"
	if err := WriteSection("test", section, opts); err != nil {
		t.Fatalf("WriteSection (update) failed: %v", err)
	}
	var read Section
	if err := ReadSection("test", &read, opts); err != nil {
		t.Fatalf("ReadSection failed: %v", err)
	}
	if read.Data != "updated" {
		t.Errorf("expected 'updated', got '%s'", read.Data)
	}
	path, _ := GetPath(opts)
	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("temp file should not exist: %s", entry.Name())
		}
	}
}
func TestRawConfig(t *testing.T) {
	raw := make(RawConfig)
	raw["test"] = json.RawMessage(`{"key":"value"}`)
	data, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var parsed RawConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	var inner struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(parsed["test"], &inner); err != nil {
		t.Fatalf("failed to unmarshal inner: %v", err)
	}
	if inner.Key != "value" {
		t.Errorf("expected key='value', got key='%s'", inner.Key)
	}
}
func TestSectionConstants(t *testing.T) {
	if SectionCLI != "cli" {
		t.Errorf("expected SectionCLI to be 'cli', got '%s'", SectionCLI)
	}
	if SectionAWS != "aws" {
		t.Errorf("expected SectionAWS to be 'aws', got '%s'", SectionAWS)
	}
	if SectionDesktop != "desktop" {
		t.Errorf("expected SectionDesktop to be 'desktop', got '%s'", SectionDesktop)
	}
}