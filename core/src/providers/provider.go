package providers
import (
	"context"
	"github.com/clidey/whodb/core/src/engine"
)
type ProviderType string
const (
	ProviderTypeManual ProviderType = "manual"
	ProviderTypeAWS ProviderType = "aws"
	ProviderTypeAzure ProviderType = "azure"
	ProviderTypeGCP ProviderType = "gcp"
)
type ConnectionProvider interface {
	Type() ProviderType
	ID() string
	Name() string
	DiscoverConnections(ctx context.Context) ([]DiscoveredConnection, error)
	TestConnection(ctx context.Context) error
	RefreshConnection(ctx context.Context, connectionID string) (bool, error)
	Close(ctx context.Context) error
}
type DiscoveredConnection struct {
	ID string
	ProviderType ProviderType
	ProviderID string
	Name string
	DatabaseType engine.DatabaseType
	Region string
	Status ConnectionStatus
	Metadata map[string]string
}
type ConnectionStatus string
const (
	ConnectionStatusAvailable ConnectionStatus = "available"
	ConnectionStatusStarting ConnectionStatus = "starting"
	ConnectionStatusStopped ConnectionStatus = "stopped"
	ConnectionStatusDeleting ConnectionStatus = "deleting"
	ConnectionStatusFailed ConnectionStatus = "failed"
	ConnectionStatusUnknown ConnectionStatus = "unknown"
)
func (s ConnectionStatus) IsAvailable() bool {
	return s == ConnectionStatusAvailable
}
type ProviderConfig struct {
	ID string
	Name string
	Enabled bool
}