package memcached
import (
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
)
func (p *MemcachedPlugin) ExportData(config *engine.PluginConfig, schema string, storageUnit string, writer func([]string) error, selectedRows []map[string]any) error {
	headers := []string{
		common.FormatCSVHeader("Value", "string"),
		common.FormatCSVHeader("Flags", "uint32"),
	}
	if err := writer(headers); err != nil {
		return err
	}
	if len(selectedRows) > 0 {
		for _, row := range selectedRows {
			if err := writer([]string{
				fmt.Sprintf("%v", row["Value"]),
				fmt.Sprintf("%v", row["Flags"]),
			}); err != nil {
				return err
			}
		}
		return nil
	}
	client, err := DB(config)
	if err != nil {
		return err
	}
	defer client.Close()
	item, err := client.Get(storageUnit)
	if err != nil {
		return err
	}
	if item == nil {
		return nil
	}
	return writer([]string{
		string(item.Value),
		strconv.FormatUint(uint64(item.Flags), 10),
	})
}
func (p *MemcachedPlugin) ExportDataNDJSON(config *engine.PluginConfig, schema string, storageUnit string, writer func(string) error, selectedRows []map[string]any) error {
	emit := func(rows []map[string]any) error {
		for _, row := range rows {
			line, err := json.Marshal(row)
			if err != nil {
				return err
			}
			if err := writer(string(line)); err != nil {
				return err
			}
		}
		return nil
	}
	if len(selectedRows) > 0 {
		return emit(selectedRows)
	}
	client, err := DB(config)
	if err != nil {
		return err
	}
	defer client.Close()
	item, err := client.Get(storageUnit)
	if err != nil {
		return err
	}
	if item == nil {
		return nil
	}
	return emit([]map[string]any{{
		"Value": string(item.Value),
		"Flags": item.Flags,
	}})
}