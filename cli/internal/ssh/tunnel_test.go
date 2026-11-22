package ssh
import (
	"os"
	"path/filepath"
	"testing"
)
func TestNewTunnel_NoAuthMethods(t *testing.T) {
	t.Setenv("SSH_AUTH_SOCK", "")
	_, err := NewTunnel("localhost", 22, "user", "", "", "dbhost", 5432)
	if err == nil {
		t.Fatal("expected error when no auth methods are provided")
	}
}
func TestNewTunnel_InvalidKeyFile(t *testing.T) {
	_, err := NewTunnel("localhost", 22, "user", "/nonexistent/key/file", "", "dbhost", 5432)
	if err == nil {
		t.Fatal("expected error for nonexistent key file")
	}
}
func TestNewTunnel_BadKeyFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "bad_key")
	if err := os.WriteFile(keyPath, []byte("not a valid ssh key"), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	_, err := NewTunnel("localhost", 22, "user", keyPath, "", "dbhost", 5432)
	if err == nil {
		t.Fatal("expected error for invalid key file content")
	}
}
func TestBuildAuthMethods_PasswordOnly(t *testing.T) {
	methods, err := buildAuthMethods("", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) == 0 {
		t.Fatal("expected at least one auth method for password")
	}
}
func TestBuildAuthMethods_NoCredentials(t *testing.T) {
	t.Setenv("SSH_AUTH_SOCK", "")
	methods, err := buildAuthMethods("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(methods) != 0 {
		t.Fatalf("expected zero auth methods, got %d", len(methods))
	}
}
func TestTunnel_LocalPort_BeforeStart(t *testing.T) {
	tunnel := &Tunnel{}
	if port := tunnel.LocalPort(); port != 0 {
		t.Fatalf("expected port 0 before start, got %d", port)
	}
}
func TestTunnel_StopWithoutStart(t *testing.T) {
	tunnel := &Tunnel{
		done: make(chan struct{}),
	}
	tunnel.Stop()
	tunnel.Stop()
}