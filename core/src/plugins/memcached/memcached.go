package memcached
import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/query"
)
type MemcachedPlugin struct {
	engine.BasePlugin
}
func (p *MemcachedPlugin) IsAvailable(ctx context.Context, config *engine.PluginConfig) bool {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).Error("Failed to connect to Memcached for availability check")
		return false
	}
	defer client.Close()
	return client.Ping() == nil
}
func (p *MemcachedPlugin) GetDatabases(config *engine.PluginConfig) ([]string, error) {
	return []string{"default"}, nil
}
func (p *MemcachedPlugin) GetStorageUnits(config *engine.PluginConfig, schema string) ([]engine.StorageUnit, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).Error("Failed to connect to Memcached for storage units retrieval")
		return nil, err
	}
	defer client.Close()
	entries, err := client.Metadump()
	if err != nil {
		log.WithError(err).Error("Failed to enumerate Memcached keys via metadump")
		return nil, err
	}
	storageUnits := make([]engine.StorageUnit, 0, len(entries))
	for _, entry := range entries {
		if entry.Key == "auth" {
			continue
		}
		expValue := "never"
		if entry.Expiration > 0 {
			expValue = time.Unix(entry.Expiration, 0).UTC().Format(time.RFC3339)
		}
		storageUnits = append(storageUnits, engine.StorageUnit{
			Name: entry.Key,
			Attributes: []engine.Record{
				{Key: "Type", Value: "string"},
				{Key: "Size", Value: strconv.Itoa(entry.Size)},
				{Key: "Expires", Value: expValue},
			},
		})
	}
	return storageUnits, nil
}
func (p *MemcachedPlugin) StorageUnitExists(config *engine.PluginConfig, schema string, storageUnit string) (bool, error) {
	client, err := DB(config)
	if err != nil {
		return false, err
	}
	defer client.Close()
	item, err := client.Get(storageUnit)
	if err != nil {
		return false, err
	}
	return item != nil, nil
}
func (p *MemcachedPlugin) GetRows(
	config *engine.PluginConfig,
	req *engine.GetRowsRequest,
) (*engine.GetRowsResult, error) {
	storageUnit := req.StorageUnit
	where := req.Where
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to connect to Memcached for rows retrieval")
		return nil, err
	}
	defer client.Close()
	item, err := client.Get(storageUnit)
	if err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to get Memcached item")
		return nil, err
	}
	if item == nil {
		return &engine.GetRowsResult{
			Columns:    memcachedColumns(),
			Rows:       [][]string{},
			TotalCount: 0,
		}, nil
	}
	row := []string{
		string(item.Value),
		strconv.FormatUint(uint64(item.Flags), 10),
	}
	rows := [][]string{row}
	if where != nil {
		rows = filterMemcachedRows(rows, where)
	}
	return &engine.GetRowsResult{
		Columns:    memcachedColumns(),
		Rows:       rows,
		TotalCount: int64(len(rows)),
	}, nil
}
func (p *MemcachedPlugin) GetRowCount(config *engine.PluginConfig, schema, storageUnit string, where *query.WhereCondition) (int64, error) {
	client, err := DB(config)
	if err != nil {
		return 0, err
	}
	defer client.Close()
	item, err := client.Get(storageUnit)
	if err != nil {
		return 0, err
	}
	if item == nil {
		return 0, nil
	}
	return 1, nil
}
func (p *MemcachedPlugin) GetColumnsForTable(config *engine.PluginConfig, schema string, storageUnit string) ([]engine.Column, error) {
	return memcachedColumns(), nil
}
func (p *MemcachedPlugin) FormatValue(val any) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}
func init() {
	engine.RegisterPlugin(NewMemcachedPlugin())
}
func NewMemcachedPlugin() *engine.Plugin {
	return &engine.Plugin{
		Type:            engine.DatabaseType_Memcached,
		PluginFunctions: &MemcachedPlugin{},
	}
}
func memcachedColumns() []engine.Column {
	return []engine.Column{
		{Name: "Value", Type: "string"},
		{Name: "Flags", Type: "uint32"},
	}
}
func filterMemcachedRows(rows [][]string, where *query.WhereCondition) [][]string {
	condition, err := convertWhereCondition(where)
	if err != nil {
		return rows
	}
	var filtered [][]string
	columnIndex := map[string]int{
		"value": 0,
		"flags": 1,
	}
	for _, row := range rows {
		match := true
		for col, filter := range condition {
			idx, ok := columnIndex[strings.ToLower(col)]
			if !ok {
				continue
			}
			if !evaluateCondition(row[idx], filter.Operator, filter.Value) {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, row)
		}
	}
	return filtered
}
type memcachedFilter struct {
	Operator string
	Value    string
}
func convertWhereCondition(where *query.WhereCondition) (map[string]memcachedFilter, error) {
	if where == nil {
		return nil, nil
	}
	switch where.Type {
	case query.WhereConditionTypeAtomic:
		if where.Atomic == nil {
			return nil, fmt.Errorf("atomic condition must have an atomicwherecondition")
		}
		return map[string]memcachedFilter{
			where.Atomic.Key: {
				Operator: strings.ToUpper(where.Atomic.Operator),
				Value:    where.Atomic.Value,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Memcached filtering condition type: %v", where.Type)
	}
}
func evaluateCondition(value, operator, target string) bool {
	switch operator {
	case "=", "EQ":
		return value == target
	case "!=", "NE", "<>":
		return value != target
	case ">":
		return value > target
	case ">=":
		return value >= target
	case "<":
		return value < target
	case "<=":
		return value <= target
	case "CONTAINS":
		return strings.Contains(value, target)
	case "STARTS WITH":
		return strings.HasPrefix(value, target)
	case "ENDS WITH":
		return strings.HasSuffix(value, target)
	case "IN":
		parts := strings.Split(target, ",")
		for _, p := range parts {
			if value == strings.TrimSpace(p) {
				return true
			}
		}
		return false
	case "NOT IN":
		parts := strings.Split(target, ",")
		for _, p := range parts {
			if value == strings.TrimSpace(p) {
				return false
			}
		}
		return true
	default:
		return false
	}
}