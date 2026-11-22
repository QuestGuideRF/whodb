package plugins
import (
	"fmt"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
)
func FindChildTables(graph []engine.GraphUnit, targetTable string) []string {
	for _, unit := range graph {
		if unit.Unit.Name == targetTable {
			var children []string
			for _, rel := range unit.Relations {
				children = append(children, rel.Name)
			}
			return children
		}
	}
	return nil
}
func NullifyFKColumn(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, childTable, parentTable string,
	graph []engine.GraphUnit,
) error {
	for _, unit := range graph {
		if unit.Unit.Name == parentTable {
			for _, rel := range unit.Relations {
				if rel.Name == childTable && rel.SourceColumn != nil {
					return plugin.NullifyFKColumn(config, schema, childTable, *rel.SourceColumn)
				}
			}
			return nil
		}
	}
	return nil
}
func ClearTableWithDependencies(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, table string,
	graph []engine.GraphUnit,
) error {
	cleared := make(map[string]bool)
	return clearTableRecursive(plugin, config, schema, table, graph, cleared)
}
func clearTableRecursive(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, table string,
	graph []engine.GraphUnit,
	cleared map[string]bool,
) error {
	if cleared[table] {
		return nil
	}
	cleared[table] = true
	children := FindChildTables(graph, table)
	for _, child := range children {
		if cleared[child] {
			if err := NullifyFKColumn(plugin, config, schema, child, table, graph); err != nil {
				log.WithFields(map[string]any{
					"child":  child,
					"parent": table,
				}).WithError(err).Error("Failed to nullify FK column for cycle breaking")
			}
			continue
		}
		if err := clearTableRecursive(plugin, config, schema, child, graph, cleared); err != nil {
			return err
		}
	}
	log.WithFields(map[string]any{
		"table":    table,
		"children": children,
	}).Debug("Clearing table data (children already cleared)")
	if _, err := plugin.ClearTableData(config, schema, table); err != nil {
		return fmt.Errorf("failed to clear table %s: %w", table, err)
	}
	return nil
}