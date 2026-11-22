package graphutil
import (
	"strings"
	"github.com/clidey/whodb/core/src/engine"
)
type Relation struct {
	Table1       string
	Table2       string
	Relation     string
	SourceColumn string
	TargetColumn string
}
func InferForeignKeys(currentUnit string, fieldNames []string, allUnits []string) map[string]string {
	foreignKeys := make(map[string]string)
	for _, fieldName := range fieldNames {
		if fieldName == "_id" {
			continue
		}
		lowerField := strings.ToLower(fieldName)
		if strings.HasSuffix(lowerField, ".id") || strings.HasSuffix(lowerField, "id") || strings.HasSuffix(lowerField, "_id") {
			for _, otherUnit := range allUnits {
				if otherUnit == currentUnit {
					continue
				}
				if strings.Contains(lowerField, strings.ToLower(otherUnit)) {
					foreignKeys[otherUnit] = fieldName
					break
				}
			}
		}
		for _, otherUnit := range allUnits {
			if otherUnit == currentUnit {
				continue
			}
			singularName := strings.TrimSuffix(otherUnit, "s")
			pluralName := otherUnit
			if !strings.HasSuffix(otherUnit, "s") {
				pluralName = otherUnit + "s"
			}
			if lowerField == strings.ToLower(singularName)+"_id" ||
				lowerField == strings.ToLower(singularName)+"id" ||
				lowerField == strings.ToLower(otherUnit)+"_id" ||
				lowerField == strings.ToLower(otherUnit)+"id" ||
				lowerField == strings.ToLower(pluralName)+"_id" ||
				lowerField == strings.ToLower(pluralName)+"id" {
				foreignKeys[otherUnit] = fieldName
				break
			}
		}
	}
	return foreignKeys
}
func BuildGraphUnits(relations []Relation, storageUnits []engine.StorageUnit) []engine.GraphUnit {
	tableMap := make(map[string][]engine.GraphUnitRelationship)
	for _, tr := range relations {
		tableMap[tr.Table1] = append(tableMap[tr.Table1], engine.GraphUnitRelationship{
			Name:             tr.Table2,
			RelationshipType: engine.GraphUnitRelationshipType(tr.Relation),
			SourceColumn:     new(tr.SourceColumn),
			TargetColumn:     new(tr.TargetColumn),
		})
	}
	tables := make([]engine.GraphUnit, 0, len(storageUnits))
	for _, storageUnit := range storageUnits {
		foundTable, ok := tableMap[storageUnit.Name]
		var rels []engine.GraphUnitRelationship
		if ok {
			rels = foundTable
		}
		tables = append(tables, engine.GraphUnit{Unit: storageUnit, Relations: rels})
	}
	return tables
}