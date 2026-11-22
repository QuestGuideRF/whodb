package mockdata
import (
	"strings"
	"github.com/clidey/whodb/core/src/env"
)
func IsMockDataGenerationAllowed(tableName string) bool {
	if env.DisableMockDataGeneration == "" {
		return true
	}
	if env.DisableMockDataGeneration == "*" {
		return false
	}
	disabledTables := strings.Split(env.DisableMockDataGeneration, ",")
	for _, disabled := range disabledTables {
		if strings.TrimSpace(disabled) == tableName {
			return false
		}
	}
	return true
}
func GetMockDataGenerationMaxRowCount() int {
	return 200
}