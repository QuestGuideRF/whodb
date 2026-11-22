package engine
import (
	"strings"
	"github.com/clidey/whodb/core/graph/model"
	"github.com/clidey/whodb/core/src/types"
)
type DatabaseType string
const (
	DatabaseType_Postgres      = "Postgres"
	DatabaseType_MySQL         = "MySQL"
	DatabaseType_MariaDB       = "MariaDB"
	DatabaseType_Sqlite3       = "Sqlite3"
	DatabaseType_MongoDB       = "MongoDB"
	DatabaseType_Redis         = "Redis"
	DatabaseType_ElasticSearch = "ElasticSearch"
	DatabaseType_CockroachDB   = "CockroachDB"
	DatabaseType_ClickHouse    = "ClickHouse"
	DatabaseType_DuckDB        = "DuckDB"
	DatabaseType_Memcached     = "Memcached"
	DatabaseType_TiDB          = "TiDB"
	DatabaseType_ElastiCache   = "ElastiCache"
	DatabaseType_DocumentDB    = "DocumentDB"
	DatabaseType_Valkey     = "Valkey"
	DatabaseType_Dragonfly  = "Dragonfly"
	DatabaseType_OpenSearch = "OpenSearch"
	DatabaseType_StarRocks  = "StarRocks"
	DatabaseType_YugabyteDB = "YugabyteDB"
	DatabaseType_QuestDB    = "QuestDB"
	DatabaseType_FerretDB   = "FerretDB"
)
type LoginProfileRetriever func() ([]types.DatabaseCredentials, error)
type Engine struct {
	Plugins           []*Plugin
	LoginProfiles     []types.DatabaseCredentials
	ProfileRetrievers []LoginProfileRetriever
}
func (e *Engine) RegistryPlugin(plugin *Plugin) {
	if e.Plugins == nil {
		e.Plugins = []*Plugin{}
	}
	e.Plugins = append(e.Plugins, plugin)
}
func (e *Engine) Choose(databaseType DatabaseType) *Plugin {
	resolvedType := resolvePluginType(databaseType)
	for _, plugin := range e.Plugins {
		if strings.EqualFold(string(plugin.Type), string(resolvedType)) {
			return plugin
		}
	}
	return nil
}
func resolvePluginType(dbType DatabaseType) DatabaseType {
	if resolved, ok := pluginTypeAliases[dbType]; ok {
		return resolved
	}
	return dbType
}
func (e *Engine) AddLoginProfile(profile types.DatabaseCredentials) {
	if e.LoginProfiles == nil {
		e.LoginProfiles = []types.DatabaseCredentials{}
	}
	e.LoginProfiles = append(e.LoginProfiles, profile)
}
func (e *Engine) RegisterProfileRetriever(retriever LoginProfileRetriever) {
	if e.ProfileRetrievers == nil {
		e.ProfileRetrievers = []LoginProfileRetriever{}
	}
	e.ProfileRetrievers = append(e.ProfileRetrievers, retriever)
}
func GetStorageUnitModel(unit StorageUnit) *model.SourceObject {
	var attributes []*model.Record
	for _, attribute := range unit.Attributes {
		attributes = append(attributes, &model.Record{
			Key:   attribute.Key,
			Value: attribute.Value,
		})
	}
	kind := model.SourceObjectKindTable
	for _, attribute := range unit.Attributes {
		if !strings.EqualFold(attribute.Key, "Type") {
			continue
		}
		switch {
		case strings.EqualFold(attribute.Value, "view"):
			kind = model.SourceObjectKindView
		case strings.EqualFold(attribute.Value, "collection"):
			kind = model.SourceObjectKindCollection
		}
		break
	}
	return &model.SourceObject{
		Ref: &model.SourceObjectRef{
			Kind: kind,
			Path: []string{unit.Name},
		},
		Kind:        kind,
		Name:        unit.Name,
		Path:        []string{unit.Name},
		HasChildren: false,
		Actions:     []model.SourceAction{},
		Metadata:    attributes,
	}
}