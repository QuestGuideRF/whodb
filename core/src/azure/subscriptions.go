package azure
import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/clidey/whodb/core/src/log"
)
type Subscription struct {
	ID          string
	DisplayName string
	State       string
	TenantID    string
}
func DiscoverSubscriptions(ctx context.Context, cred azcore.TokenCredential) ([]Subscription, error) {
	client, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriptions client: %w", HandleAzureError(err))
	}
	var subscriptions []Subscription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return subscriptions, HandleAzureError(err)
		}
		for _, sub := range page.Value {
			if sub == nil || sub.SubscriptionID == nil {
				continue
			}
			s := Subscription{
				ID:          *sub.SubscriptionID,
				DisplayName: derefString(sub.DisplayName),
				State:       string(*sub.State),
				TenantID:    derefString(sub.TenantID),
			}
			subscriptions = append(subscriptions, s)
		}
	}
	log.Infof("Azure subscriptions: discovered %d subscriptions", len(subscriptions))
	return subscriptions, nil
}
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}