package azure
import (
	"context"
	"strings"
	"time"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos/v3"
	azureinfra "github.com/clidey/whodb/core/src/azure"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/providers"
)
func (p *Provider) discoverCosmosDB(ctx context.Context) ([]providers.DiscoveredConnection, error) {
	start := time.Now()
	var connections []providers.DiscoveredConnection
	log.Debugf("Azure CosmosDB: starting discovery for provider %s", p.config.ID)
	var pager interface {
		More() bool
		NextPage(ctx context.Context) (armcosmos.DatabaseAccountsClientListResponse, error)
	}
	if p.config.ResourceGroup != "" {
		pager = toCosmosListPager(p.cosmosClient.NewListByResourceGroupPager(p.config.ResourceGroup, nil))
	} else {
		pager = p.cosmosClient.NewListPager(nil)
	}
	for pager.More() {
		if ctx.Err() != nil {
			log.Warnf("Azure CosmosDB: context cancelled, returning %d results so far", len(connections))
			return connections, ctx.Err()
		}
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Errorf("Azure CosmosDB: list failed: %v", err)
			return connections, azureinfra.HandleAzureError(err)
		}
		for _, account := range page.Value {
			if account == nil || account.Name == nil {
				continue
			}
			if !isMongoDBAccount(account) {
				continue
			}
			conn := p.cosmosAccountToConnection(account)
			connections = append(connections, conn)
		}
	}
	log.Infof("Azure CosmosDB: discovered %d MongoDB accounts in %v", len(connections), time.Since(start))
	return connections, nil
}
func isMongoDBAccount(account *armcosmos.DatabaseAccountGetResults) bool {
	if account.Kind != nil && *account.Kind == armcosmos.DatabaseAccountKindMongoDB {
		return true
	}
	if account.Properties != nil {
		for _, cap := range account.Properties.Capabilities {
			if cap != nil && cap.Name != nil && strings.EqualFold(*cap.Name, "EnableMongo") {
				return true
			}
		}
	}
	return false
}
func (p *Provider) cosmosAccountToConnection(account *armcosmos.DatabaseAccountGetResults) providers.DiscoveredConnection {
	name := derefStr(account.Name)
	location := derefStr(account.Location)
	resourceGroup := extractResourceGroup(derefStr(account.ID))
	metadata := map[string]string{
		"port":          "10255",
		"location":      location,
		"resourceGroup": resourceGroup,
	}
	if account.Kind != nil {
		metadata["kind"] = string(*account.Kind)
	}
	var status providers.ConnectionStatus
	if account.Properties != nil {
		if account.Properties.DocumentEndpoint != nil {
			endpoint := *account.Properties.DocumentEndpoint
			endpoint = strings.TrimPrefix(endpoint, "https://")
			endpoint = strings.TrimPrefix(endpoint, "http://")
			endpoint = strings.TrimSuffix(endpoint, "/")
			if idx := strings.Index(endpoint, ":"); idx != -1 {
				endpoint = endpoint[:idx]
			}
			metadata["endpoint"] = endpoint
		}
		status = mapCosmosStatus(account.Properties.ProvisioningState)
	} else {
		status = providers.ConnectionStatusUnknown
	}
	return providers.DiscoveredConnection{
		ID:           p.connectionID("cosmos-" + name),
		ProviderType: providers.ProviderTypeAzure,
		ProviderID:   p.config.ID,
		Name:         name,
		DatabaseType: engine.DatabaseType_MongoDB,
		Region:       location,
		Status:       status,
		Metadata:     metadata,
	}
}
func mapCosmosStatus(state *string) providers.ConnectionStatus {
	if state == nil {
		return providers.ConnectionStatusUnknown
	}
	switch strings.ToLower(*state) {
	case "succeeded":
		return providers.ConnectionStatusAvailable
	case "creating", "updating":
		return providers.ConnectionStatusStarting
	case "deleting":
		return providers.ConnectionStatusDeleting
	case "failed":
		return providers.ConnectionStatusFailed
	default:
		return providers.ConnectionStatusUnknown
	}
}
type cosmosRGPagerWrapper struct {
	pager interface {
		More() bool
		NextPage(ctx context.Context) (armcosmos.DatabaseAccountsClientListByResourceGroupResponse, error)
	}
}
func (w *cosmosRGPagerWrapper) More() bool {
	return w.pager.More()
}
func (w *cosmosRGPagerWrapper) NextPage(ctx context.Context) (armcosmos.DatabaseAccountsClientListResponse, error) {
	resp, err := w.pager.NextPage(ctx)
	if err != nil {
		return armcosmos.DatabaseAccountsClientListResponse{}, err
	}
	return armcosmos.DatabaseAccountsClientListResponse{
		DatabaseAccountsListResult: resp.DatabaseAccountsListResult,
	}, nil
}
func toCosmosListPager(pager interface {
	More() bool
	NextPage(ctx context.Context) (armcosmos.DatabaseAccountsClientListByResourceGroupResponse, error)
}) interface {
	More() bool
	NextPage(ctx context.Context) (armcosmos.DatabaseAccountsClientListResponse, error)
} {
	return &cosmosRGPagerWrapper{pager: pager}
}