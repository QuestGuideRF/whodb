package bamlinit
import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)
func init() {
	configureBamlLibraryPath()
	if os.Getenv("BAML_LOG") != "" {
		return
	}
	level := strings.ToLower(os.Getenv("WHODB_LOG_LEVEL"))
	var bamlLevel string
	switch level {
	case "debug":
		bamlLevel = "debug"
	case "none", "off", "disabled":
		bamlLevel = "off"
	default:
		bamlLevel = "error"
	}
	os.Setenv("BAML_LOG", bamlLevel)
}
func configureBamlLibraryPath() {
	if existing := os.Getenv("BAML_LIBRARY_PATH"); existing != "" {
		return
	}
	if runtime.GOOS != "darwin" {
		return
	}
	execPath, err := os.Executable()
	if err != nil {
		return
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return
	}
	macosDir := filepath.Dir(execPath)
	contentsDir := filepath.Dir(macosDir)
	if filepath.Base(macosDir) != "MacOS" || filepath.Base(contentsDir) != "Contents" {
		return
	}
	var dylibName string
	switch runtime.GOARCH {
	case "arm64":
		dylibName = "libbaml_cffi-aarch64-apple-darwin.dylib"
	case "amd64":
		dylibName = "libbaml_cffi-x86_64-apple-darwin.dylib"
	default:
		return
	}
	frameworksDir := filepath.Join(contentsDir, "Frameworks")
	dylibPath := filepath.Join(frameworksDir, dylibName)
	if _, err := os.Stat(dylibPath); err == nil {
		os.Setenv("BAML_LIBRARY_PATH", dylibPath)
	}
}