package cmd
import (
	"github.com/clidey/whodb/cli/internal/sourcetypes"
	"github.com/clidey/whodb/core/src/source"
)
func lookupDatabaseType(input string) (source.TypeSpec, bool) {
	return sourcetypes.Find(input)
}
func normalizeDBType(dbType string) string {
	return sourcetypes.Normalize(dbType)
}
func getDefaultPort(dbType string) int {
	port, ok := sourcetypes.DefaultPort(dbType)
	if !ok {
		return 0
	}
	return port
}
func isFileBasedDatabaseType(dbType string) bool {
	spec, ok := lookupSourceTypeSpec(dbType)
	if !ok {
		return false
	}
	return spec.Traits.Connection.Transport == source.ConnectionTransportFile
}
func lookupSourceTypeSpec(input string) (source.TypeSpec, bool) {
	return sourcetypes.Find(input)
}
func isConnectionFieldRequired(dbType string, key string) bool {
	return sourcetypes.ConnectionFieldRequired(dbType, key)
}