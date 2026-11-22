package azure
import (
	"context"
	"time"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/postgresql/armpostgresqlflexibleservers"
	azureinfra "github.com/clidey/whodb/core/src/azure"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/providers"
)
func (p *Provider) discoverPostgreSQL(ctx context.Context) ([]providers.DiscoveredConnection, error) {
	start := time.Now()
	var connections []providers.DiscoveredConnection
	log.Debugf("Azure PostgreSQL: starting discovery for provider %s", p.config.ID)
	var pager interface {
		More() bool
		NextPage(ctx context.Context) (armpostgresqlflexibleservers.ServersClientListResponse, error)
	}
	if p.config.ResourceGroup != "" {
		pager = toPostgresListPager(p.postgresClient.NewListByResourceGroupPager(p.config.ResourceGroup, nil))
	} else {
		pager = p.postgresClient.NewListPager(nil)
	}
	for pager.More() {
		if ctx.Err() != nil {
			log.Warnf("Azure PostgreSQL: context cancelled, returning %d results so far", len(connections))
			return connections, ctx.Err()
		}
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Errorf("Azure PostgreSQL: list failed: %v", err)
			return connections, azureinfra.HandleAzureError(err)
		}
		for _, server := range page.Value {
			if server == nil || server.Name == nil {
				continue
			}
			conn := p.postgresServerToConnection(server)
			connections = append(connections, conn)
		}
	}
	log.Infof("Azure PostgreSQL: discovered %d servers in %v", len(connections), time.Since(start))
	return connections, nil
}
func (p *Provider) postgresServerToConnection(server *armpostgresqlflexibleservers.Server) providers.DiscoveredConnection {
	name := derefStr(server.Name)
	location := derefStr(server.Location)
	resourceGroup := extractResourceGroup(derefStr(server.ID))
	metadata := map[string]string{
		"port":          "5432",
		"location":      location,
		"resourceGroup": resourceGroup,
	}
	var status providers.ConnectionStatus
	if server.Properties != nil {
		if server.Properties.FullyQualifiedDomainName != nil {
			metadata["endpoint"] = *server.Properties.FullyQualifiedDomainName
		}
		if server.Properties.Version != nil {
			metadata["version"] = string(*server.Properties.Version)
		}
		status = mapPostgresStatus(server.Properties.State)
	} else {
		status = providers.ConnectionStatusUnknown
	}
	if server.SKU != nil && server.SKU.Name != nil {
		metadata["sku"] = *server.SKU.Name
	}
	return providers.DiscoveredConnection{
		ID:           p.connectionID("pg-" + name),
		ProviderType: providers.ProviderTypeAzure,
		ProviderID:   p.config.ID,
		Name:         name,
		DatabaseType: engine.DatabaseType_Postgres,
		Region:       location,
		Status:       status,
		Metadata:     metadata,
	}
}
func mapPostgresStatus(state *armpostgresqlflexibleservers.ServerState) providers.ConnectionStatus {
	if state == nil {
		return providers.ConnectionStatusUnknown
	}
	switch *state {
	case armpostgresqlflexibleservers.ServerStateReady:
		return providers.ConnectionStatusAvailable
	case armpostgresqlflexibleservers.ServerStateStarting,
		armpostgresqlflexibleservers.ServerStateUpdating:
		return providers.ConnectionStatusStarting
	case armpostgresqlflexibleservers.ServerStateStopped,
		armpostgresqlflexibleservers.ServerStateStopping,
		armpostgresqlflexibleservers.ServerStateDisabled:
		return providers.ConnectionStatusStopped
	case armpostgresqlflexibleservers.ServerStateDropping:
		return providers.ConnectionStatusDeleting
	default:
		return providers.ConnectionStatusUnknown
	}
}
type postgresRGPagerWrapper struct {
	pager interface {
		More() bool
		NextPage(ctx context.Context) (armpostgresqlflexibleservers.ServersClientListByResourceGroupResponse, error)
	}
}
func (w *postgresRGPagerWrapper) More() bool {
	return w.pager.More()
}
func (w *postgresRGPagerWrapper) NextPage(ctx context.Context) (armpostgresqlflexibleservers.ServersClientListResponse, error) {
	resp, err := w.pager.NextPage(ctx)
	if err != nil {
		return armpostgresqlflexibleservers.ServersClientListResponse{}, err
	}
	return armpostgresqlflexibleservers.ServersClientListResponse{
		ServerListResult: resp.ServerListResult,
	}, nil
}
func toPostgresListPager(pager interface {
	More() bool
	NextPage(ctx context.Context) (armpostgresqlflexibleservers.ServersClientListByResourceGroupResponse, error)
}) interface {
	More() bool
	NextPage(ctx context.Context) (armpostgresqlflexibleservers.ServersClientListResponse, error)
} {
	return &postgresRGPagerWrapper{pager: pager}
}
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}