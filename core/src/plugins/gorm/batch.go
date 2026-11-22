package gorm_plugin
import (
	"fmt"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)
type BatchConfig struct {
	BatchSize             int
	UseBulkInsert         bool
	FailOnError           bool
	LogProgress           bool
	UpsertPKColumns       []string
	SkipConflictPKColumns []string
}
func DefaultBatchConfig() *BatchConfig {
	return &BatchConfig{
		BatchSize:     1000,
		UseBulkInsert: true,
		FailOnError:   true,
		LogProgress:   false,
	}
}
type BatchProcessor struct {
	plugin GormPluginFunctions
	config *BatchConfig
	dbType engine.DatabaseType
}
func NewBatchProcessor(plugin GormPluginFunctions, dbType engine.DatabaseType, config *BatchConfig) *BatchProcessor {
	if config == nil {
		config = DefaultBatchConfig()
	}
	return &BatchProcessor{
		plugin: plugin,
		config: config,
		dbType: dbType,
	}
}
func (b *BatchProcessor) getMaxParametersForDB() int {
	return b.plugin.GetMaxBulkInsertParameters()
}
func (b *BatchProcessor) calculateBatchSize(columnCount int) int {
	if columnCount == 0 {
		return b.config.BatchSize
	}
	maxParams := b.getMaxParametersForDB()
	safeMaxParams := int(float64(maxParams) * 0.9)
	maxRowsForDB := safeMaxParams / columnCount
	if maxRowsForDB < 1 {
		maxRowsForDB = 1
	}
	if maxRowsForDB < b.config.BatchSize {
		log.WithFields(map[string]any{
			"configuredBatchSize": b.config.BatchSize,
			"calculatedMaxRows":   maxRowsForDB,
			"columnCount":         columnCount,
			"maxParameters":       maxParams,
			"dbType":              b.dbType,
		}).Debug("Reducing batch size due to database parameter limit")
		return maxRowsForDB
	}
	return b.config.BatchSize
}
func (b *BatchProcessor) buildSkipConflictClause() clause.OnConflict {
	return b.plugin.BuildSkipConflictClause(b.config.SkipConflictPKColumns)
}
func (b *BatchProcessor) buildUpsertClause(record map[string]any) clause.OnConflict {
	conflictCols := make([]clause.Column, len(b.config.UpsertPKColumns))
	for i, col := range b.config.UpsertPKColumns {
		conflictCols[i] = clause.Column{Name: col}
	}
	pkSet := make(map[string]bool, len(b.config.UpsertPKColumns))
	for _, pk := range b.config.UpsertPKColumns {
		pkSet[pk] = true
	}
	var updateCols []string
	for col := range record {
		if !pkSet[col] {
			updateCols = append(updateCols, col)
		}
	}
	if len(updateCols) == 0 {
		return clause.OnConflict{
			Columns:   conflictCols,
			DoNothing: true,
		}
	}
	return clause.OnConflict{
		Columns:   conflictCols,
		DoUpdates: clause.AssignmentColumns(updateCols),
	}
}
func (b *BatchProcessor) InsertBatch(db *gorm.DB, schema, tableName string, records []map[string]any) error {
	if len(records) == 0 {
		return nil
	}
	columnCount := 0
	if len(records) > 0 {
		columnCount = len(records[0])
	}
	if columnCount == 0 {
		log.WithFields(map[string]any{
			"table":       tableName,
			"recordCount": len(records),
		}).Error("All records are empty - no columns to insert")
		return fmt.Errorf("cannot insert empty records into %s: all columns were skipped (check for incorrect auto-increment/computed detection)", tableName)
	}
	effectiveBatchSize := b.calculateBatchSize(columnCount)
	fullTableName := b.plugin.FormTableName(schema, tableName)
	if b.config.UseBulkInsert {
		query := db.Table(fullTableName)
		if len(b.config.SkipConflictPKColumns) > 0 {
			query = query.Clauses(b.buildSkipConflictClause())
		} else if len(b.config.UpsertPKColumns) > 0 {
			query = query.Clauses(b.buildUpsertClause(records[0]))
		}
		result := query.CreateInBatches(records, effectiveBatchSize)
		if result.Error != nil {
			log.WithError(result.Error).
				WithField("table", fullTableName).
				WithField("recordCount", len(records)).
				Error("Batch insert failed")
			return result.Error
		}
		if b.config.LogProgress {
			log.WithField("table", fullTableName).
				WithField("recordsInserted", result.RowsAffected).
				Info("Batch insert completed")
		}
		return nil
	}
	totalRecords := len(records)
	for i := 0; i < totalRecords; i += effectiveBatchSize {
		end := i + effectiveBatchSize
		if end > totalRecords {
			end = totalRecords
		}
		batch := records[i:end]
		err := db.Transaction(func(tx *gorm.DB) error {
			for _, record := range batch {
				query := tx.Table(fullTableName)
				if len(b.config.SkipConflictPKColumns) > 0 {
					query = query.Clauses(b.buildSkipConflictClause())
				} else if len(b.config.UpsertPKColumns) > 0 {
					query = query.Clauses(b.buildUpsertClause(record))
				}
				if err := query.Create(record).Error; err != nil {
					if b.config.FailOnError {
						return err
					}
					log.WithError(err).
						WithField("table", fullTableName).
						Warn("Failed to insert record, continuing")
				}
			}
			return nil
		})
		if err != nil && b.config.FailOnError {
			return fmt.Errorf("batch insert failed at batch %d: %w", i/effectiveBatchSize, err)
		}
		if b.config.LogProgress {
			log.WithField("progress", fmt.Sprintf("%d/%d", end, totalRecords)).
				WithField("table", fullTableName).
				Debug("Batch progress")
		}
	}
	return nil
}
func (b *BatchProcessor) ExportInBatches(db *gorm.DB, schema, tableName string, columns []string, writer func([]map[string]any) error) error {
	fullTableName := b.plugin.FormTableName(schema, tableName)
	query := db.Table(fullTableName)
	if len(columns) > 0 {
		query = query.Select(columns)
	}
	offset := 0
	totalExported := 0
	for {
		var batch []map[string]any
		result := query.Limit(b.config.BatchSize).Offset(offset).Find(&batch)
		if result.Error != nil {
			return fmt.Errorf("failed to fetch batch at offset %d: %w", offset, result.Error)
		}
		if len(batch) == 0 {
			break
		}
		if err := writer(batch); err != nil {
			return fmt.Errorf("failed to write batch at offset %d: %w", offset, err)
		}
		totalExported += len(batch)
		offset += b.config.BatchSize
		if b.config.LogProgress {
			log.WithField("recordsExported", totalExported).
				WithField("table", fullTableName).
				Debug("Export progress")
		}
		if len(batch) < b.config.BatchSize {
			break
		}
	}
	if b.config.LogProgress {
		log.WithField("table", fullTableName).
			WithField("totalRecordsExported", totalExported).
			Info("Export completed")
	}
	return nil
}
func (p *GormPlugin) BulkAddRows(config *engine.PluginConfig, schema string, storageUnit string, rows [][]engine.Record) (bool, error) {
	return plugins.WithConnection(config, p.DB, func(db *gorm.DB) (bool, error) {
		var records []map[string]any
		for _, row := range rows {
			record, err := p.ConvertRecordValuesToMap(row)
			if err != nil {
				return false, fmt.Errorf("failed to convert row: %w", err)
			}
			records = append(records, record)
		}
		batchCfg := &BatchConfig{
			BatchSize:     1000,
			UseBulkInsert: true,
			FailOnError:   true,
			LogProgress:   len(records) > 10000,
		}
		if config != nil {
			batchCfg.UpsertPKColumns = config.UpsertPKColumns
			batchCfg.SkipConflictPKColumns = config.SkipConflictPKColumns
		}
		processor := NewBatchProcessor(p.GormPluginFunctions, p.Type, batchCfg)
		err := processor.InsertBatch(db, schema, storageUnit, records)
		return err == nil, err
	})
}