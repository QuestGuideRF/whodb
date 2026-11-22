package gcp
import (
	"errors"
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
)
type AuthMethod string
const (
	AuthMethodDefault AuthMethod = "default"
	AuthMethodServiceAccountKey AuthMethod = "service-account-key"
)
const (
	AdvancedKeyAuthMethod            = "Auth Method"
	AdvancedKeyServiceAccountKeyPath = "Service Account Key Path"
	AdvancedKeyProjectID             = "Project ID"
)
type GCPCredentialConfig struct {
	ProjectID             string
	Region                string
	AuthMethod            AuthMethod
	ServiceAccountKeyPath string
}
func ParseFromWhoDB(creds *engine.Credentials) (*GCPCredentialConfig, error) {
	if creds == nil {
		return nil, errors.New("credentials cannot be nil")
	}
	region := strings.TrimSpace(creds.Hostname)
	if region == "" {
		region = common.GetRecordValueOrDefault(creds.Advanced, "Region", "")
	}
	projectID := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyProjectID, "")
	authMethodStr := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyAuthMethod, string(AuthMethodDefault))
	serviceAccountKeyPath := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyServiceAccountKeyPath, "")
	config := &GCPCredentialConfig{
		ProjectID:             projectID,
		Region:                region,
		AuthMethod:            AuthMethod(strings.ToLower(strings.TrimSpace(authMethodStr))),
		ServiceAccountKeyPath: serviceAccountKeyPath,
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}
func (c *GCPCredentialConfig) Validate() error {
	if c.Region == "" {
		return ErrRegionRequired
	}
	if c.ProjectID == "" {
		return ErrProjectIDRequired
	}
	switch c.AuthMethod {
	case AuthMethodServiceAccountKey:
		if c.ServiceAccountKeyPath == "" {
			return ErrServiceAccountKeyPathRequired
		}
	case AuthMethodDefault:
	default:
		return ErrInvalidAuthMethod
	}
	return nil
}
func (c *GCPCredentialConfig) IsServiceAccountKeyAuth() bool {
	return c.AuthMethod == AuthMethodServiceAccountKey
}