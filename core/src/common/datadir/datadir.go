package datadir
import (
	"os"
	"path/filepath"
	"runtime"
)
type Options struct {
	AppName string
	EnterpriseEdition bool
	Development bool
}
func Get(opts Options) (string, error) {
	if opts.AppName == "" {
		opts.AppName = "whodb"
	}
	base, err := getBaseDir()
	if err != nil {
		return "", err
	}
	appName := opts.AppName
	if opts.EnterpriseEdition {
		appName += "-ee"
	}
	if opts.Development {
		appName += "-dev"
	}
	dataDir := filepath.Join(base, appName)
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return "", err
	}
	return dataDir, nil
}
func getBaseDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "AppData", "Roaming"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return xdgData, nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".local", "share"), nil
	}
}