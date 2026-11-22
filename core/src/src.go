package src
import (
	"fmt"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/envconfig"
	"github.com/clidey/whodb/core/src/llm"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/migrate"
	"github.com/clidey/whodb/core/src/mockdata"
	"github.com/clidey/whodb/core/src/plugins/sqlite3"
	"github.com/clidey/whodb/core/src/sourcecatalog"
	_ "github.com/clidey/whodb/core/src/sources/database"
	"github.com/clidey/whodb/core/src/types"
)
var MainEngine *engine.Engine
func InitializeEngine() *engine.Engine {
	MainEngine = &engine.Engine{}
	for _, p := range engine.RegisteredPlugins() {
		MainEngine.RegistryPlugin(p)
	}
	registerDuckDB(MainEngine)
	MainEngine.AddLoginProfile(sqlite3.GetSampleProfile())
	for _, provider := range envconfig.ParseGenericProviders() {
		llm.RegisterGenericProviders(provider.Name, provider.ProviderId, provider.Models, provider.ClientType, provider.BaseURL, provider.APIKey)
	}
	for _, warning := range migrate.Warnings() {
		log.Warn(warning)
	}
	return MainEngine
}
func GetLoginProfiles() []types.DatabaseCredentials {
	profiles := []types.DatabaseCredentials{}
	profiles = append(profiles, MainEngine.LoginProfiles...)
	for _, retriever := range MainEngine.ProfileRetrievers {
		retrievedProfiles, err := retriever()
		if err != nil {
			continue
		}
		profiles = append(profiles, retrievedProfiles...)
	}
	for _, databaseType := range sourcecatalog.IDs() {
		databaseProfiles := envconfig.GetDefaultDatabaseCredentials(databaseType)
		for _, databaseProfile := range databaseProfiles {
			databaseProfile.Type = databaseType
			databaseProfile.IsProfile = true
			profiles = append(profiles, databaseProfile)
		}
	}
	return profiles
}
func GetLoginProfileId(index int, profile types.DatabaseCredentials) string {
	if len(profile.CustomId) > 0 {
		return profile.CustomId
	}
	if len(profile.Alias) > 0 {
		return profile.Alias
	}
	return fmt.Sprintf("#%v - %v@%v [%v]", index+1, profile.Username, profile.Hostname, profile.Database)
}
func GetLoginCredentials(profile types.DatabaseCredentials) *engine.Credentials {
	advanced := []engine.Record{}
	if port := defaultPortForProfile(profile.Type, profile.Port); port != "" {
		advanced = append(advanced, engine.Record{
			Key:   "Port",
			Value: port,
		})
	}
	for key, value := range profile.Advanced {
		advanced = append(advanced, engine.Record{
			Key:   key,
			Value: value,
		})
	}
	return &engine.Credentials{
		Type:      profile.Type,
		Hostname:  profile.Hostname,
		Username:  profile.Username,
		Password:  profile.Password,
		Database:  profile.Database,
		Advanced:  advanced,
		IsProfile: profile.IsProfile,
	}
}
func defaultPortForProfile(sourceType string, configuredPort string) string {
	if configuredPort != "" {
		return configuredPort
	}
	spec, ok := sourcecatalog.Find(sourceType)
	if !ok {
		return ""
	}
	field, ok := spec.ConnectionFieldByKey("Port")
	if !ok {
		return ""
	}
	return field.DefaultValue
}
func NewMockDataGenerator(fkDensityRatio int) *mockdata.Generator {
	return mockdata.NewGenerator(fkDensityRatio)
}