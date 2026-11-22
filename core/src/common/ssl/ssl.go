package ssl
import (
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/sourcecatalog"
)
type SSLMode string
const (
	SSLModeDisabled       SSLMode = "disabled"
	SSLModePreferred      SSLMode = "preferred"
	SSLModeRequired       SSLMode = "required"
	SSLModeVerifyCA       SSLMode = "verify-ca"
	SSLModeVerifyIdentity SSLMode = "verify-identity"
	SSLModeEnabled        SSLMode = "enabled"
	SSLModeInsecure       SSLMode = "insecure"
)
const (
	KeySSLMode              = "SSL Mode"
	KeySSLCACertContent     = "SSL CA Content"
	KeySSLClientCertContent = "SSL Client Cert Content"
	KeySSLClientKeyContent  = "SSL Client Key Content"
	KeySSLServerName        = "SSL Server Name"
	KeySSLCACertPath     = "SSL CA Path"
	KeySSLClientCertPath = "SSL Client Cert Path"
	KeySSLClientKeyPath  = "SSL Client Key Path"
)
type CertificateInput struct {
	Content string
	Path    string
}
type SSLConfig struct {
	Mode       SSLMode
	CACert     CertificateInput
	ClientCert CertificateInput
	ClientKey  CertificateInput
	ServerName string
}
type SSLModeInfo struct {
	Value       SSLMode
	Label       string
	Description string
}
var (
	ModeInfoDisabled       = SSLModeInfo{Value: SSLModeDisabled, Label: "Disabled", Description: "No SSL/TLS encryption"}
	ModeInfoPreferred      = SSLModeInfo{Value: SSLModePreferred, Label: "Preferred", Description: "Use TLS if server supports it"}
	ModeInfoRequired       = SSLModeInfo{Value: SSLModeRequired, Label: "Required", Description: "Require TLS, skip certificate verification"}
	ModeInfoVerifyCA       = SSLModeInfo{Value: SSLModeVerifyCA, Label: "Verify CA", Description: "Verify server certificate against CA"}
	ModeInfoVerifyIdentity = SSLModeInfo{Value: SSLModeVerifyIdentity, Label: "Verify Identity", Description: "Verify CA and server hostname"}
	ModeInfoEnabled        = SSLModeInfo{Value: SSLModeEnabled, Label: "Enabled", Description: "Enable TLS with certificate verification"}
	ModeInfoInsecure       = SSLModeInfo{Value: SSLModeInsecure, Label: "Insecure", Description: "Enable TLS, skip certificate verification"}
)
func GetSSLModes(dbType engine.DatabaseType) []SSLModeInfo {
	spec, ok := sourcecatalog.Find(string(dbType))
	if !ok || len(spec.SSLModes) == 0 {
		return nil
	}
	modes := make([]SSLModeInfo, 0, len(spec.SSLModes))
	for _, mode := range spec.SSLModes {
		modes = append(modes, SSLModeInfo{
			Value:       SSLMode(mode.Value),
			Label:       mode.Label,
			Description: mode.Description,
		})
	}
	return modes
}
func ValidateSSLMode(dbType engine.DatabaseType, mode SSLMode) bool {
	for _, item := range GetSSLModes(dbType) {
		if item.Value == mode {
			return true
		}
	}
	return false
}
func NormalizeSSLMode(dbType engine.DatabaseType, mode string) SSLMode {
	trimmed := strings.TrimSpace(mode)
	spec, ok := sourcecatalog.Find(string(dbType))
	if !ok {
		return SSLMode(trimmed)
	}
	for _, item := range spec.SSLModes {
		if strings.EqualFold(trimmed, item.Value) {
			return SSLMode(item.Value)
		}
		for _, alias := range item.Aliases {
			if strings.EqualFold(trimmed, alias) {
				return SSLMode(item.Value)
			}
		}
	}
	return SSLMode(trimmed)
}
func GetSSLModeAliases(dbType engine.DatabaseType, mode SSLMode) []string {
	spec, ok := sourcecatalog.Find(string(dbType))
	if !ok {
		return nil
	}
	for _, item := range spec.SSLModes {
		if item.Value == string(mode) {
			return append([]string(nil), item.Aliases...)
		}
	}
	return nil
}
func HasSSLSupport(dbType engine.DatabaseType) bool {
	return len(GetSSLModes(dbType)) > 0
}
func (c *SSLConfig) IsEnabled() bool {
	return c != nil && c.Mode != SSLModeDisabled && c.Mode != ""
}
func ParseSSLConfig(dbType engine.DatabaseType, advanced []engine.Record, hostname string, isProfile bool) *SSLConfig {
	log.Debugf("[SSL] ParseSSLConfig: received %d advanced records", len(advanced))
	for _, rec := range advanced {
		log.Debugf("[SSL] ParseSSLConfig: key=%q value=%q", rec.Key, rec.Value)
	}
	modeStr := common.GetRecordValueOrDefault(advanced, KeySSLMode, string(SSLModeDisabled))
	log.Debugf("[SSL] ParseSSLConfig: modeStr=%q (looking for key=%q)", modeStr, KeySSLMode)
	mode := NormalizeSSLMode(dbType, modeStr)
	if !ValidateSSLMode(dbType, mode) {
		return nil
	}
	if mode == SSLModeDisabled {
		return nil
	}
	config := &SSLConfig{
		Mode: mode,
		CACert: CertificateInput{
			Content: common.GetRecordValueOrDefault(advanced, KeySSLCACertContent, ""),
		},
		ClientCert: CertificateInput{
			Content: common.GetRecordValueOrDefault(advanced, KeySSLClientCertContent, ""),
		},
		ClientKey: CertificateInput{
			Content: common.GetRecordValueOrDefault(advanced, KeySSLClientKeyContent, ""),
		},
		ServerName: common.GetRecordValueOrDefault(advanced, KeySSLServerName, hostname),
	}
	if isProfile {
		config.CACert.Path = common.GetRecordValueOrDefault(advanced, KeySSLCACertPath, "")
		config.ClientCert.Path = common.GetRecordValueOrDefault(advanced, KeySSLClientCertPath, "")
		config.ClientKey.Path = common.GetRecordValueOrDefault(advanced, KeySSLClientKeyPath, "")
	}
	return config
}