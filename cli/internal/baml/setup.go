package baml
import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"github.com/clidey/whodb/cli/pkg/identity"
)
var BAMLVersion = "0.217.0"
var libraryInfo = map[string]struct {
	filename string
	url      string
}{
	"darwin/amd64": {
		filename: "libbaml_cffi.dylib",
		url:      "https://github.com/BoundaryML/baml/releases/download/" + BAMLVersion + "/libbaml_cffi-x86_64-apple-darwin.dylib",
	},
	"darwin/arm64": {
		filename: "libbaml_cffi.dylib",
		url:      "https://github.com/BoundaryML/baml/releases/download/" + BAMLVersion + "/libbaml_cffi-aarch64-apple-darwin.dylib",
	},
	"linux/amd64": {
		filename: "libbaml_cffi.so",
		url:      "https://github.com/BoundaryML/baml/releases/download/" + BAMLVersion + "/libbaml_cffi-x86_64-unknown-linux-gnu.so",
	},
	"linux/arm64": {
		filename: "libbaml_cffi.so",
		url:      "https://github.com/BoundaryML/baml/releases/download/" + BAMLVersion + "/libbaml_cffi-aarch64-unknown-linux-gnu.so",
	},
	"windows/amd64": {
		filename: "baml_cffi.dll",
		url:      "https://github.com/BoundaryML/baml/releases/download/" + BAMLVersion + "/baml_cffi-x86_64-pc-windows-msvc.dll",
	},
	"windows/arm64": {
		filename: "baml_cffi.dll",
		url:      "https://github.com/BoundaryML/baml/releases/download/" + BAMLVersion + "/baml_cffi-aarch64-pc-windows-msvc.dll",
	},
}
var unsupportedPlatforms = map[string]bool{
	"linux/arm":     true,
	"linux/riscv64": true,
}
var ensureOnce sync.Once
func Ensure() {
	ensureOnce.Do(func() {
		if err := Setup(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: BAML setup failed: %v\n", err)
			fmt.Fprintf(os.Stderr, "AI features will be unavailable.\n")
		}
	})
}
func Setup() error {
	if os.Getenv("BAML_LIBRARY_PATH") != "" {
		return nil
	}
	platform := runtime.GOOS + "/" + runtime.GOARCH
	if unsupportedPlatforms[platform] {
		return fmt.Errorf("BAML is not supported on %s", platform)
	}
	info, ok := libraryInfo[platform]
	if !ok {
		return fmt.Errorf("unsupported platform: %s", platform)
	}
	libDir, err := getLibraryDir()
	if err != nil {
		return fmt.Errorf("failed to get library directory: %w", err)
	}
	libPath := filepath.Join(libDir, info.filename)
	if _, err := os.Stat(libPath); err == nil {
		os.Setenv("BAML_LIBRARY_PATH", libPath)
		return nil
	}
	fmt.Fprintf(os.Stderr, "Downloading AI components (one-time setup)... ")
	if err := downloadLibrary(info.url, libPath); err != nil {
		fmt.Fprintf(os.Stderr, "failed\n")
		return fmt.Errorf("failed to download BAML library: %w", err)
	}
	fmt.Fprintf(os.Stderr, "done\n")
	os.Setenv("BAML_LIBRARY_PATH", libPath)
	return nil
}
func getLibraryDir() (string, error) {
	libDir, err := identity.HomePath("lib")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return "", err
	}
	return libDir, nil
}
func downloadLibrary(url, destPath string) error {
	destDir := filepath.Dir(destPath)
	tmpFile, err := os.CreateTemp(destDir, "baml-download-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()
	resp, err := http.Get(url)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		tmpFile.Close()
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if !strings.HasSuffix(destPath, ".dll") {
		if err := os.Chmod(tmpPath, 0755); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	tmpPath = ""
	return nil
}