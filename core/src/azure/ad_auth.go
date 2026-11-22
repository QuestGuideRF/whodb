package azure
import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/clidey/whodb/core/src/log"
)
const (
	ScopePostgreSQLMySQL = "https://ossrdbms-aad.database.windows.net/.default"
	ScopeRedis = "https://redis.azure.com/.default"
)
func GenerateADToken(ctx context.Context, cred azcore.TokenCredential, scope string) (string, error) {
	log.Infof("Azure AD Auth: generating token for scope %s", scope)
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{scope},
	})
	if err != nil {
		log.Errorf("Azure AD Auth: token generation failed for scope %s: %v", scope, err)
		return "", HandleAzureError(err)
	}
	log.Infof("Azure AD Auth: token generated successfully (token length=%d)", len(token.Token))
	return token.Token, nil
}