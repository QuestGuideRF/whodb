package gorm_plugin
import (
	"fmt"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
)
type tableRelations struct {
	Table1       string
	Table2       string
	Relation     string
	SourceColumn *string
	TargetColumn *string
}
func (p *GormPlugin) GetGraph(config *engine.PluginConfig, schema string) ([]engine.GraphUnit, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) ([]engine.GraphUnit, error) {
		var tableRelations []tableRelations
		schema = p.GormPluginFunctions.ResolveGraphSchema(config, schema)
		if err := p.GetGraphQueryDB(db, schema).Scan(&tableRelations).Error; err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to execute graph query for schema: %s", schema))
			return nil, err
		}
		tableMap := make(map[string][]engine.GraphUnitRelationship)
		for _, tr := range tableRelations {
			tableMap[tr.Table1] = append(tableMap[tr.Table1], engine.GraphUnitRelationship{
				Name:             tr.Table2,
				RelationshipType: engine.GraphUnitRelationshipType(tr.Relation),
				SourceColumn:     tr.SourceColumn,
				TargetColumn:     tr.TargetColumn,
			})
		}
		storageUnits, err := p.GetStorageUnits(config, schema)
		if err != nil {
			log.WithError(err).Error(fmt.Sprintf("Failed to get storage units for graph generation in schema: %s", schema))
			return nil, err
		}
		storageUnitsMap := map[string]engine.StorageUnit{}
		for _, storageUnit := range storageUnits {
			storageUnitsMap[storageUnit.Name] = storageUnit
		}
		var tables []engine.GraphUnit
		for _, storageUnit := range storageUnits {
			foundTable, ok := tableMap[storageUnit.Name]
			var relations []engine.GraphUnitRelationship
			if ok {
				relations = foundTable
			}
			tables = append(tables, engine.GraphUnit{Unit: storageUnit, Relations: relations})
		}
		return tables, nil
	})
}