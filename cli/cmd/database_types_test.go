package cmd
import (
	"testing"
)
func TestSourceCatalogHelpersUseSourceTraits(t *testing.T) {
	t.Parallel()
	if !isFileBasedDatabaseType("sqlite") {
		t.Fatalf("expected sqlite synonym to resolve as file-based")
	}
	if !isFileBasedDatabaseType("DuckDB") {
		t.Fatalf("expected DuckDB to resolve as file-based")
	}
	if isFileBasedDatabaseType("Postgres") {
		t.Fatalf("expected Postgres to remain network-based")
	}
	if !isConnectionFieldRequired("Postgres", "Hostname") {
		t.Fatalf("expected Postgres hostname to remain required")
	}
	if isConnectionFieldRequired("Redis", "Database") {
		t.Fatalf("expected Redis database field to remain optional/absent")
	}
}