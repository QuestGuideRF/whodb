package sourcecatalog
import (
	"slices"
	"sort"
	"strings"
	"sync"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/source"
	"github.com/clidey/whodb/core/src/sourcecatalog/specs"
)
var (
	sessionMetadataMu    sync.RWMutex
	sessionMetadataSpecs = map[string]source.TypeSessionMetadata{}
	discoveryPrefillMu    sync.RWMutex
	discoveryPrefillSpecs = map[string]source.DiscoveryPrefill{}
)
func init() {
	registerSessionMetadata()
	RegisterDiscoveryPrefill(connectorPostgres, source.DiscoveryPrefill{
		AdvancedDefaults: []source.DiscoveryAdvancedDefault{
			{Key: "SSL Mode", Value: "require"},
		},
	})
	RegisterDiscoveryPrefill(connectorMySQL, source.DiscoveryPrefill{
		AdvancedDefaults: []source.DiscoveryAdvancedDefault{
			{Key: "SSL Mode", Value: "require"},
		},
	})
	RegisterDiscoveryPrefill(connectorMariaDB, source.DiscoveryPrefill{
		AdvancedDefaults: []source.DiscoveryAdvancedDefault{
			{Key: "SSL Mode", Value: "require"},
		},
	})
	RegisterDiscoveryPrefill("ElastiCache", source.DiscoveryPrefill{
		AdvancedDefaults: []source.DiscoveryAdvancedDefault{
			{
				Key:        "TLS",
				Value:      "true",
				Conditions: []source.DiscoveryMetadataCondition{{Key: "transitEncryption", Value: "true"}},
			},
		},
	})
	RegisterDiscoveryPrefill("Valkey", source.DiscoveryPrefill{
		AdvancedDefaults: []source.DiscoveryAdvancedDefault{
			{
				Key:        "TLS",
				Value:      "true",
				Conditions: []source.DiscoveryMetadataCondition{{Key: "transitEncryption", Value: "true"}},
			},
		},
	})
	RegisterDiscoveryPrefill("DocumentDB", source.DiscoveryPrefill{
		AdvancedDefaults: []source.DiscoveryAdvancedDefault{
			{
				Key:   "URL Params",
				Value: "?tls=true&tlsInsecure=true&replicaSet=rs0&retryWrites=false&readPreference=secondaryPreferred",
			},
		},
	})
	RegisterDiscoveryPrefill(connectorRedis, source.DiscoveryPrefill{
		AdvancedDefaults: []source.DiscoveryAdvancedDefault{
			{Key: "TLS", Value: "true", ProviderTypes: []string{"Azure"}},
		},
	})
	RegisterDiscoveryPrefill(connectorMongoDB, source.DiscoveryPrefill{
		AdvancedDefaults: []source.DiscoveryAdvancedDefault{
			{
				Key:           "URL Params",
				Value:         "?tls=true&tlsInsecure=true&retryWrites=false",
				ProviderTypes: []string{"Azure"},
			},
		},
	})
}
func registerSessionMetadata() {
	RegisterSessionMetadataAliases(
		SessionMetadataFromOperatorMap(specs.PostgresTypeDefinitions, specs.PostgreSQLSupportedOperators, specs.PostgresAliasMap),
		string(engine.DatabaseType_Postgres),
		string(engine.DatabaseType_QuestDB),
	)
	RegisterSessionMetadata(
		string(engine.DatabaseType_CockroachDB),
		SessionMetadataFromOperatorMap(specs.CockroachDBTypeDefinitions, specs.PostgreSQLSupportedOperators, specs.PostgresAliasMap),
	)
	RegisterSessionMetadataAliases(
		SessionMetadataFromOperatorMap(specs.MySQLTypeDefinitions, specs.MySQLSupportedOperators, specs.MySQLAliasMap),
		string(engine.DatabaseType_MySQL),
		string(engine.DatabaseType_MariaDB),
		string(engine.DatabaseType_TiDB),
	)
	RegisterSessionMetadata(
		string(engine.DatabaseType_ClickHouse),
		SessionMetadataFromOperatorMap(specs.ClickHouseTypeDefinitions, specs.ClickHouseSupportedOperators, specs.ClickHouseAliasMap),
	)
	RegisterSessionMetadata(
		string(engine.DatabaseType_Sqlite3),
		SessionMetadataFromOperatorMap(specs.SQLiteTypeDefinitions, specs.SQLiteSupportedOperators, specs.SQLiteAliasMap),
	)
	RegisterSessionMetadata(
		string(engine.DatabaseType_DuckDB),
		SessionMetadataFromOperatorMap(specs.DuckDBTypeDefinitions, specs.DuckDBSupportedOperators, specs.DuckDBAliasMap),
	)
	RegisterSessionMetadata(
		string(engine.DatabaseType_ElasticSearch),
		SessionMetadataFromOperatorMap(specs.ElasticSearchTypeDefinitions, specs.ElasticSearchSupportedOperators, nil),
	)
	RegisterSessionMetadata(
		string(engine.DatabaseType_MongoDB),
		SessionMetadataFromOperatorMap(specs.MongoDBTypeDefinitions, specs.MongoDBSupportedOperators, nil),
	)
	RegisterSessionMetadata(
		string(engine.DatabaseType_Redis),
		SessionMetadataFromOperatorMap(nil, specs.RedisOperators, nil),
	)
	RegisterSessionMetadata(
		string(engine.DatabaseType_Memcached),
		SessionMetadataFromOperatorMap(nil, specs.MemcachedOperators, nil),
	)
}
func RegisterSessionMetadata(id string, metadata source.TypeSessionMetadata) {
	if strings.TrimSpace(id) == "" {
		return
	}
	sessionMetadataMu.Lock()
	defer sessionMetadataMu.Unlock()
	sessionMetadataSpecs[strings.ToLower(id)] = cloneSessionMetadata(metadata)
}
func RegisterSessionMetadataAliases(metadata source.TypeSessionMetadata, ids ...string) {
	for _, id := range ids {
		RegisterSessionMetadata(id, metadata)
	}
}
func ResolveSessionMetadata(ids ...string) (*source.TypeSessionMetadata, bool) {
	sessionMetadataMu.RLock()
	defer sessionMetadataMu.RUnlock()
	for _, id := range ids {
		metadata, ok := sessionMetadataSpecs[strings.ToLower(strings.TrimSpace(id))]
		if !ok {
			continue
		}
		cloned := cloneSessionMetadata(metadata)
		return &cloned, true
	}
	return nil, false
}
func RegisterDiscoveryPrefill(id string, prefill source.DiscoveryPrefill) {
	if strings.TrimSpace(id) == "" {
		return
	}
	discoveryPrefillMu.Lock()
	defer discoveryPrefillMu.Unlock()
	discoveryPrefillSpecs[strings.ToLower(id)] = cloneDiscoveryPrefill(prefill)
}
func RegisterDiscoveryPrefillAliases(prefill source.DiscoveryPrefill, ids ...string) {
	for _, id := range ids {
		RegisterDiscoveryPrefill(id, prefill)
	}
}
func ResolveDiscoveryPrefill(ids ...string) (source.DiscoveryPrefill, bool) {
	discoveryPrefillMu.RLock()
	defer discoveryPrefillMu.RUnlock()
	for _, id := range ids {
		prefill, ok := discoveryPrefillSpecs[strings.ToLower(strings.TrimSpace(id))]
		if !ok {
			continue
		}
		return cloneDiscoveryPrefill(prefill), true
	}
	return source.DiscoveryPrefill{}, false
}
func SessionMetadataFromOperators(
	typeDefinitions []source.TypeDefinition,
	operators []string,
	aliasMap map[string]string,
) source.TypeSessionMetadata {
	clonedOperators := slices.Clone(operators)
	sort.Strings(clonedOperators)
	return source.TypeSessionMetadata{
		TypeDefinitions: slices.Clone(typeDefinitions),
		Operators:       clonedOperators,
		AliasMap:        cloneAliasMap(aliasMap),
	}
}
func SessionMetadataFromOperatorMap(
	typeDefinitions []source.TypeDefinition,
	operators map[string]string,
	aliasMap map[string]string,
) source.TypeSessionMetadata {
	keys := make([]string, 0, len(operators))
	for key := range operators {
		keys = append(keys, key)
	}
	return SessionMetadataFromOperators(typeDefinitions, keys, aliasMap)
}
func cloneSessionMetadata(metadata source.TypeSessionMetadata) source.TypeSessionMetadata {
	return source.TypeSessionMetadata{
		TypeDefinitions: slices.Clone(metadata.TypeDefinitions),
		Operators:       slices.Clone(metadata.Operators),
		AliasMap:        cloneAliasMap(metadata.AliasMap),
	}
}
func cloneAliasMap(aliasMap map[string]string) map[string]string {
	if len(aliasMap) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(aliasMap))
	for key, value := range aliasMap {
		cloned[key] = value
	}
	return cloned
}
func cloneDiscoveryPrefill(prefill source.DiscoveryPrefill) source.DiscoveryPrefill {
	cloned := source.DiscoveryPrefill{
		AdvancedDefaults: make([]source.DiscoveryAdvancedDefault, 0, len(prefill.AdvancedDefaults)),
	}
	for _, item := range prefill.AdvancedDefaults {
		cloned.AdvancedDefaults = append(cloned.AdvancedDefaults, source.DiscoveryAdvancedDefault{
			Key:           item.Key,
			Value:         item.Value,
			MetadataKey:   item.MetadataKey,
			DefaultValue:  item.DefaultValue,
			ProviderTypes: slices.Clone(item.ProviderTypes),
			Conditions:    slices.Clone(item.Conditions),
		})
	}
	return cloned
}