package connectionopts
import (
	"fmt"
	"os"
	"slices"
	"strings"
	"github.com/clidey/whodb/cli/internal/sourcetypes"
	coressl "github.com/clidey/whodb/core/src/common/ssl"
	"github.com/clidey/whodb/core/src/source"
)
type SSLSettings struct {
	Mode           string
	CAFile         string
	ClientCertFile string
	ClientKeyFile  string
	ServerName     string
}
func HasSSLSupport(dbType string) bool {
	return len(sourcetypes.SSLModes(dbType)) > 0
}
func SSLSettingsFromAdvanced(advanced map[string]string) SSLSettings {
	return SSLSettings{
		Mode:       strings.TrimSpace(advanced[coressl.KeySSLMode]),
		ServerName: strings.TrimSpace(advanced[coressl.KeySSLServerName]),
	}
}
func ApplySSLSettings(dbType string, advanced map[string]string, settings SSLSettings) (map[string]string, error) {
	spec, ok := sourcetypes.Find(dbType)
	if !ok {
		return nil, fmt.Errorf("unsupported database type %q", dbType)
	}
	canonicalMode, modeSet, err := resolveSSLMode(spec.SSLModes, settings.Mode)
	if err != nil {
		return nil, err
	}
	if !modeSet && strings.TrimSpace(settings.CAFile) == "" && strings.TrimSpace(settings.ClientCertFile) == "" &&
		strings.TrimSpace(settings.ClientKeyFile) == "" && strings.TrimSpace(settings.ServerName) == "" {
		return cloneAdvanced(advanced), nil
	}
	if len(spec.SSLModes) == 0 {
		return nil, fmt.Errorf("%s does not support SSL options", spec.ID)
	}
	if !modeSet {
		return nil, fmt.Errorf("--ssl-mode is required when SSL options are set")
	}
	next := cloneAdvanced(advanced)
	if next == nil {
		next = make(map[string]string)
	}
	clearSSLAdvanced(next)
	next[coressl.KeySSLMode] = canonicalMode
	if canonicalMode == string(coressl.SSLModeDisabled) {
		return normalizeAdvanced(next), nil
	}
	if strings.TrimSpace(settings.CAFile) != "" {
		content, err := readSSLFile(settings.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read SSL CA file: %w", err)
		}
		next[coressl.KeySSLCACertContent] = content
	}
	if strings.TrimSpace(settings.ClientCertFile) != "" {
		content, err := readSSLFile(settings.ClientCertFile)
		if err != nil {
			return nil, fmt.Errorf("read SSL client certificate file: %w", err)
		}
		next[coressl.KeySSLClientCertContent] = content
	}
	if strings.TrimSpace(settings.ClientKeyFile) != "" {
		content, err := readSSLFile(settings.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("read SSL client key file: %w", err)
		}
		next[coressl.KeySSLClientKeyContent] = content
	}
	if serverName := strings.TrimSpace(settings.ServerName); serverName != "" {
		next[coressl.KeySSLServerName] = serverName
	}
	return normalizeAdvanced(next), nil
}
func resolveSSLMode(modes []source.SSLModeInfo, raw string) (string, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false, nil
	}
	for _, mode := range modes {
		if strings.EqualFold(trimmed, mode.Value) {
			return mode.Value, true, nil
		}
		for _, alias := range mode.Aliases {
			if strings.EqualFold(trimmed, alias) {
				return mode.Value, true, nil
			}
		}
	}
	return "", false, fmt.Errorf("invalid SSL mode %q (valid: %s)", raw, strings.Join(sslModeValues(modes), ", "))
}
func sslModeValues(modes []source.SSLModeInfo) []string {
	values := make([]string, 0, len(modes))
	for _, mode := range modes {
		values = append(values, string(mode.Value))
	}
	slices.Sort(values)
	return values
}
func readSSLFile(path string) (string, error) {
	data, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
func cloneAdvanced(advanced map[string]string) map[string]string {
	if len(advanced) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(advanced))
	for key, value := range advanced {
		cloned[key] = value
	}
	return cloned
}
func clearSSLAdvanced(advanced map[string]string) {
	if advanced == nil {
		return
	}
	delete(advanced, coressl.KeySSLMode)
	delete(advanced, coressl.KeySSLCACertContent)
	delete(advanced, coressl.KeySSLClientCertContent)
	delete(advanced, coressl.KeySSLClientKeyContent)
	delete(advanced, coressl.KeySSLServerName)
}
func normalizeAdvanced(advanced map[string]string) map[string]string {
	if len(advanced) == 0 {
		return nil
	}
	return advanced
}