package mongodb
import (
	"context"
	"fmt"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)
type MongoDBPlugin struct {
	engine.BasePlugin
}
func (p *MongoDBPlugin) IsAvailable(ctx context.Context, config *engine.PluginConfig) bool {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to connect to MongoDB for availability check")
		return false
	}
	defer disconnectClient(client)
	return true
}
func (p *MongoDBPlugin) GetDatabases(config *engine.PluginConfig) ([]string, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to connect to MongoDB for database listing")
		return nil, err
	}
	ctx, cancel := opCtx(config)
	defer cancel()
	defer disconnectClient(client)
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to list MongoDB database names")
		return nil, err
	}
	return databases, nil
}
func (p *MongoDBPlugin) GetAllSchemas(config *engine.PluginConfig) ([]string, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to connect to MongoDB for schema listing")
		return nil, err
	}
	ctx, cancel := opCtx(config)
	defer cancel()
	defer disconnectClient(client)
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.WithError(err).WithField("hostname", config.Credentials.Hostname).Error("Failed to list MongoDB database names")
		return nil, err
	}
	return databases, nil
}
func (p *MongoDBPlugin) GetStorageUnits(config *engine.PluginConfig, database string) ([]engine.StorageUnit, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithFields(map[string]any{
			"hostname": config.Credentials.Hostname,
			"database": database,
		}).Error("Failed to connect to MongoDB for storage unit listing")
		return nil, err
	}
	ctx, cancel := opCtx(config)
	defer cancel()
	defer disconnectClient(client)
	db := client.Database(database)
	listOpts := options.ListCollections().SetAuthorizedCollections(true)
	cursor, err := db.ListCollections(ctx, bson.M{}, listOpts)
	if err != nil {
		log.WithError(err).WithFields(map[string]any{
			"hostname": config.Credentials.Hostname,
			"database": database,
		}).Error("Failed to list MongoDB collections")
		return nil, err
	}
	defer cursor.Close(ctx)
	storageUnits := []engine.StorageUnit{}
	for cursor.Next(ctx) {
		var collectionInfo bson.M
		if err := cursor.Decode(&collectionInfo); err != nil {
			log.WithError(err).WithFields(map[string]any{
				"hostname": config.Credentials.Hostname,
				"database": database,
			}).Error("Failed to decode MongoDB collection info")
			return nil, err
		}
		collectionName, _ := collectionInfo["name"].(string)
		collectionType, _ := collectionInfo["type"].(string)
		if strings.HasPrefix(collectionName, "system.") {
			continue
		}
		storageUnit := engine.StorageUnit{Name: collectionName}
		if collectionType == "view" {
			viewOn, _ := collectionInfo["options"].(map[string]any)["viewOn"].(string)
			storageUnit.Attributes = []engine.Record{
				{Key: "Type", Value: "View"},
				{Key: "View On", Value: viewOn},
			}
		} else {
			stats := bson.M{}
			err := db.RunCommand(ctx, bson.D{{Key: "collStats", Value: collectionName}}).Decode(&stats)
			if err != nil {
				log.WithError(err).WithFields(map[string]any{
					"hostname":   config.Credentials.Hostname,
					"database":   database,
					"collection": collectionName,
				}).Error("Failed to get MongoDB collection statistics")
				return nil, err
			}
			storageUnit.Attributes = []engine.Record{
				{Key: "Type", Value: "Collection"},
				{Key: "Storage Size", Value: fmt.Sprintf("%v", stats["storageSize"])},
				{Key: "Count", Value: fmt.Sprintf("%v", stats["count"])},
			}
		}
		storageUnits = append(storageUnits, storageUnit)
	}
	if err := cursor.Err(); err != nil {
		log.WithError(err).WithFields(map[string]any{
			"hostname": config.Credentials.Hostname,
			"database": database,
		}).Error("MongoDB cursor error while listing collections")
		return nil, err
	}
	return storageUnits, nil
}
func (p *MongoDBPlugin) StorageUnitExists(config *engine.PluginConfig, database string, collection string) (bool, error) {
	client, err := DB(config)
	if err != nil {
		return false, err
	}
	ctx, cancel := opCtx(config)
	defer cancel()
	defer disconnectClient(client)
	db := client.Database(database)
	names, err := db.ListCollectionNames(ctx, bson.M{"name": collection})
	if err != nil {
		return false, err
	}
	return len(names) > 0, nil
}
func (p *MongoDBPlugin) FormatValue(val any) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}
func (p *MongoDBPlugin) GetColumnConstraints(config *engine.PluginConfig, schema string, storageUnit string) (map[string]map[string]any, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithFields(map[string]any{
			"hostname":    config.Credentials.Hostname,
			"schema":      schema,
			"storageUnit": storageUnit,
		}).Error("Failed to connect to MongoDB for column constraints")
		return make(map[string]map[string]any), nil
	}
	ctx, cancel := opCtx(config)
	defer cancel()
	defer disconnectClient(client)
	db := client.Database(schema)
	filter := bson.M{"name": storageUnit}
	cursor, err := db.ListCollections(ctx, filter)
	if err != nil {
		log.WithError(err).WithFields(map[string]any{
			"schema":      schema,
			"storageUnit": storageUnit,
		}).Debug("Failed to list collections for schema validation")
		return make(map[string]map[string]any), nil
	}
	defer cursor.Close(ctx)
	var collInfo bson.M
	if !cursor.Next(ctx) {
		return make(map[string]map[string]any), nil
	}
	if err := cursor.Decode(&collInfo); err != nil {
		log.WithError(err).WithField("collection", storageUnit).Debug("Failed to decode collection info")
		return make(map[string]map[string]any), nil
	}
	opts, ok := collInfo["options"].(map[string]any)
	if !ok {
		return make(map[string]map[string]any), nil
	}
	validator, ok := opts["validator"].(map[string]any)
	if !ok {
		return make(map[string]map[string]any), nil
	}
	jsonSchema, ok := validator["$jsonSchema"].(map[string]any)
	if !ok {
		return make(map[string]map[string]any), nil
	}
	constraints := parseMongoDBJsonSchema(jsonSchema)
	log.WithFields(map[string]any{
		"collection":      storageUnit,
		"constraintCount": len(constraints),
	}).Debug("Parsed MongoDB schema validation constraints")
	return constraints, nil
}
func parseMongoDBJsonSchema(schema map[string]any) map[string]map[string]any {
	constraints := make(map[string]map[string]any)
	requiredFields := make(map[string]bool)
	if required, ok := schema["required"].([]any); ok {
		for _, field := range required {
			if fieldName, ok := field.(string); ok {
				requiredFields[fieldName] = true
			}
		}
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return constraints
	}
	for fieldName, fieldSchemaRaw := range properties {
		fieldSchema, ok := fieldSchemaRaw.(map[string]any)
		if !ok {
			continue
		}
		colConstraints := make(map[string]any)
		colConstraints["nullable"] = !requiredFields[fieldName]
		if bsonType, ok := fieldSchema["bsonType"].(string); ok {
			colConstraints["type"] = bsonType
		}
		if enum, ok := fieldSchema["enum"].([]any); ok {
			values := make([]string, 0, len(enum))
			for _, v := range enum {
				if s, ok := v.(string); ok {
					values = append(values, s)
				}
			}
			if len(values) > 0 {
				colConstraints["check_values"] = values
			}
		}
		if minVal := toFloat64(fieldSchema["minimum"]); minVal != nil {
			colConstraints["check_min"] = *minVal
		}
		if maxVal := toFloat64(fieldSchema["maximum"]); maxVal != nil {
			colConstraints["check_max"] = *maxVal
		}
		if maxLen := toFloat64(fieldSchema["maxLength"]); maxLen != nil {
			colConstraints["length"] = int(*maxLen)
		}
		if pattern, ok := fieldSchema["pattern"].(string); ok {
			colConstraints["pattern"] = pattern
		}
		if len(colConstraints) > 0 {
			constraints[fieldName] = colConstraints
		}
	}
	return constraints
}
func toFloat64(v any) *float64 {
	var result float64
	switch val := v.(type) {
	case float64:
		result = val
	case float32:
		result = float64(val)
	case int:
		result = float64(val)
	case int32:
		result = float64(val)
	case int64:
		result = float64(val)
	default:
		return nil
	}
	return &result
}
func (p *MongoDBPlugin) ClearTableData(config *engine.PluginConfig, schema string, storageUnit string) (bool, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithFields(map[string]any{
			"hostname":    config.Credentials.Hostname,
			"schema":      schema,
			"storageUnit": storageUnit,
		}).Error("Failed to connect to MongoDB to clear collection")
		return false, err
	}
	ctx, cancel := opCtx(config)
	defer cancel()
	defer disconnectClient(client)
	collection := client.Database(schema).Collection(storageUnit)
	result, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.WithError(err).WithFields(map[string]any{
			"schema":      schema,
			"storageUnit": storageUnit,
		}).Error("Failed to clear MongoDB collection")
		return false, err
	}
	log.WithFields(map[string]any{
		"schema":       schema,
		"storageUnit":  storageUnit,
		"deletedCount": result.DeletedCount,
	}).Info("Cleared MongoDB collection for mock data generation")
	return true, nil
}
func init() {
	engine.RegisterPlugin(NewMongoDBPlugin())
}
func NewMongoDBPlugin() *engine.Plugin {
	return &engine.Plugin{
		Type:            engine.DatabaseType_MongoDB,
		PluginFunctions: &MongoDBPlugin{},
	}
}