package azure
import (
	"errors"
	"strings"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
)
type AuthMethod string
const (
	AuthMethodDefault AuthMethod = "default"
	AuthMethodServicePrincipal AuthMethod = "service-principal"
)
const (
	AdvancedKeyAuthMethod     = "Auth Method"
	AdvancedKeyTenantID       = "Tenant ID"
	AdvancedKeyClientID       = "Client ID"
	AdvancedKeyClientSecret   = "Client Secret"
	AdvancedKeyResourceGroup  = "Resource Group"
	AdvancedKeySubscriptionID = "Subscription ID"
)
type AzureCredentialConfig struct {
	SubscriptionID string
	TenantID       string
	ClientID       string
	ClientSecret   string
	AuthMethod     AuthMethod
	ResourceGroup  string
}
func ParseFromWhoDB(creds *engine.Credentials) (*AzureCredentialConfig, error) {
	if creds == nil {
		return nil, errors.New("credentials cannot be nil")
	}
	subscriptionID := strings.TrimSpace(creds.Hostname)
	if subscriptionID == "" {
		subscriptionID = common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeySubscriptionID, "")
	}
	authMethodStr := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyAuthMethod, string(AuthMethodDefault))
	tenantID := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyTenantID, "")
	clientID := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyClientID, "")
	clientSecret := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyClientSecret, "")
	resourceGroup := common.GetRecordValueOrDefault(creds.Advanced, AdvancedKeyResourceGroup, "")
	config := &AzureCredentialConfig{
		SubscriptionID: subscriptionID,
		TenantID:       tenantID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		AuthMethod:     AuthMethod(strings.ToLower(strings.TrimSpace(authMethodStr))),
		ResourceGroup:  resourceGroup,
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}
func (c *AzureCredentialConfig) Validate() error {
	if c.SubscriptionID == "" {
		return ErrSubscriptionRequired
	}
	switch c.AuthMethod {
	case AuthMethodServicePrincipal:
		if c.TenantID == "" {
			return ErrTenantIDRequired
		}
		if c.ClientID == "" {
			return ErrClientIDRequired
		}
		if c.ClientSecret == "" {
			return ErrClientSecretRequired
		}
	case AuthMethodDefault:
	default:
		return ErrInvalidAuthMethod
	}
	return nil
}
func (c *AzureCredentialConfig) IsServicePrincipalAuth() bool {
	return c.AuthMethod == AuthMethodServicePrincipal
}