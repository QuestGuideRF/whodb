package sourcetypes
import (
	"slices"
	"strconv"
	"strings"
	"github.com/clidey/whodb/cli/internal/bootstrap"
	"github.com/clidey/whodb/core/src/source"
	"github.com/clidey/whodb/core/src/sourcecatalog"
)
var dbTypeSynonyms = map[string]string{
	"postgresql": "Postgres",
	"sqlite":     "Sqlite3",
	"cockroach":  "CockroachDB",
	"yugabyte":   "YugabyteDB",
	"quest":      "QuestDB",
	"elastic":    "ElasticSearch",
}
func IDs() []string {
	bootstrap.Ensure()
	return slices.Clone(sourcecatalog.IDs())
}
func Synonyms() []string {
	keys := make([]string, 0, len(dbTypeSynonyms))
	for key := range dbTypeSynonyms {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
func Find(input string) (source.TypeSpec, bool) {
	bootstrap.Ensure()
	if strings.TrimSpace(input) == "" {
		return source.TypeSpec{}, false
	}
	normalizedInput := normalizeKey(input)
	if alias, ok := dbTypeSynonyms[normalizedInput]; ok {
		return sourcecatalog.Find(alias)
	}
	for _, id := range sourcecatalog.IDs() {
		if normalizeKey(id) == normalizedInput {
			return sourcecatalog.Find(id)
		}
	}
	return source.TypeSpec{}, false
}
func Normalize(input string) string {
	spec, ok := Find(input)
	if !ok {
		return strings.TrimSpace(input)
	}
	return spec.ID
}
func ConnectionField(input string, key string) (source.ConnectionField, bool) {
	spec, ok := Find(input)
	if !ok {
		return source.ConnectionField{}, false
	}
	return spec.ConnectionFieldByKey(key)
}
func ConnectionFieldRequired(input string, key string) bool {
	field, ok := ConnectionField(input, key)
	return ok && field.Required
}
func DefaultPort(input string) (int, bool) {
	field, ok := ConnectionField(input, "Port")
	if !ok {
		if IsFileTransport(input) {
			return 0, true
		}
		return 0, false
	}
	port, err := strconv.Atoi(strings.TrimSpace(field.DefaultValue))
	if err != nil {
		return 0, false
	}
	return port, true
}
func ExplainMode(input string) (source.QueryExplainMode, bool) {
	spec, ok := Find(input)
	if !ok || spec.Traits.Query.ExplainMode == "" || spec.Traits.Query.ExplainMode == source.QueryExplainModeNone {
		return source.QueryExplainModeNone, false
	}
	return spec.Traits.Query.ExplainMode, true
}
func IsFileTransport(input string) bool {
	spec, ok := Find(input)
	return ok && spec.Traits.Connection.Transport == source.ConnectionTransportFile
}
func IsNetworkTransport(input string) bool {
	spec, ok := Find(input)
	return ok && spec.Traits.Connection.Transport == source.ConnectionTransportNetwork
}
func SSLModes(input string) []source.SSLModeInfo {
	spec, ok := Find(input)
	if !ok {
		return nil
	}
	return slices.Clone(spec.SSLModes)
}
func DiscoveryAdvanced(input string, providerType string, metadata map[string]string) map[string]string {
	spec, ok := Find(input)
	if !ok {
		return nil
	}
	advanced := make(map[string]string)
	for _, item := range spec.DiscoveryPrefill.AdvancedDefaults {
		if !providerMatches(item.ProviderTypes, providerType) {
			continue
		}
		if !conditionsMatch(item.Conditions, metadata) {
			continue
		}
		value := resolveDiscoveryValue(item, metadata)
		if strings.TrimSpace(value) == "" {
			continue
		}
		advanced[item.Key] = value
	}
	if len(advanced) == 0 {
		return nil
	}
	return advanced
}
func providerMatches(allowedProviders []string, providerType string) bool {
	if len(allowedProviders) == 0 {
		return true
	}
	if strings.TrimSpace(providerType) == "" {
		return false
	}
	for _, candidate := range allowedProviders {
		if strings.EqualFold(candidate, providerType) {
			return true
		}
	}
	return false
}
func conditionsMatch(conditions []source.DiscoveryMetadataCondition, metadata map[string]string) bool {
	for _, condition := range conditions {
		if strings.TrimSpace(metadata[condition.Key]) != condition.Value {
			return false
		}
	}
	return true
}
func resolveDiscoveryValue(item source.DiscoveryAdvancedDefault, metadata map[string]string) string {
	if item.MetadataKey != "" {
		if value := strings.TrimSpace(metadata[item.MetadataKey]); value != "" {
			return value
		}
	}
	if strings.TrimSpace(item.Value) != "" {
		return item.Value
	}
	return item.DefaultValue
}
func normalizeKey(value string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(value)))
}