package elasticsearch
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *ElasticSearchPlugin) DeleteRow(config *engine.PluginConfig, database string, storageUnit string, values map[string]string) (bool, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to connect to ElasticSearch while deleting row")
		return false, err
	}
	documentJSON, ok := values["document"]
	if !ok {
		err := errors.New("missing 'document' key in values map")
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Missing document key in delete request values")
		return false, err
	}
	var jsonValues map[string]any
	if err := json.Unmarshal([]byte(documentJSON), &jsonValues); err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to unmarshal document JSON for deletion")
		return false, err
	}
	id, ok := jsonValues["_id"]
	if !ok {
		err := errors.New("missing '_id' field in the document")
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Missing _id field in document for deletion")
		return false, err
	}
	idStr, ok := id.(string)
	if !ok || strings.TrimSpace(idStr) == "" {
		return false, fmt.Errorf("invalid '_id' field; expected non-empty string")
	}
	res, err := client.Delete(
		storageUnit,
		idStr,
		client.Delete.WithContext(context.Background()),
		client.Delete.WithRefresh("true"),
	)
	if err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).WithField("documentId", idStr).Error("Failed to execute ElasticSearch delete operation")
		return false, fmt.Errorf("failed to execute delete: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		err := fmt.Errorf("error deleting document: %s", formatElasticError(res))
		log.WithError(err).WithField("storageUnit", storageUnit).WithField("documentId", idStr).Error("ElasticSearch delete API returned error")
		return false, err
	}
	var deleteResponse map[string]any
	if err := json.NewDecoder(res.Body).Decode(&deleteResponse); err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).WithField("documentId", id).Error("Failed to decode ElasticSearch delete response")
		return false, err
	}
	if result, ok := deleteResponse["result"].(string); ok && result != "deleted" {
		err := errors.New("no documents were deleted")
		log.WithError(err).WithField("storageUnit", storageUnit).WithField("documentId", id).WithField("result", result).Error("ElasticSearch delete operation did not delete any documents")
		return false, err
	}
	return true, nil
}