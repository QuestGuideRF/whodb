package version
import (
	"runtime"
	"strings"
	"testing"
)
func TestGet(t *testing.T) {
	info := Get()
	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %s, want %s", info.GoVersion, runtime.Version())
	}
	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("Platform = %s, want %s", info.Platform, expectedPlatform)
	}
}
func TestInfo_String(t *testing.T) {
	info := Info{
		Version:   "1.0.0",
		Commit:    "abc123",
		BuildDate: "2026-01-01",
		GoVersion: "go1.21.0",
		Platform:  "linux/amd64",
	}
	str := info.String()
	if !strings.Contains(str, "1.0.0") {
		t.Error("String() should contain version")
	}
	if !strings.Contains(str, "abc123") {
		t.Error("String() should contain commit")
	}
	if !strings.Contains(str, "2026-01-01") {
		t.Error("String() should contain build date")
	}
	if !strings.Contains(str, "go1.21.0") {
		t.Error("String() should contain Go version")
	}
	if !strings.Contains(str, "linux/amd64") {
		t.Error("String() should contain platform")
	}
}
func TestShort(t *testing.T) {
	origVersion := Version
	origCommit := Commit
	defer func() {
		Version = origVersion
		Commit = origCommit
	}()
	Version = "1.2.3"
	Commit = "def456"
	short := Short()
	if !strings.Contains(short, "1.2.3") {
		t.Errorf("Short() = %s, should contain version 1.2.3", short)
	}
	if !strings.Contains(short, "def456") {
		t.Errorf("Short() = %s, should contain commit def456", short)
	}
	if !strings.Contains(short, "whodb-cli") {
		t.Errorf("Short() = %s, should contain 'whodb-cli'", short)
	}
}
func TestDefaultValues(t *testing.T) {
	if Version == "" {
		t.Error("Version should have a default value")
	}
	if Commit == "" {
		t.Error("Commit should have a default value")
	}
	if BuildDate == "" {
		t.Error("BuildDate should have a default value")
	}
}