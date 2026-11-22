package mockdata
import (
	"fmt"
	"strings"
	"time"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/clidey/whodb/core/src/common"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
)
const (
	DefaultForeignKeyDensityRatio = 20
	NullableFKProbability = 0.2
	RegularNullProbability = 0.1
	MaxPKRetries = 100
	MaxExistingPKsToLoad = 10000
)
var (
	intTypes      = common.IntTypes
	uintTypes     = common.UintTypes
	floatTypes    = common.FloatTypes
	boolTypes     = common.BoolTypes
	dateTypes     = common.DateTypes
	dateTimeTypes = common.DateTimeTypes
	uuidTypes     = common.UuidTypes
	textTypes     = common.TextTypes
	jsonTypes     = common.JsonTypes
)
type TableDependency struct {
	Table            string
	DependsOn        []string
	RowCount         int
	IsBlocked        bool
	UsesExistingData bool
}
type GenerationResult struct {
	TotalGenerated int
	Details        []TableDetail
	Warnings       []string
}
type TableDetail struct {
	Table            string
	RowsGenerated    int
	UsedExistingData bool
}
type DependencyAnalysis struct {
	GenerationOrder []string
	Tables          []TableDependency
	TotalRows       int
	Warnings        []string
	Error           string
}
type Generator struct {
	faker          *gofakeit.Faker
	fkDensityRatio int
	generatedPKs   map[string][]map[string]any
	existingPKs    map[string][]map[string]any
	usedPKValues   map[string]map[string]bool
	databaseType   string
}
func NewGenerator(fkDensityRatio int) *Generator {
	if fkDensityRatio <= 0 {
		fkDensityRatio = DefaultForeignKeyDensityRatio
	}
	return &Generator{
		faker:          gofakeit.New(uint64(time.Now().UnixNano())),
		fkDensityRatio: fkDensityRatio,
		generatedPKs:   make(map[string][]map[string]any),
		existingPKs:    make(map[string][]map[string]any),
		usedPKValues:   make(map[string]map[string]bool),
	}
}
func (g *Generator) AnalyzeDependencies(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, table string,
	rowCount int,
) (*DependencyAnalysis, error) {
	analysis := &DependencyAnalysis{
		Warnings: []string{},
	}
	visited := make(map[string]bool)
	tableInfo := make(map[string]*TableDependency)
	adjacency := make(map[string][]string)
	if err := g.collectDependencies(plugin, config, schema, table, rowCount, visited, tableInfo, adjacency); err != nil {
		analysis.Error = err.Error()
		return analysis, nil
	}
	order, err := topoSort(adjacency)
	if err != nil {
		analysis.Error = err.Error()
		return analysis, nil
	}
	analysis.GenerationOrder = order
	for _, tbl := range order {
		if info, ok := tableInfo[tbl]; ok {
			analysis.Tables = append(analysis.Tables, *info)
			analysis.TotalRows += info.RowCount
			if info.UsesExistingData {
				analysis.Warnings = append(analysis.Warnings,
					fmt.Sprintf("Table '%s' has mock data disabled; will use existing data", tbl))
			}
		}
	}
	return analysis, nil
}
func (g *Generator) collectDependencies(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, table string,
	rowCount int,
	visited map[string]bool,
	tableInfo map[string]*TableDependency,
	adjacency map[string][]string,
) error {
	if visited[table] {
		log.WithField("table", table).Debug("Table already visited, skipping")
		return nil
	}
	visited[table] = true
	log.WithFields(map[string]any{
		"table":    table,
		"rowCount": rowCount,
	}).Debug("Collecting dependencies for table")
	fks, err := plugin.GetForeignKeyRelationships(config, schema, table)
	if err != nil {
		log.WithError(err).WithField("table", table).Warn("Failed to get FK relationships")
		fks = make(map[string]*engine.ForeignKeyRelationship)
	}
	if len(fks) > 0 {
		fkColumns := make([]string, 0, len(fks))
		for col := range fks {
			fkColumns = append(fkColumns, col)
		}
		log.WithFields(map[string]any{
			"table":     table,
			"fkColumns": fkColumns,
			"fkCount":   len(fks),
		}).Debug("Found foreign key relationships")
	}
	isBlocked := !IsMockDataGenerationAllowed(table)
	if isBlocked {
		log.WithField("table", table).Debug("Mock data generation blocked for table")
	}
	info := &TableDependency{
		Table:            table,
		RowCount:         rowCount,
		IsBlocked:        isBlocked,
		UsesExistingData: isBlocked,
		DependsOn:        []string{},
	}
	if _, exists := adjacency[table]; !exists {
		adjacency[table] = []string{}
	}
	for colName, fk := range fks {
		if fk.ReferencedTable == table {
			log.WithFields(map[string]any{
				"table":  table,
				"column": colName,
			}).Debug("Skipping self-referencing FK")
			continue
		}
		info.DependsOn = append(info.DependsOn, fk.ReferencedTable)
		adjacency[table] = append(adjacency[table], fk.ReferencedTable)
		parentRowCount := max(1, rowCount/g.fkDensityRatio)
		log.WithFields(map[string]any{
			"childTable":     table,
			"parentTable":    fk.ReferencedTable,
			"column":         colName,
			"parentRowCount": parentRowCount,
		}).Debug("Following FK to parent table")
		if err := g.collectDependencies(plugin, config, schema, fk.ReferencedTable, parentRowCount, visited, tableInfo, adjacency); err != nil {
			return err
		}
	}
	tableInfo[table] = info
	return nil
}
func topoSort(adjacency map[string][]string) ([]string, error) {
	inDegree := make(map[string]int)
	reverse := make(map[string][]string)
	for node, deps := range adjacency {
		inDegree[node] = len(deps)
		for _, dep := range deps {
			if _, ok := inDegree[dep]; !ok {
				inDegree[dep] = 0
			}
			reverse[dep] = append(reverse[dep], node)
		}
	}
	log.WithFields(map[string]any{
		"inDegree": inDegree,
	}).Debug("Topological sort in-degree calculated")
	var queue []string
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}
	log.WithField("initialQueue", queue).Debug("Topological sort starting with root tables")
	var result []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)
		for _, child := range reverse[node] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}
	if len(result) != len(inDegree) {
		var cycleNodes []string
		for node, degree := range inDegree {
			if degree > 0 {
				cycleNodes = append(cycleNodes, node)
			}
		}
		return nil, fmt.Errorf("circular dependency detected involving tables: %s", strings.Join(cycleNodes, " → "))
	}
	log.WithField("order", result).Debug("Topological sort complete")
	return result, nil
}
func (g *Generator) Generate(
	plugin engine.PluginFunctions,
	config *engine.PluginConfig,
	schema, table string,
	rowCount int,
	overwrite bool,
) (*GenerationResult, error) {
	log.WithFields(map[string]any{
		"schema":    schema,
		"table":     table,
		"rowCount":  rowCount,
		"overwrite": overwrite,
	}).Info("Starting mock data generation")
	g.generatedPKs = make(map[string][]map[string]any)
	g.existingPKs = make(map[string][]map[string]any)
	g.usedPKValues = make(map[string]map[string]bool)
	if config != nil && config.Credentials != nil {
		g.databaseType = config.Credentials.Type
	}
	result := &GenerationResult{
		Details:  []TableDetail{},
		Warnings: []string{},
	}
	log.Debug("Analyzing table dependencies")
	analysis, err := g.AnalyzeDependencies(plugin, config, schema, table, rowCount)
	if err != nil {
		return nil, err
	}
	if analysis.Error != "" {
		log.WithField("error", analysis.Error).Error("Dependency analysis failed")
		return nil, fmt.Errorf("%s", analysis.Error)
	}
	log.WithFields(map[string]any{
		"generationOrder": analysis.GenerationOrder,
		"totalTables":     len(analysis.Tables),
		"totalRows":       analysis.TotalRows,
	}).Info("Dependency analysis complete")
	result.Warnings = analysis.Warnings
	if overwrite {
		log.Debug("Overwrite mode: clearing target table and child tables")
		graph, err := plugin.GetGraph(config, schema)
		if err != nil {
			log.WithError(err).Warn("Failed to get graph for FK-safe clearing, attempting direct clear")
			if _, err := plugin.ClearTableData(config, schema, table); err != nil {
				return nil, fmt.Errorf("failed to clear table %s: %w", table, err)
			}
		} else {
			if err := plugins.ClearTableWithDependencies(plugin, config, schema, table, graph); err != nil {
				return nil, fmt.Errorf("failed to clear tables for overwrite: %w", err)
			}
		}
	}
	err = plugin.WithTransaction(config, func(tx any) error {
		txConfig := &engine.PluginConfig{
			Credentials:   config.Credentials,
			ExternalModel: config.ExternalModel,
			Transaction:   tx,
		}
		for i, tblDep := range analysis.Tables {
			log.WithFields(map[string]any{
				"step":      fmt.Sprintf("%d/%d", i+1, len(analysis.Tables)),
				"table":     tblDep.Table,
				"rowCount":  tblDep.RowCount,
				"dependsOn": tblDep.DependsOn,
			}).Info("Processing table")
			if tblDep.UsesExistingData {
				log.WithField("table", tblDep.Table).Debug("Table uses existing data, loading PKs")
				if err := g.loadExistingPKs(plugin, txConfig, schema, tblDep.Table); err != nil {
					return fmt.Errorf("failed to load existing PKs for %s: %w", tblDep.Table, err)
				}
				result.Details = append(result.Details, TableDetail{
					Table:            tblDep.Table,
					RowsGenerated:    0,
					UsedExistingData: true,
				})
				continue
			}
			isTargetTable := tblDep.Table == table
			generated, err := g.generateTableRows(plugin, txConfig, schema, tblDep.Table, tblDep.RowCount, overwrite && isTargetTable, isTargetTable)
			if err != nil {
				return fmt.Errorf("failed to generate rows for %s: %w", tblDep.Table, err)
			}
			result.TotalGenerated += generated
			result.Details = append(result.Details, TableDetail{
				Table:            tblDep.Table,
				RowsGenerated:    generated,
				UsedExistingData: false,
			})
		}
		return nil
	})
	if err != nil {
		log.WithError(err).Error("Mock data generation failed, transaction rolled back")
		return nil, err
	}
	log.WithFields(map[string]any{
		"totalGenerated": result.TotalGenerated,
		"tablesCount":    len(result.Details),
	}).Info("Mock data generation complete")
	return result, nil
}