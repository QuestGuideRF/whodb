package mockdata
import (
	"fmt"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	gorm_plugin "github.com/clidey/whodb/core/src/plugins/gorm"
)
func (g *Generator) generateTableRows(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, table string,
	rowCount int,
	overwrite bool,
	isTargetTable bool,
) (int, error) {
	insertMode := "single"
	if isTargetTable {
		insertMode = "bulk"
	}
	log.WithFields(map[string]any{
		"table":      table,
		"rowCount":   rowCount,
		"overwrite":  overwrite,
		"insertMode": insertMode,
	}).Info("Generating mock data for table")
	columns, err := plugin.GetColumnsForTable(config, schema, table)
	if err != nil {
		return 0, fmt.Errorf("failed to get columns: %w", err)
	}
	if err := plugin.MarkGeneratedColumns(config, schema, table, columns); err != nil {
		log.WithError(err).Warn("Failed to mark generated columns for " + table)
	}
	columnDetails := make([]map[string]any, 0, len(columns))
	for _, col := range columns {
		if col.IsPrimary || col.IsAutoIncrement || col.IsComputed {
			columnDetails = append(columnDetails, map[string]any{
				"name":       col.Name,
				"type":       col.Type,
				"isPrimary":  col.IsPrimary,
				"isAutoIncr": col.IsAutoIncrement,
				"isComputed": col.IsComputed,
			})
		}
	}
	log.WithFields(map[string]any{
		"table":          table,
		"columnCount":    len(columns),
		"specialColumns": columnDetails,
	}).Debug("Retrieved column definitions")
	constraints, err := plugin.GetColumnConstraints(config, schema, table)
	if err != nil {
		log.WithError(err).WithField("table", table).Warn("Failed to get constraints, using defaults")
		constraints = make(map[string]map[string]any)
	} else {
		constraintDetails := make(map[string][]string)
		for col, c := range constraints {
			var details []string
			if checkVals, ok := c["check_values"]; ok {
				details = append(details, fmt.Sprintf("check_values=%v", checkVals))
			}
			if checkMin, ok := c["check_min"]; ok {
				details = append(details, fmt.Sprintf("check_min=%v", checkMin))
			}
			if checkMax, ok := c["check_max"]; ok {
				details = append(details, fmt.Sprintf("check_max=%v", checkMax))
			}
			if len(details) > 0 {
				constraintDetails[col] = details
			}
		}
		log.WithFields(map[string]any{
			"table":            table,
			"constraintCount":  len(constraints),
			"checkConstraints": constraintDetails,
		}).Debug("Retrieved column constraints")
	}
	fks, err := plugin.GetForeignKeyRelationships(config, schema, table)
	if err != nil {
		log.WithError(err).WithField("table", table).Warn("Failed to get FK relationships")
		fks = make(map[string]*engine.ForeignKeyRelationship)
	} else if len(fks) > 0 {
		log.WithFields(map[string]any{
			"table":   table,
			"fkCount": len(fks),
		}).Debug("Retrieved FK relationships")
	}
	if !overwrite {
		if err := g.loadExistingPKsForUniqueness(plugin, config, schema, table, columns); err != nil {
			log.WithError(err).WithField("table", table).Warn("Failed to load existing PKs for uniqueness check")
		}
	}
	if isTargetTable {
		return g.generateRowsBulk(plugin, config, schema, table, columns, constraints, fks, rowCount)
	}
	return g.generateRowsSingle(plugin, config, schema, table, columns, constraints, fks, rowCount)
}
func (g *Generator) generateRowsBulk(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, table string,
	columns []engine.Column,
	constraints map[string]map[string]any,
	fks map[string]*engine.ForeignKeyRelationship,
	rowCount int,
) (int, error) {
	log.WithFields(map[string]any{
		"table":    table,
		"rowCount": rowCount,
	}).Debug("Starting bulk row generation")
	rows := make([][]engine.Record, 0, rowCount)
	failedCount := 0
	pkCollisions := 0
	for i := 0; i < rowCount; i++ {
		var row []engine.Record
		var err error
		for attempt := 0; attempt < MaxPKRetries; attempt++ {
			row, err = g.generateRow(columns, constraints, fks, table)
			if err != nil {
				break
			}
			if !g.isPKValueUsed(table, columns, row) {
				g.markPKValueUsed(table, columns, row)
				break
			}
			if attempt == MaxPKRetries-1 {
				err = fmt.Errorf("failed to generate unique PK after %d attempts", MaxPKRetries)
				pkCollisions++
			}
		}
		if err != nil {
			log.WithError(err).WithField("table", table).WithField("row", i).Error("Failed to generate row")
			failedCount++
			continue
		}
		rows = append(rows, row)
	}
	if failedCount > 0 {
		log.WithFields(map[string]any{
			"table":        table,
			"failedCount":  failedCount,
			"pkCollisions": pkCollisions,
		}).Warn("Some rows failed to generate")
	}
	if len(rows) == 0 {
		log.WithField("table", table).Warn("No rows generated")
		return 0, nil
	}
	log.WithFields(map[string]any{
		"table":    table,
		"rowCount": len(rows),
	}).Debug("Generated rows, starting bulk insert")
	if _, err := plugin.BulkAddRows(config, schema, table, rows); err != nil {
		log.WithError(err).WithField("table", table).WithField("rowCount", len(rows)).Error("Failed to bulk insert rows")
		return 0, fmt.Errorf("failed to bulk insert rows: %w", err)
	}
	log.WithFields(map[string]any{
		"table":     table,
		"generated": len(rows),
	}).Info("Completed bulk mock data generation")
	return len(rows), nil
}
func (g *Generator) generateRowsSingle(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, table string,
	columns []engine.Column,
	constraints map[string]map[string]any,
	fks map[string]*engine.ForeignKeyRelationship,
	rowCount int,
) (int, error) {
	log.WithFields(map[string]any{
		"table":    table,
		"rowCount": rowCount,
	}).Debug("Starting single-row generation (PK tracking enabled)")
	var autoIncrementPKCol *engine.Column
	for i := range columns {
		if columns[i].IsPrimary && columns[i].IsAutoIncrement {
			autoIncrementPKCol = &columns[i]
			log.WithFields(map[string]any{
				"table":  table,
				"column": columns[i].Name,
			}).Debug("Found auto-increment PK column for tracking")
			break
		}
	}
	generated := 0
	failedGenerate := 0
	failedInsert := 0
	pkCollisions := 0
	for i := 0; i < rowCount; i++ {
		var row []engine.Record
		var err error
		for attempt := 0; attempt < MaxPKRetries; attempt++ {
			row, err = g.generateRow(columns, constraints, fks, table)
			if err != nil {
				break
			}
			if autoIncrementPKCol == nil && g.isPKValueUsed(table, columns, row) {
				if attempt == MaxPKRetries-1 {
					err = fmt.Errorf("failed to generate unique PK after %d attempts", MaxPKRetries)
					pkCollisions++
				}
				continue
			}
			if autoIncrementPKCol == nil {
				g.markPKValueUsed(table, columns, row)
			}
			break
		}
		if err != nil {
			log.WithError(err).WithField("table", table).WithField("row", i).Error("Failed to generate row")
			failedGenerate++
			continue
		}
		generatedID, err := plugin.AddRowReturningID(config, schema, table, row)
		if err != nil {
			log.WithError(err).WithField("table", table).WithField("row", i).Error("Failed to insert row")
			return 0, fmt.Errorf("failed to insert row %d into %s: %w", i, table, err)
		}
		if autoIncrementPKCol != nil {
			if generatedID > 0 {
				g.trackAutoGeneratedPK(table, autoIncrementPKCol.Name, generatedID)
			} else {
				log.WithFields(map[string]any{
					"table":  table,
					"column": autoIncrementPKCol.Name,
					"row":    i,
				}).Warn("Failed to get auto-generated PK - FK references may fail")
			}
		} else {
			g.trackGeneratedPK(table, columns, row)
		}
		generated++
		if rowCount >= 100 && generated%100 == 0 {
			log.WithFields(map[string]any{
				"table":    table,
				"progress": fmt.Sprintf("%d/%d", generated, rowCount),
			}).Debug("Row generation progress")
		}
	}
	if failedGenerate > 0 || failedInsert > 0 {
		log.WithFields(map[string]any{
			"table":          table,
			"failedGenerate": failedGenerate,
			"failedInsert":   failedInsert,
			"pkCollisions":   pkCollisions,
		}).Warn("Some rows failed during generation")
	}
	log.WithFields(map[string]any{
		"table":      table,
		"generated":  generated,
		"trackedPKs": len(g.generatedPKs[table]),
	}).Info("Completed single-row mock data generation")
	return generated, nil
}
func (g *Generator) generateRow(
	columns []engine.Column,
	constraints map[string]map[string]any,
	fks map[string]*engine.ForeignKeyRelationship,
	table string,
) ([]engine.Record, error) {
	records := make([]engine.Record, 0, len(columns))
	selectedParentRows := make(map[string]map[string]any)
	skippedAutoIncr := 0
	skippedComputed := 0
	skippedNil := 0
	for _, col := range columns {
		if col.IsAutoIncrement {
			log.WithFields(map[string]any{
				"table":  table,
				"column": col.Name,
			}).Debug("Skipping auto-increment column")
			skippedAutoIncr++
			continue
		}
		if col.IsComputed {
			log.WithFields(map[string]any{
				"table":  table,
				"column": col.Name,
				"type":   col.Type,
			}).Debug("Skipping computed column")
			skippedComputed++
			continue
		}
		colConstraints := getConstraintsForColumn(constraints, col.Name)
		var value any
		var err error
		if fk, isFk := fks[col.Name]; isFk {
			value, err = g.generateFKValue(col, colConstraints, fk, table, selectedParentRows)
			if err != nil {
				return nil, fmt.Errorf("failed to generate FK value for column %s: %w", col.Name, err)
			}
		} else {
			value = g.generateColumnValue(col, colConstraints)
		}
		if value == nil {
			skippedNil++
			continue
		}
		constraintType := ""
		if colConstraints != nil {
			if t, ok := colConstraints["type"].(string); ok {
				constraintType = t
			}
		}
		record := valueToRecord(col, value, constraintType)
		records = append(records, record)
	}
	if len(records) == 0 {
		log.WithFields(map[string]any{
			"table":           table,
			"totalColumns":    len(columns),
			"skippedAutoIncr": skippedAutoIncr,
			"skippedComputed": skippedComputed,
			"skippedNil":      skippedNil,
		}).Warn("No records generated - all columns were skipped")
	}
	return records, nil
}
func (g *Generator) generateFKValue(
	col engine.Column,
	constraints map[string]any,
	fk *engine.ForeignKeyRelationship,
	currentTable string,
	selectedParentRows map[string]map[string]any,
) (any, error) {
	isNullable := gorm_plugin.IsNullable(constraints)
	log.WithFields(map[string]any{
		"table":      currentTable,
		"column":     col.Name,
		"isNullable": isNullable,
		"fkTarget":   fk.ReferencedTable,
	}).Debug("Generating FK value")
	if fk.ReferencedTable == currentTable {
		if isNullable {
			log.WithFields(map[string]any{
				"table":  currentTable,
				"column": col.Name,
			}).Debug("Self-referencing FK set to NULL")
			return nil, nil
		}
		return nil, fmt.Errorf("cannot generate data for non-nullable self-referencing FK column %s", col.Name)
	}
	if selectedRow, exists := selectedParentRows[fk.ReferencedTable]; exists {
		if selectedRow == nil {
			return nil, nil
		}
		value := selectedRow[fk.ReferencedColumn]
		log.WithFields(map[string]any{
			"table":            currentTable,
			"column":           col.Name,
			"parentTable":      fk.ReferencedTable,
			"referencedColumn": fk.ReferencedColumn,
			"value":            value,
		}).Debug("Using cached parent row for composite FK")
		return value, nil
	}
	if isNullable && g.faker.Float64() < NullableFKProbability {
		log.WithFields(map[string]any{
			"table":  currentTable,
			"column": col.Name,
		}).Debug("Nullable FK set to NULL (random)")
		selectedParentRows[fk.ReferencedTable] = nil
		return nil, nil
	}
	parentRows := g.getPKRowsForTable(fk.ReferencedTable)
	log.WithFields(map[string]any{
		"column":          col.Name,
		"referencedTable": fk.ReferencedTable,
		"parentRowCount":  len(parentRows),
	}).Debug("Looking up parent PKs for FK")
	if len(parentRows) == 0 {
		if isNullable {
			log.WithFields(map[string]any{
				"column":          col.Name,
				"referencedTable": fk.ReferencedTable,
			}).Debug("No parent PKs available, using NULL for nullable FK")
			selectedParentRows[fk.ReferencedTable] = nil
			return nil, nil
		}
		return nil, fmt.Errorf("no parent rows available for non-nullable FK column %s (references %s)", col.Name, fk.ReferencedTable)
	}
	idx := g.faker.Number(0, len(parentRows)-1)
	selectedRow := parentRows[idx]
	selectedParentRows[fk.ReferencedTable] = selectedRow
	value := selectedRow[fk.ReferencedColumn]
	log.WithFields(map[string]any{
		"table":            currentTable,
		"column":           col.Name,
		"parentTable":      fk.ReferencedTable,
		"referencedColumn": fk.ReferencedColumn,
		"value":            value,
		"availableRows":    len(parentRows),
	}).Debug("Selected FK value from parent table")
	return value, nil
}
func (g *Generator) generateColumnValue(col engine.Column, constraints map[string]any) any {
	if gorm_plugin.IsNullable(constraints) {
		if g.faker.Float64() < RegularNullProbability {
			return nil
		}
	}
	if constraints != nil {
		if isJSON, ok := constraints["is_json"].(bool); ok && isJSON {
			return GenerateByType("json", g.databaseType, nil, g.faker)
		}
	}
	effectiveConstraints := constraints
	needsCopy := false
	if col.Length != nil && *col.Length > 0 && (constraints == nil || constraints["length"] == nil) {
		needsCopy = true
	}
	if col.Precision != nil && *col.Precision > 0 && (constraints == nil || constraints["precision"] == nil) {
		needsCopy = true
	}
	if needsCopy {
		effectiveConstraints = make(map[string]any)
		for k, v := range constraints {
			effectiveConstraints[k] = v
		}
		if col.Length != nil && *col.Length > 0 {
			effectiveConstraints["length"] = *col.Length
		}
		if col.Precision != nil && *col.Precision > 0 {
			effectiveConstraints["precision"] = *col.Precision
		}
		if col.Scale != nil && *col.Scale >= 0 {
			effectiveConstraints["scale"] = *col.Scale
		}
	}
	dbType := detectDatabaseType(col.Type)
	if dbType == "text" {
		if effectiveConstraints != nil {
			if _, hasCheckValues := effectiveConstraints["check_values"]; hasCheckValues {
				return GenerateByType(col.Type, g.databaseType, effectiveConstraints, g.faker)
			}
		}
		maxLen := 0
		if col.Length != nil {
			maxLen = *col.Length
		}
		if val, matched := MatchColumnName(col.Name, maxLen, g.faker); matched {
			return val
		}
	}
	return GenerateByType(col.Type, g.databaseType, effectiveConstraints, g.faker)
}