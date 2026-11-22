package querysuggestions
import (
	"strings"
	"testing"
	"github.com/clidey/whodb/core/src/engine"
)
func TestFromStorageUnitsCapsAtThree(t *testing.T) {
	suggestions := FromStorageUnits([]engine.StorageUnit{
		{Name: "users"},
		{Name: "orders"},
		{Name: "payments"},
		{Name: "ignored"},
	})
	if len(suggestions) != 3 {
		t.Fatalf("expected 3 suggestions, got %#v", suggestions)
	}
	if !strings.Contains(suggestions[0].Description, "users") {
		t.Fatalf("expected first suggestion to mention users, got %#v", suggestions[0])
	}
	if suggestions[1].Category != "AGGREGATE" {
		t.Fatalf("expected second suggestion category AGGREGATE, got %#v", suggestions[1])
	}
	for _, suggestion := range suggestions {
		if strings.Contains(suggestion.Description, "ignored") {
			t.Fatalf("did not expect ignored table to appear, got %#v", suggestions)
		}
	}
}
func TestFromStorageUnitsEmpty(t *testing.T) {
	suggestions := FromStorageUnits(nil)
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions, got %#v", suggestions)
	}
}