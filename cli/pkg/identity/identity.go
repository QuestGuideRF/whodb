package identity
import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)
const (
	defaultCommandName        = "whodb-cli"
	defaultDisplayName        = "WhoDB CLI"
	defaultHomeDirName        = ".whodb-cli"
	defaultViperEnvPrefix     = "WHODB_CLI"
	defaultAnalyticsDisabled  = "WHODB_CLI_ANALYTICS_DISABLED"
	defaultRuntimeEnv         = "WHODB_CLI"
	defaultKeyringService     = "WhoDB-CLI"
	defaultIssueURL           = "https://github.com/clidey/whodb/issues/new?template=bug_report.md"
	defaultUpdateCheckAPIURL  = "https://api.github.com/repos/clidey/whodb/releases/latest"
	defaultUpdateCheckPageURL = "https://github.com/clidey/whodb/releases/latest"
)
type Config struct {
	Edition              string
	CommandName          string
	DisplayName          string
	VersionName          string
	HomeDirName          string
	ViperEnvPrefix       string
	AnalyticsDisabledEnv string
	RuntimeEnv           string
	KeyringService       string
	IssueURL             string
	UpdateCheckAPIURL    string
	UpdateCheckPageURL   string
	RootLongAppend       string
}
var (
	currentMu sync.RWMutex
	current   = CE()
)
func CE() Config {
	return Config{
		Edition:              "ce",
		CommandName:          defaultCommandName,
		DisplayName:          defaultDisplayName,
		VersionName:          defaultCommandName,
		HomeDirName:          defaultHomeDirName,
		ViperEnvPrefix:       defaultViperEnvPrefix,
		AnalyticsDisabledEnv: defaultAnalyticsDisabled,
		RuntimeEnv:           defaultRuntimeEnv,
		KeyringService:       defaultKeyringService,
		IssueURL:             defaultIssueURL,
		UpdateCheckAPIURL:    defaultUpdateCheckAPIURL,
		UpdateCheckPageURL:   defaultUpdateCheckPageURL,
	}
}
func Current() Config {
	currentMu.RLock()
	defer currentMu.RUnlock()
	return current
}
func SetCurrent(cfg Config) {
	currentMu.Lock()
	defer currentMu.Unlock()
	current = normalize(cfg)
}
func ReplaceText(text string) string {
	cfg := Current()
	replacer := strings.NewReplacer(
		"WhoDB CLI", cfg.DisplayName,
		"whodb-cli", cfg.CommandName,
		".whodb-cli", cfg.HomeDirName,
	)
	return replacer.Replace(text)
}
func HomePath(parts ...string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	allParts := []string{homeDir, Current().HomeDirName}
	allParts = append(allParts, parts...)
	return filepath.Join(allParts...), nil
}
func normalize(cfg Config) Config {
	base := CE()
	if strings.TrimSpace(cfg.Edition) != "" {
		base.Edition = cfg.Edition
	}
	if strings.TrimSpace(cfg.CommandName) != "" {
		base.CommandName = cfg.CommandName
	}
	if strings.TrimSpace(cfg.DisplayName) != "" {
		base.DisplayName = cfg.DisplayName
	}
	if strings.TrimSpace(cfg.VersionName) != "" {
		base.VersionName = cfg.VersionName
	}
	if strings.TrimSpace(cfg.HomeDirName) != "" {
		base.HomeDirName = cfg.HomeDirName
	}
	if strings.TrimSpace(cfg.ViperEnvPrefix) != "" {
		base.ViperEnvPrefix = cfg.ViperEnvPrefix
	}
	if strings.TrimSpace(cfg.AnalyticsDisabledEnv) != "" {
		base.AnalyticsDisabledEnv = cfg.AnalyticsDisabledEnv
	}
	if strings.TrimSpace(cfg.RuntimeEnv) != "" {
		base.RuntimeEnv = cfg.RuntimeEnv
	}
	if strings.TrimSpace(cfg.KeyringService) != "" {
		base.KeyringService = cfg.KeyringService
	}
	if strings.TrimSpace(cfg.IssueURL) != "" {
		base.IssueURL = cfg.IssueURL
	}
	if strings.TrimSpace(cfg.UpdateCheckAPIURL) != "" {
		base.UpdateCheckAPIURL = cfg.UpdateCheckAPIURL
	}
	if strings.TrimSpace(cfg.UpdateCheckPageURL) != "" {
		base.UpdateCheckPageURL = cfg.UpdateCheckPageURL
	}
	if strings.TrimSpace(cfg.RootLongAppend) != "" {
		base.RootLongAppend = cfg.RootLongAppend
	}
	return base
}