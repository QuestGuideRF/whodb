package dbcatalog
import (
	"strconv"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/source"
)
type FieldVisibility struct {
	Hostname   bool
	Username   bool
	Password   bool
	Database   bool
	SearchPath bool
}
type FieldRequirements struct {
	Hostname bool
	Username bool
	Password bool
	Database bool
}
type ConnectableDatabase struct {
	ID             engine.DatabaseType
	Label          string
	PluginType     engine.DatabaseType
	Extra          map[string]string
	Fields         FieldVisibility
	RequiredFields FieldRequirements
	IsAWSManaged   bool
	SSLModes       []source.SSLModeInfo
}
var catalog = []ConnectableDatabase{
	{
		ID:         engine.DatabaseType_Postgres,
		Label:      "Postgres",
		PluginType: engine.DatabaseType_Postgres,
		Extra:      map[string]string{"Port": "5432"},
		Fields: FieldVisibility{
			Hostname:   true,
			Username:   true,
			Password:   true,
			Database:   true,
			SearchPath: true,
		},
		RequiredFields: FieldRequirements{Hostname: true, Username: true, Password: true, Database: true},
		SSLModes:       sslModesFor(engine.DatabaseType_Postgres),
	},
	{
		ID:         engine.DatabaseType_MySQL,
		Label:      "MySQL",
		PluginType: engine.DatabaseType_MySQL,
		Extra: map[string]string{
			"Port":                       "3306",
			"Parse Time":                 "True",
			"Loc":                        "UTC",
			"Allow clear text passwords": "0",
		},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
			Database: true,
		},
		RequiredFields: FieldRequirements{Hostname: true, Username: true, Password: true, Database: true},
		SSLModes:       sslModesFor(engine.DatabaseType_MySQL),
	},
	{
		ID:         engine.DatabaseType_MariaDB,
		Label:      "MariaDB",
		PluginType: engine.DatabaseType_MariaDB,
		Extra: map[string]string{
			"Port":                       "3306",
			"Parse Time":                 "True",
			"Loc":                        "UTC",
			"Allow clear text passwords": "0",
		},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
			Database: true,
		},
		RequiredFields: FieldRequirements{Hostname: true, Username: true, Password: true, Database: true},
		SSLModes:       sslModesFor(engine.DatabaseType_MariaDB),
	},
	{
		ID:         engine.DatabaseType_CockroachDB,
		Label:      "CockroachDB",
		PluginType: engine.DatabaseType_CockroachDB,
		Extra:      map[string]string{"Port": "26257"},
		Fields: FieldVisibility{
			Hostname:   true,
			Username:   true,
			Password:   true,
			Database:   true,
			SearchPath: true,
		},
		RequiredFields: FieldRequirements{Hostname: true, Username: true, Password: true, Database: true},
		SSLModes:       sslModesFor(engine.DatabaseType_CockroachDB),
	},
	{
		ID:         engine.DatabaseType_Sqlite3,
		Label:      "Sqlite3",
		PluginType: engine.DatabaseType_Sqlite3,
		Extra:      map[string]string{},
		Fields: FieldVisibility{
			Database: true,
		},
		RequiredFields: FieldRequirements{Database: true},
	},
	{
		ID:         engine.DatabaseType_MongoDB,
		Label:      "MongoDB",
		PluginType: engine.DatabaseType_MongoDB,
		Extra: map[string]string{
			"Port":        "27017",
			"URL Params":  "?",
			"DNS Enabled": "false",
		},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
			Database: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		SSLModes:       sslModesFor(engine.DatabaseType_MongoDB),
	},
	{
		ID:         engine.DatabaseType_Redis,
		Label:      "Redis",
		PluginType: engine.DatabaseType_Redis,
		Extra:      map[string]string{"Port": "6379"},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		SSLModes:       sslModesFor(engine.DatabaseType_Redis),
	},
	{
		ID:         engine.DatabaseType_ElasticSearch,
		Label:      "ElasticSearch",
		PluginType: engine.DatabaseType_ElasticSearch,
		Extra:      map[string]string{"Port": "9200"},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		SSLModes:       sslModesFor(engine.DatabaseType_ElasticSearch),
	},
	{
		ID:         engine.DatabaseType_ClickHouse,
		Label:      "ClickHouse",
		PluginType: engine.DatabaseType_ClickHouse,
		Extra: map[string]string{
			"Port":          "9000",
			"HTTP Protocol": "disable",
			"Readonly":      "disable",
			"Debug":         "disable",
		},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
			Database: true,
		},
		RequiredFields: FieldRequirements{Hostname: true, Username: true, Password: true, Database: true},
		SSLModes:       sslModesFor(engine.DatabaseType_ClickHouse),
	},
	{
		ID:         engine.DatabaseType_DuckDB,
		Label:      "DuckDB",
		PluginType: engine.DatabaseType_DuckDB,
		Extra:      map[string]string{},
		Fields: FieldVisibility{
			Database: true,
		},
		RequiredFields: FieldRequirements{Database: true},
	},
	{
		ID:         engine.DatabaseType_Memcached,
		Label:      "Memcached",
		PluginType: engine.DatabaseType_Memcached,
		Extra:      map[string]string{"Port": "11211"},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		SSLModes:       sslModesFor(engine.DatabaseType_Memcached),
	},
	{
		ID:         engine.DatabaseType_TiDB,
		Label:      "TiDB",
		PluginType: engine.DatabaseType_TiDB,
		Extra: map[string]string{
			"Port":                       "4000",
			"Parse Time":                 "True",
			"Loc":                        "UTC",
			"Allow clear text passwords": "0",
		},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
			Database: true,
		},
		RequiredFields: FieldRequirements{Hostname: true, Username: true, Password: true, Database: true},
		SSLModes:       sslModesFor(engine.DatabaseType_TiDB),
	},
	{
		ID:         engine.DatabaseType_Valkey,
		Label:      "Valkey",
		PluginType: engine.DatabaseType_Redis,
		Extra:      map[string]string{"Port": "6379"},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		SSLModes:       sslModesFor(engine.DatabaseType_Redis),
	},
	{
		ID:         engine.DatabaseType_Dragonfly,
		Label:      "Dragonfly",
		PluginType: engine.DatabaseType_Redis,
		Extra:      map[string]string{"Port": "6379"},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		SSLModes:       sslModesFor(engine.DatabaseType_Redis),
	},
	{
		ID:         engine.DatabaseType_OpenSearch,
		Label:      "OpenSearch",
		PluginType: engine.DatabaseType_ElasticSearch,
		Extra:      map[string]string{"Port": "9200"},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		SSLModes:       sslModesFor(engine.DatabaseType_ElasticSearch),
	},
	{
		ID:         engine.DatabaseType_YugabyteDB,
		Label:      "YugabyteDB",
		PluginType: engine.DatabaseType_Postgres,
		Extra:      map[string]string{"Port": "5433"},
		Fields: FieldVisibility{
			Hostname:   true,
			Username:   true,
			Password:   true,
			Database:   true,
			SearchPath: true,
		},
		RequiredFields: FieldRequirements{Hostname: true, Username: true, Password: true, Database: true},
		SSLModes:       sslModesFor(engine.DatabaseType_Postgres),
	},
	{
		ID:         engine.DatabaseType_QuestDB,
		Label:      "QuestDB",
		PluginType: engine.DatabaseType_QuestDB,
		Extra:      map[string]string{"Port": "8812"},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
			Database: true,
		},
		RequiredFields: FieldRequirements{Hostname: true, Username: true, Password: true, Database: true},
		SSLModes:       sslModesFor(engine.DatabaseType_Postgres),
	},
	{
		ID:         engine.DatabaseType_FerretDB,
		Label:      "FerretDB",
		PluginType: engine.DatabaseType_MongoDB,
		Extra: map[string]string{
			"Port":        "27017",
			"URL Params":  "?",
			"DNS Enabled": "false",
		},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
			Database: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		SSLModes:       sslModesFor(engine.DatabaseType_MongoDB),
	},
	{
		ID:         engine.DatabaseType_ElastiCache,
		Label:      "ElastiCache",
		PluginType: engine.DatabaseType_Redis,
		Extra: map[string]string{
			"Port": "6379",
			"TLS":  "true",
		},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		IsAWSManaged:   true,
		SSLModes:       sslModesFor(engine.DatabaseType_Redis),
	},
	{
		ID:         engine.DatabaseType_DocumentDB,
		Label:      "DocumentDB",
		PluginType: engine.DatabaseType_MongoDB,
		Extra:      map[string]string{"Port": "27017"},
		Fields: FieldVisibility{
			Hostname: true,
			Username: true,
			Password: true,
			Database: true,
		},
		RequiredFields: FieldRequirements{Hostname: true},
		IsAWSManaged:   true,
		SSLModes:       sslModesFor(engine.DatabaseType_MongoDB),
	},
}
var registeredCatalog []ConnectableDatabase
func init() {
	for _, entry := range catalog {
		registerPluginAlias(entry)
	}
}
func Register(entries ...ConnectableDatabase) {
	for _, entry := range entries {
		registeredCatalog = append(registeredCatalog, cloneEntry(entry))
		registerPluginAlias(entry)
	}
}
func All() []ConnectableDatabase {
	entries := make([]ConnectableDatabase, 0, len(catalog)+len(registeredCatalog))
	for _, entry := range catalog {
		entries = append(entries, cloneEntry(entry))
	}
	for _, entry := range registeredCatalog {
		entries = append(entries, cloneEntry(entry))
	}
	return entries
}
func IDs() []string {
	ids := make([]string, 0, len(catalog)+len(registeredCatalog))
	for _, entry := range catalog {
		ids = append(ids, string(entry.ID))
	}
	for _, entry := range registeredCatalog {
		ids = append(ids, string(entry.ID))
	}
	return ids
}
func Find(id string) (ConnectableDatabase, bool) {
	for _, entry := range catalog {
		if strings.EqualFold(string(entry.ID), id) {
			return cloneEntry(entry), true
		}
	}
	for _, entry := range registeredCatalog {
		if strings.EqualFold(string(entry.ID), id) {
			return cloneEntry(entry), true
		}
	}
	return ConnectableDatabase{}, false
}
func DefaultPort(id string) (int, bool) {
	entry, ok := Find(id)
	if !ok {
		return 0, false
	}
	return parsePort(entry.Extra["Port"])
}
func IsNetworkDatabase(id string) bool {
	entry, ok := Find(id)
	return ok && entry.Fields.Hostname
}
func cloneEntry(entry ConnectableDatabase) ConnectableDatabase {
	cloned := entry
	if entry.Extra != nil {
		cloned.Extra = make(map[string]string, len(entry.Extra))
		for key, value := range entry.Extra {
			cloned.Extra[key] = value
		}
	}
	if entry.SSLModes != nil {
		cloned.SSLModes = append([]source.SSLModeInfo(nil), entry.SSLModes...)
	}
	return cloned
}
func parsePort(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return port, true
}
func registerPluginAlias(entry ConnectableDatabase) {
	if entry.ID == entry.PluginType {
		return
	}
	engine.RegisterPluginTypeAlias(entry.ID, entry.PluginType)
}