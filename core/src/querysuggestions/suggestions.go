package querysuggestions
import (
	"fmt"
	"github.com/clidey/whodb/core/src/source"
)
const maxSuggestions = 3
type Suggestion struct {
	Description string
	Category    string
}
func FromStorageUnits(units []source.StorageUnit) []Suggestion {
	if len(units) == 0 {
		return []Suggestion{}
	}
	if len(units) > maxSuggestions {
		units = units[:maxSuggestions]
	}
	suggestions := make([]Suggestion, 0, len(units))
	for i, unit := range units {
		tableName := unit.Name
		suggestion := Suggestion{}
		switch i % 3 {
		case 0:
			suggestion = Suggestion{
				Description: fmt.Sprintf("What are the most recent records in %s?", tableName),
				Category:    "SELECT",
			}
		case 1:
			suggestion = Suggestion{
				Description: fmt.Sprintf("How many total entries are in %s?", tableName),
				Category:    "AGGREGATE",
			}
		default:
			suggestion = Suggestion{
				Description: fmt.Sprintf("Show me all the data in %s", tableName),
				Category:    "SELECT",
			}
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions
}