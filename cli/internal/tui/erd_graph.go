package tui
import "github.com/clidey/whodb/core/src/engine"
type erdForeignKeyTarget struct {
	table  string
	column string
}
func buildERDTablesFromUnits(graphUnits []engine.GraphUnit) []tableWithColumns {
	tables := make([]tableWithColumns, len(graphUnits))
	for i, graphUnit := range graphUnits {
		tables[i] = tableWithColumns{StorageUnit: graphUnit.Unit}
	}
	return tables
}
func buildERDTablesFromGraph(graphUnits []engine.GraphUnit, columnLoader func(storageUnit string) ([]engine.Column, error)) []tableWithColumns {
	fkTargets := buildERDForeignKeyTargets(graphUnits)
	tables := make([]tableWithColumns, 0, len(graphUnits))
	for _, graphUnit := range graphUnits {
		columns, err := columnLoader(graphUnit.Unit.Name)
		if err != nil {
			continue
		}
		tables = append(tables, tableWithColumns{
			StorageUnit: graphUnit.Unit,
			Columns:     applyERDForeignKeyTargets(graphUnit.Unit.Name, columns, fkTargets),
		})
	}
	return tables
}
func buildERDForeignKeyTargets(graphUnits []engine.GraphUnit) map[string]map[string]erdForeignKeyTarget {
	fkTargets := make(map[string]map[string]erdForeignKeyTarget)
	for _, graphUnit := range graphUnits {
		for _, relation := range graphUnit.Relations {
			if relation.SourceColumn == nil || relation.TargetColumn == nil {
				continue
			}
			sourceTable := graphUnit.Unit.Name
			targetTable := relation.Name
			if relation.RelationshipType == "OneToMany" {
				sourceTable = relation.Name
				targetTable = graphUnit.Unit.Name
			}
			if _, ok := fkTargets[sourceTable]; !ok {
				fkTargets[sourceTable] = make(map[string]erdForeignKeyTarget)
			}
			fkTargets[sourceTable][*relation.SourceColumn] = erdForeignKeyTarget{
				table:  targetTable,
				column: *relation.TargetColumn,
			}
		}
	}
	return fkTargets
}
func applyERDForeignKeyTargets(storageUnit string, columns []engine.Column, fkTargets map[string]map[string]erdForeignKeyTarget) []engine.Column {
	targets := fkTargets[storageUnit]
	if len(targets) == 0 {
		return columns
	}
	enriched := make([]engine.Column, len(columns))
	copy(enriched, columns)
	for i := range enriched {
		target, ok := targets[enriched[i].Name]
		if !ok {
			continue
		}
		enriched[i].IsForeignKey = true
		if enriched[i].ReferencedTable == nil {
			tableName := target.table
			enriched[i].ReferencedTable = &tableName
		}
		if enriched[i].ReferencedColumn == nil {
			columnName := target.column
			enriched[i].ReferencedColumn = &columnName
		}
	}
	return enriched
}
func countGraphRelationships(graphUnits []engine.GraphUnit) int {
	count := 0
	for _, graphUnit := range graphUnits {
		count += len(graphUnit.Relations)
	}
	return count
}