package elasticsearch
import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
type ElasticSearchPlugin struct {
	engine.BasePlugin
}
func (p *ElasticSearchPlugin) IsAvailable(ctx context.Context, config *engine.PluginConfig) bool {
	client, err := DB(config)
	if err != nil {
		return false
	}
	res, err := client.Info()
	if err != nil || res.IsError() {
		return false
	}
	return true
}
func (p *ElasticSearchPlugin) GetDatabases(config *engine.PluginConfig) ([]string, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).Error("Failed to connect to ElasticSearch while listing indices")
		return nil, err
	}
	res, err := client.Cat.Indices(
		client.Cat.Indices.WithContext(config.OperationContext()),
		client.Cat.Indices.WithFormat("json"),
	)
	if err != nil {
		log.WithError(err).Error("Failed to get ElasticSearch indices")
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		err := fmt.Errorf("error getting indices: %s", res.String())
		log.WithError(err).Error("ElasticSearch Cat Indices API returned error")
		return nil, err
	}
	var indices []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		log.WithError(err).Error("Failed to decode ElasticSearch indices response")
		return nil, err
	}
	databases := make([]string, 0, len(indices))
	for _, idx := range indices {
		indexName, ok := idx["index"].(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(indexName, ".") {
			continue
		}
		databases = append(databases, indexName)
	}
	return databases, nil
}
func (p *ElasticSearchPlugin) GetStorageUnits(config *engine.PluginConfig, database string) ([]engine.StorageUnit, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).Error("Failed to connect to ElasticSearch while getting storage units")
		return nil, err
	}
	res, err := client.Indices.Stats(client.Indices.Stats.WithContext(config.OperationContext()))
	if err != nil {
		log.WithError(err).Error("Failed to get ElasticSearch indices stats")
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		err := fmt.Errorf("error getting stats for indices: %s", res.String())
		log.WithError(err).Error("ElasticSearch indices stats API returned error")
		return nil, err
	}
	var stats map[string]any
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		log.WithError(err).Error("Failed to decode ElasticSearch indices stats response")
		return nil, err
	}
	indicesStats, ok := stats["indices"].(map[string]any)
	if !ok {
		log.WithField("stats", stats).Error("Unexpected indices stats format from ElasticSearch")
		return nil, fmt.Errorf("unexpected indices stats format")
	}
	storageUnits := make([]engine.StorageUnit, 0, len(indicesStats))
	for indexName, indexStatsInterface := range indicesStats {
		if strings.HasPrefix(indexName, ".") {
			continue
		}
		indexStats, ok := indexStatsInterface.(map[string]any)
		if !ok {
			log.Warnf("Skipping index %s: unexpected stats format", indexName)
			continue
		}
		primaries, ok := indexStats["primaries"].(map[string]any)
		if !ok {
			log.Warnf("Skipping index %s: missing primaries data", indexName)
			continue
		}
		docs, ok := primaries["docs"].(map[string]any)
		if !ok {
			log.Warnf("Skipping index %s: missing docs data", indexName)
			continue
		}
		store, ok := primaries["store"].(map[string]any)
		if !ok {
			log.Warnf("Skipping index %s: missing store data", indexName)
			continue
		}
		storageUnit := engine.StorageUnit{
			Name: indexName,
			Attributes: []engine.Record{
				{Key: "Type", Value: "Index"},
				{Key: "Storage Size", Value: fmt.Sprintf("%v", store["size_in_bytes"])},
				{Key: "Count", Value: fmt.Sprintf("%v", docs["count"])},
			},
		}
		storageUnits = append(storageUnits, storageUnit)
	}
	return storageUnits, nil
}
func (p *ElasticSearchPlugin) StorageUnitExists(config *engine.PluginConfig, database string, index string) (bool, error) {
	client, err := DB(config)
	if err != nil {
		return false, err
	}
	res, err := client.Indices.Exists([]string{index})
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	return res.StatusCode == 200, nil
}
func (p *ElasticSearchPlugin) FormatValue(val any) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}
func init() {
	engine.RegisterPlugin(NewElasticSearchPlugin())
}
func NewElasticSearchPlugin() *engine.Plugin {
	return &engine.Plugin{
		Type:            engine.DatabaseType_ElasticSearch,
		PluginFunctions: &ElasticSearchPlugin{},
	}
}