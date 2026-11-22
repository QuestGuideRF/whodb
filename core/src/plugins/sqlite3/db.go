package sqlite3
import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"github.com/clidey/whodb/core/src/engine"
	"github.com/clidey/whodb/core/src/env"
	"github.com/clidey/whodb/core/src/log"
	"github.com/clidey/whodb/core/src/plugins"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)
func getDefaultDirectory() string {
	directory := "/db/"
	if env.IsDevelopment {
		directory = "tmp/"
	}
	return directory
}
var errDoesNotExist = errors.New("unauthorized or the database doesn't exist")
func (p *Sqlite3Plugin) DB(config *engine.PluginConfig) (*gorm.DB, error) {
	connectionInput, err := p.ParseConnectionConfig(config)
	if err != nil {
		return nil, err
	}
	database := connectionInput.Database
	if IsSampleDatabase(database) {
		return GetSampleDatabase()
	}
	var fileNameDatabase string
	if env.GetIsLocalMode() {
		fileNameDatabase = database
		if _, err := os.Stat(fileNameDatabase); errors.Is(err, os.ErrNotExist) {
			log.WithError(err).WithFields(map[string]any{
				"database": database,
				"path":     fileNameDatabase,
			}).Error("SQLite database file does not exist")
			return nil, errDoesNotExist
		}
	} else {
		fileNameDatabase = filepath.Join(getDefaultDirectory(), database)
		fileNameDatabase, err = filepath.EvalSymlinks(fileNameDatabase)
		if err != nil {
			log.WithError(err).WithFields(map[string]any{
				"database": database,
				"path":     fileNameDatabase,
			}).Error("Failed to evaluate SQLite database symlinks")
			return nil, err
		}
		if !strings.HasPrefix(fileNameDatabase, getDefaultDirectory()) {
			log.WithFields(map[string]any{
				"database":         database,
				"path":             fileNameDatabase,
				"defaultDirectory": getDefaultDirectory(),
			}).Error("SQLite database path is outside allowed directory")
			return nil, errDoesNotExist
		}
		if _, err := os.Stat(fileNameDatabase); errors.Is(err, os.ErrNotExist) {
			log.WithError(err).WithFields(map[string]any{
				"database": database,
				"path":     fileNameDatabase,
			}).Error("SQLite database file does not exist")
			return nil, errDoesNotExist
		}
	}
	l := log.WithFields(map[string]any{
		"database": database,
		"path":     fileNameDatabase,
	})
	db, err := gorm.Open(sqlite.Open(fileNameDatabase), &gorm.Config{Logger: logger.Default.LogMode(plugins.GetGormLogConfig())})
	if err != nil {
		l.WithError(err).Error("Failed to connect to SQLite database")
		return nil, err
	}
	if err := plugins.ConfigureConnectionPool(db); err != nil {
		l.WithError(err).Warn("Failed to configure connection pool")
	}
	return db, nil
}