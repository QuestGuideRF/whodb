package elasticsearch
import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/clidey/whodb/core/src/common/graphutil"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *ElasticSearchPlugin) GetGraph(config *engine.PluginConfig, database string) ([]engine.GraphUnit, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).Error("Failed to connect to ElasticSearch while getting graph")
		return nil, err
	}
	res, err := client.Indices.Stats()
	if err != nil {
		log.WithError(err).Error("Failed to get ElasticSearch indices stats for graph generation")
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		err := fmt.Errorf("error getting indices: %s", res.String())
		log.WithError(err).Error("ElasticSearch indices stats API returned error for graph generation")
		return nil, err
	}
	var stats map[string]any
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		log.WithError(err).Error("Failed to decode ElasticSearch indices stats for graph generation")
		return nil, err
	}
	indicesStats := stats["indices"].(map[string]any)
	indices := []string{}
	for indexName := range indicesStats {
		indices = append(indices, indexName)
	}
	var relations []graphutil.Relation
	uniqueRelations := make(map[string]bool)
	for indexName := range indicesStats {
		var buf bytes.Buffer
		query := map[string]any{
			"size": 100,
			"query": map[string]any{
				"match_all": map[string]any{},
			},
		}
		if err := json.NewEncoder(&buf).Encode(query); err != nil {
			log.WithError(err).WithField("indexName", indexName).Error("Failed to encode ElasticSearch query for graph generation")
			return nil, err
		}
		res, err := client.Search(
			client.Search.WithContext(context.Background()),
			client.Search.WithIndex(indexName),
			client.Search.WithBody(&buf),
		)
		if err != nil {
			log.WithError(err).WithField("indexName", indexName).Error("Failed to execute ElasticSearch search query for graph generation")
			return nil, err
		}
		if res.IsError() {
			err := fmt.Errorf("error searching documents: %s", formatElasticError(res))
			res.Body.Close()
			log.WithError(err).WithField("indexName", indexName).Error("ElasticSearch search API returned error for graph generation")
			return nil, err
		}
		var searchResult map[string]any
		if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
			res.Body.Close()
			log.WithError(err).WithField("indexName", indexName).Error("Failed to decode ElasticSearch search result for graph generation")
			return nil, err
		}
		res.Body.Close()
		hits := searchResult["hits"].(map[string]any)["hits"].([]any)
		if len(hits) > 0 {
			fieldSet := make(map[string]struct{})
			for _, h := range hits {
				doc, ok := h.(map[string]any)["_source"].(map[string]any)
				if !ok {
					continue
				}
				for fieldName := range doc {
					fieldSet[fieldName] = struct{}{}
				}
			}
			fieldNames := make([]string, 0, len(fieldSet))
			for fieldName := range fieldSet {
				fieldNames = append(fieldNames, fieldName)
			}
			foreignKeys := graphutil.InferForeignKeys(indexName, fieldNames, indices)
			for fk, fieldName := range foreignKeys {
				relKey := indexName + ":" + fk + ":ManyToOne"
				if !uniqueRelations[relKey] {
					uniqueRelations[relKey] = true
					relations = append(relations, graphutil.Relation{
						Table1:       indexName,
						Table2:       fk,
						Relation:     "ManyToOne",
						SourceColumn: fieldName,
						TargetColumn: "_id",
					})
				}
			}
		}
	}
	storageUnits, err := p.GetStorageUnits(config, database)
	if err != nil {
		log.WithError(err).Error("Failed to get storage units for ElasticSearch graph generation")
		return nil, err
	}
	return graphutil.BuildGraphUnits(relations, storageUnits), nil
}