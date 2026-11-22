package crash
import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)
func TestPrintCrashReport(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	printCrashReport("test panic error")
	w.Close()
	os.Stderr = oldStderr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	expectedStrings := []string{
		"WhoDB CLI crashed unexpectedly",
		"test panic error",
		"github.com/clidey/whodb/issues",
		"Stack Trace",
		"Describe the bug",
	}
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Crash report should contain %q", expected)
		}
	}
}
func TestPrintCrashReport_ContainsVersionInfo(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	printCrashReport("version test error")
	w.Close()
	os.Stderr = oldStderr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	if !strings.Contains(output, "OS:") {
		t.Error("Crash report should contain OS info")
	}
	if !strings.Contains(output, "Architecture:") {
		t.Error("Crash report should contain architecture info")
	}
	if !strings.Contains(output, "Go Version:") {
		t.Error("Crash report should contain Go version")
	}
}
func TestPrintCrashReport_ContainsCommandLine(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	printCrashReport("command line test")
	w.Close()
	os.Stderr = oldStderr
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()
	if !strings.Contains(output, "To Reproduce") {
		t.Error("Crash report should contain 'To Reproduce' section")
	}
}
func TestHandler_NoPanic(t *testing.T) {
	func() {
		defer Handler()
	}()
}