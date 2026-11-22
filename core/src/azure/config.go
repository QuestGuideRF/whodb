package azure
import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func LoadAzureCredential(creds *engine.Credentials) (azcore.TokenCredential, string, error) {
	azureCreds, err := ParseFromWhoDB(creds)
	if err != nil {
		return nil, "", err
	}
	credential, err := buildCredential(azureCreds)
	if err != nil {
		return nil, "", err
	}
	log.WithFields(map[string]any{
		"subscriptionID": azureCreds.SubscriptionID,
		"authMethod":     azureCreds.AuthMethod,
		"hasResourceGroup": azureCreds.ResourceGroup != "",
	}).Debug("Azure credential loaded successfully")
	return credential, azureCreds.SubscriptionID, nil
}
func BuildCredentialFromConfig(config *AzureCredentialConfig) (azcore.TokenCredential, error) {
	return buildCredential(config)
}
func buildCredential(config *AzureCredentialConfig) (azcore.TokenCredential, error) {
	switch config.AuthMethod {
	case AuthMethodDefault:
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			log.WithError(err).Error("Failed to create default Azure credential")
			return nil, HandleAzureError(err)
		}
		return cred, nil
	case AuthMethodServicePrincipal:
		cred, err := azidentity.NewClientSecretCredential(
			config.TenantID,
			config.ClientID,
			config.ClientSecret,
			nil,
		)
		if err != nil {
			log.WithError(err).Error("Failed to create service principal credential")
			return nil, HandleAzureError(err)
		}
		return cred, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidAuthMethod, config.AuthMethod)
	}
}