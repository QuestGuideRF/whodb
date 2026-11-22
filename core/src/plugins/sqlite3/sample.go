package sqlite3
import (
	_ "embed"
	"sync"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"github.com/clidey/whodb/core/src/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)
var sampleSQL string
const SampleDatabaseName = "whodb-sample"
const sampleDatabaseURI = "file:" + SampleDatabaseName + "?mode=memory&cache=shared"
var (
	sampleDBOnce sync.Once
	sampleDBErr  error
)
func IsSampleDatabase(name string) bool {
	return name == SampleDatabaseName
}
func GetSampleProfile() types.DatabaseCredentials {
	return types.DatabaseCredentials{
		Alias:     "Sample SQLite Database",
		Database:  SampleDatabaseName,
		Type:      engine.DatabaseType_Sqlite3,
		IsProfile: true,
		Source:    "builtin",
	}
}
func GetSampleDatabase() (*gorm.DB, error) {
	sampleDBOnce.Do(func() {
		db, err := gorm.Open(sqlite.Open(sampleDatabaseURI), &gorm.Config{
			Logger: logger.Default.LogMode(plugins.GetGormLogConfig()),
		})
		if err != nil {
			log.WithError(err).Error("Failed to create sample in-memory database")
			sampleDBErr = err
			return
		}
		if err := db.Exec(sampleSQL).Error; err != nil {
			log.WithError(err).Error("Failed to initialize sample database schema")
			sampleDBErr = err
			return
		}
	})
	if sampleDBErr != nil {
		return nil, sampleDBErr
	}
	return gorm.Open(sqlite.Open(sampleDatabaseURI), &gorm.Config{
		Logger: logger.Default.LogMode(plugins.GetGormLogConfig()),
	})
}