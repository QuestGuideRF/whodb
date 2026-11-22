package memcached
import (
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func (p *MemcachedPlugin) DeleteRow(config *engine.PluginConfig, schema string, storageUnit string, values map[string]string) (bool, error) {
	client, err := DB(config)
	if err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to connect to Memcached for row deletion")
		return false, err
	}
	defer client.Close()
	if err := client.Delete(storageUnit); err != nil {
		log.WithError(err).WithField("storageUnit", storageUnit).Error("Failed to delete Memcached item")
		return false, err
	}
	return true, nil
}