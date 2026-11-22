package sourcecatalog
import (
	"strings"
	"sync"
	"github.com/clidey/whodb/core/src/azure"
)
var (
	azureADScopeMu    sync.RWMutex
	azureADScopeSpecs = map[string]string{}
)
func init() {
	RegisterAzureADScopeAliases(azure.ScopePostgreSQLMySQL, connectorPostgres, connectorMySQL)
	RegisterAzureADScope(connectorRedis, azure.ScopeRedis)
}
func RegisterAzureADScope(id string, scope string) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(scope) == "" {
		return
	}
	azureADScopeMu.Lock()
	defer azureADScopeMu.Unlock()
	azureADScopeSpecs[strings.ToLower(strings.TrimSpace(id))] = strings.TrimSpace(scope)
}
func RegisterAzureADScopeAliases(scope string, ids ...string) {
	for _, id := range ids {
		RegisterAzureADScope(id, scope)
	}
}
func ResolveAzureADScope(ids ...string) (string, bool) {
	azureADScopeMu.RLock()
	defer azureADScopeMu.RUnlock()
	for _, id := range ids {
		scope, ok := azureADScopeSpecs[strings.ToLower(strings.TrimSpace(id))]
		if ok {
			return scope, true
		}
	}
	return "", false
}