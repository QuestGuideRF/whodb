package redis
import (
	"context"
	"fmt"
	"strconv"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/go-redis/redis/v8"
)
func (p *RedisPlugin) AddStorageUnit(config *engine.PluginConfig, schema string, storageUnit string, fields []engine.Record) (bool, error) {
	client, err := DB(config)
	if err != nil {
		return false, err
	}
	defer client.Close()
	ctx := context.Background()
	keyType := "string"
	if len(fields) > 0 {
		if t, ok := fields[0].Extra["type"]; ok && t != "" {
			keyType = t
		}
	}
	switch keyType {
	case "string":
		value := ""
		if len(fields) > 0 {
			value = fields[0].Value
		}
		if err := client.Set(ctx, storageUnit, value, 0).Err(); err != nil {
			log.WithError(err).WithField("key", storageUnit).Error("Failed to create Redis string key")
			return false, err
		}
	case "hash":
		fieldValues := map[string]any{}
		for _, f := range fields {
			fieldValues[f.Key] = f.Value
		}
		if len(fieldValues) == 0 {
			fieldValues["default"] = ""
		}
		if err := client.HSet(ctx, storageUnit, fieldValues).Err(); err != nil {
			log.WithError(err).WithField("key", storageUnit).Error("Failed to create Redis hash key")
			return false, err
		}
	case "list":
		values := make([]any, 0, len(fields))
		for _, f := range fields {
			values = append(values, f.Value)
		}
		if len(values) == 0 {
			values = append(values, "")
		}
		if err := client.RPush(ctx, storageUnit, values...).Err(); err != nil {
			log.WithError(err).WithField("key", storageUnit).Error("Failed to create Redis list key")
			return false, err
		}
	case "set":
		members := make([]any, 0, len(fields))
		for _, f := range fields {
			members = append(members, f.Value)
		}
		if len(members) == 0 {
			members = append(members, "")
		}
		if err := client.SAdd(ctx, storageUnit, members...).Err(); err != nil {
			log.WithError(err).WithField("key", storageUnit).Error("Failed to create Redis set key")
			return false, err
		}
	case "zset":
		for _, f := range fields {
			score, parseErr := strconv.ParseFloat(f.Value, 64)
			if parseErr != nil {
				score = 0
			}
			if err := client.ZAdd(ctx, storageUnit, &redis.Z{Score: score, Member: f.Key}).Err(); err != nil {
				log.WithError(err).WithField("key", storageUnit).Error("Failed to create Redis sorted set key")
				return false, err
			}
		}
	default:
		return false, fmt.Errorf("unsupported Redis key type: %s", keyType)
	}
	return true, nil
}
func (p *RedisPlugin) AddRow(config *engine.PluginConfig, schema string, storageUnit string, values []engine.Record) (bool, error) {
	client, err := DB(config)
	if err != nil {
		return false, err
	}
	defer client.Close()
	ctx := context.Background()
	keyType, err := client.Type(ctx, storageUnit).Result()
	if err != nil {
		return false, err
	}
	valuesMap := map[string]string{}
	for _, v := range values {
		valuesMap[v.Key] = v.Value
	}
	switch keyType {
	case "string":
		value := valuesMap["value"]
		if err := client.Set(ctx, storageUnit, value, 0).Err(); err != nil {
			return false, err
		}
	case "hash":
		field := valuesMap["field"]
		value := valuesMap["value"]
		if field == "" {
			return false, fmt.Errorf("field name is required for hash")
		}
		if err := client.HSet(ctx, storageUnit, field, value).Err(); err != nil {
			return false, err
		}
	case "list":
		value := valuesMap["value"]
		if err := client.RPush(ctx, storageUnit, value).Err(); err != nil {
			return false, err
		}
	case "set":
		value := valuesMap["value"]
		if err := client.SAdd(ctx, storageUnit, value).Err(); err != nil {
			return false, err
		}
	case "zset":
		member := valuesMap["member"]
		scoreStr := valuesMap["score"]
		score, parseErr := strconv.ParseFloat(scoreStr, 64)
		if parseErr != nil {
			score = 0
		}
		if err := client.ZAdd(ctx, storageUnit, &redis.Z{Score: score, Member: member}).Err(); err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("unsupported Redis data type: %s", keyType)
	}
	return true, nil
}