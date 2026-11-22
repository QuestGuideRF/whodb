package aws
import (
	"errors"
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
)
type AuthMethod string
const (
	AuthMethodProfile AuthMethod = "profile"
	AuthMethodDefault AuthMethod = "default"
)
const (
	AdvancedKeyAuthMethod  = "Auth Method"
	AdvancedKeyProfileName = "Profile Name"
)
type AWSCredentialConfig struct {
	Region      string
	AuthMethod  AuthMethod
	ProfileName string
}
func ParseFromWhoDB(creds *engine.Credentials) (*AWSCredentialConfig, error) {
	if creds == nil {
		return nil, errors.New("credentials cannot be nil")
	}
	region := strings.TrimSpace(creds.Hostname)
	if region == "" {
		region = common.GetRecordValueOrDefault(creds.Advanced, "Region", "")
	}
	authMethodStr := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyAuthMethod, string(AuthMethodDefault))
	profileName := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyProfileName, "")
	config := &AWSCredentialConfig{
		Region:      region,
		AuthMethod:  AuthMethod(strings.ToLower(strings.TrimSpace(authMethodStr))),
		ProfileName: profileName,
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}
func (c *AWSCredentialConfig) Validate() error {
	if c.Region == "" {
		return ErrRegionRequired
	}
	switch c.AuthMethod {
	case AuthMethodProfile:
		if c.ProfileName == "" {
			return ErrProfileNameRequired
		}
	case AuthMethodDefault:
	default:
		return ErrInvalidAuthMethod
	}
	return nil
}
func (c *AWSCredentialConfig) IsProfileAuth() bool {
	return c.AuthMethod == AuthMethodProfile
}