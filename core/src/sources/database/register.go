package database
import (
	"maps"
	"slices"
	"github.com/clidey/whodb/core/src/dbcatalog"
	"github.com/clidey/whodb/core/src/source"
	"github.com/clidey/whodb/core/src/sourcecatalog"
)
func init() {
	Register()
}
func Register() {
	for _, entry := range dbcatalog.All() {
		spec, ok := sourcecatalog.BuildTypeSpec(sourcecatalog.DatabaseEntry{
			ID:             string(entry.ID),
			Label:          entry.Label,
			Connector:      string(entry.PluginType),
			Extra:          maps.Clone(entry.Extra),
			Fields:         sourcecatalog.FieldVisibility(entry.Fields),
			RequiredFields: sourcecatalog.FieldRequirements(entry.RequiredFields),
			IsAWSManaged:   entry.IsAWSManaged,
			SSLModes:       slices.Clone(entry.SSLModes),
		})
		if !ok {
			continue
		}
		source.RegisterType(spec)
	}
}