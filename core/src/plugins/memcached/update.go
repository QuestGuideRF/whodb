package memcached
import (
	"strconv"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *MemcachedPlugin) UpdateStorageUnit(config *engine.PluginConfig, schema string, storageUnit string, values map[string]string, updatedColumns []string) (bool, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to connect to Memcached for update")
		return false, err
	}
	defer client.Close()
	current, err := client.Get(storageUnit)
	if err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to get current Memcached item for update")
		return false, err
	}
	if current == nil {
		return false, nil
	}
	item := &Item{
		Key:   storageUnit,
		Value: current.Value,
		Flags: current.Flags,
	}
	if val, ok := values["Value"]; ok {
		item.Value = []byte(val)
	}
	if flagStr, ok := values["Flags"]; ok {
		flags, err := strconv.ParseUint(flagStr, 10, 32)
		if err == nil {
			item.Flags = uint32(flags)
		}
	}
	if err := client.Replace(item); err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to update Memcached item")
		return false, err
	}
	return true, nil
}